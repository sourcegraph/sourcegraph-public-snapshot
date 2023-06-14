package context

import (
	"context"
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"

	"google.golang.org/protobuf/proto"

	//"github.com/sourcegraph/scip/bindings/go/scip"

	"github.com/sourcegraph/scip/bindings/go/scip"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/codenav"
	codenavtypes "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/codenav"
	codenavshared "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/codenav/shared"
	scipstore "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/context/internal/scipstore"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	resolverstubs "github.com/sourcegraph/sourcegraph/internal/codeintel/resolvers"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/gosyntect"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type Service struct {
	codenavSvc      CodeNavService
	scipstore       scipstore.ScipStore
	repostore       database.RepoStore
	syntectClient   *gosyntect.Client
	gitserverClient gitserver.Client
	operations      *operations
}

func newService(
	observationCtx *observation.Context,
	scipstore scipstore.ScipStore,
	repostore database.RepoStore,
	codenavSvc CodeNavService,
	syntectClient *gosyntect.Client,
	gitserverClient gitserver.Client,
) *Service {
	return &Service{
		codenavSvc:      codenavSvc,
		scipstore:       scipstore,
		repostore:       repostore,
		syntectClient:   syntectClient,
		gitserverClient: gitserverClient,
		operations:      newOperations(observationCtx),
	}
}

// TODO move this to a config file
// Flagrantly taken from default value in enterprise/cmd/frontend/internal/codeintel/config.go
const (
	maximumIndexesPerMonikerSearch = 500
	hunkCacheSize                  = 1000
)

func (s *Service) FindMostRelevantSCIPSymbols(ctx context.Context, args *resolverstubs.FindMostRelevantSCIPSymbolsArgs) (string, error) {
	filename := args.Args.EditorState.ActiveFile
	content := args.Args.EditorState.ActiveFileContent
	commitID := args.Args.CommitID

	repoID, err := strconv.Atoi(args.Args.Repository)
	if err != nil {
		return "", err
	}

	// TODO: Get uploads from codenav service
	uploads, err := s.codenavSvc.GetClosestDumpsForBlob(ctx, repoID, commitID, filename, true, "")
	if err != nil {
		return "", err
	}

	syntectDocument, err := s.getSCIPDocumentByContent(ctx, content, filename)
	if err != nil {
		return "", err
	}

	usersDescriptorsMap := map[string]struct{}{}
	symbolNames := make([]string, 0, len(syntectDocument.Occurrences))
	for _, occurrence := range syntectDocument.Occurrences {
		symbolNames = append(symbolNames, occurrence.Symbol)
		e := strings.Split(occurrence.Symbol, "/")
		key := e[len(e)-1]
		if key != "" {
			usersDescriptorsMap[key] = struct{}{}
		}
	}

	var scipDocs []*scip.Document
	for _, upload := range uploads {
		sd, err := s.codenavSvc.GetSCIPDocumentsBySymbolNames(ctx, upload.ID, symbolNames)
		if err != nil {
			return "", err
		}
		scipDocs = append(scipDocs, sd...)
	}

	type filteredIndex struct {
		Document   *scip.Document
		Occurrence *scip.Occurrence
	}

	usersDescriptorToIndexMap := map[string][]filteredIndex{}
	for _, sdoc := range scipDocs {
		for _, occ := range sdoc.Occurrences {
			e := strings.Split(occ.Symbol, "/")
			key := e[len(e)-1]

			if _, ok := usersDescriptorsMap[key]; ok {
				usersDescriptorToIndexMap[key] = append(usersDescriptorToIndexMap[key], filteredIndex{
					Document:   sdoc,
					Occurrence: occ,
				})
			}
		}
	}

	repo, err := s.repostore.GetByIDs(ctx, api.RepoID(repoID))
	if err != nil {
		return "", err
	}
	if len(repo) == 0 {
		return "", fmt.Errorf("repo not found")
	}

	// {

	//     symbol: "New()."
	//     repository: "github.com/sourcegraph/sourcegraph",
	//     type: DEFINITION,
	//     text: "func New() *Hello {\n\tm := world.New()\n\treturn &Hello{World: m}\n}",
	// },

	type preciseData struct {
		symbolName        string
		syntectDescriptor string
		repository        string
		symbolRole        int32
		confidence        string
		text              string
		location          []codenavshared.UploadLocation
	}

	definitionMap := map[string]*preciseData{}

	seenOccurrences := map[string]struct{}{}
	var defsList []codenavshared.UploadLocation
	var preciseDataList []*preciseData
	for _, oap := range usersDescriptorToIndexMap {
		for _, el := range oap {
			r := scip.NewRange(el.Occurrence.Range)
			args := codenavtypes.RequestArgs{
				RepositoryID: repoID,
				Commit:       commitID,
				Path:         el.Document.RelativePath,
				Line:         int(r.Start.Line),
				Character:    int(r.Start.Character),
				Limit:        100, //! MAGIC NUMBER
				RawCursor:    "",
			}
			hunkCache, err := codenav.NewHunkCache(hunkCacheSize)
			if err != nil {
				return "", err
			}
			reqState := codenavtypes.NewRequestState(
				uploads,
				s.repostore,
				authz.DefaultSubRepoPermsChecker,
				s.gitserverClient,
				repo[0],
				commitID,
				el.Document.RelativePath,
				maximumIndexesPerMonikerSearch,
				hunkCache,
			)

			if _, ok := seenOccurrences[el.Occurrence.Symbol]; ok {
				continue
			}
			seenOccurrences[el.Occurrence.Symbol] = struct{}{}

			// Shit still works
			// test, err := s.codenavSvc.GetDefinitions(ctx, args, reqState)
			// if err != nil {
			// 	return "", err
			// }
			// defsList = append(defsList, test...)
			for _, upload := range uploads {
				loc, _, err := s.codenavSvc.GetScipDefinitionsLocation(ctx, el.Document, el.Occurrence, upload.ID, el.Document.RelativePath, 100, 0)
				if err != nil {
					return "", err
				}

				ul, err := s.codenavSvc.GetUploadLocations(ctx, args, reqState, loc, true)
				if err != nil {
					return "", err
				}
				fmt.Println("HERE IS THE ul \n", ul)
				defsList = append(defsList, ul...)
				pd := &preciseData{
					symbolName: el.Occurrence.Symbol,
					repository: string(repo[0].Name),
					symbolRole: int32(el.Occurrence.SymbolRoles),
					confidence: "PRECISE",
					location:   ul,
				}
				preciseDataList = append(preciseDataList, pd)
				e := strings.Split(el.Occurrence.Symbol, "/")
				key := e[len(e)-1]
				definitionMap[key] = pd
				// if key != "" {

				// }

				fmt.Println("HERE IS THE defsList \n", defsList[0])
			}
		}
	}

	keysInString := []string{}
	textInString := []string{}
	snippetToPreciseDataMap := map[string]*preciseData{}
	clippedContent := map[string]struct{}{}
	// var syntectDocsList []*scip.Document
	for _, pd := range definitionMap {
		// for _, pd := range preciseDataList {
		for _, l := range pd.location {
			file, err := s.gitserverClient.ReadFile(
				ctx,
				authz.DefaultSubRepoPermsChecker,
				api.RepoName(l.Dump.RepositoryName),
				api.CommitID(l.Dump.Commit),
				l.Path,
			)
			if err != nil {
				return "", err
			}
			c := strings.Split(string(file), "\n")
			// fmt.Println("HERE IS THE c", c)

			syntectDocs, err := s.getSCIPDocumentByContent(ctx, string(file), l.Path)
			if err != nil {
				return "", err
			}
			fmt.Println("HERE IS THE syntectDocs \n", syntectDocs)

			for _, occ := range syntectDocs.Occurrences {

				r := scip.NewRange(occ.EnclosingRange)
				fmt.Println("HERE IS THE c \n", c)
				fmt.Println("HERE IS THE r \n", r)
				fmt.Println("HERE IS THE occ \n", occ)
				fmt.Println("HERE IS THE clippedContent \n", clippedContent)
				snpt := extractSnippet(c, r.Start.Line, r.End.Line, r.Start.Character, r.End.Character)
				clippedContent[snpt] = struct{}{}

				e := strings.Split(occ.Symbol, "/")
				key := e[len(e)-1]
				keysInString = append(keysInString, key)
				textInString = append(textInString, snpt)

				keyLookup := fmt.Sprintf("%s$$$$%s", key, snpt)
				// pd.text = snpt

				// HERE check the World New()
				data := &preciseData{
					confidence: "SEARCH",
				}
				if _, ok := definitionMap[key]; ok {
					data = definitionMap[key]
				}
				snippetToPreciseDataMap[keyLookup] = data

				// if key == "" || snpt == "" {
				// 	continue
				// }
				// if _, ok := definitionMap[key]; ok {
				// 	definitionMap[key].text = snpt
				// }
				// prd := scip.NewRange(occ.Range)
				// pprd := precise.RangeData{
				// 	StartLine:      int(prd.Start.Line),
				// 	EndLine:        int(prd.End.Line),
				// 	StartCharacter: int(prd.Start.Character),
				// 	EndCharacter:   int(prd.End.Character),
				// }
				// r := scip.NewRange(occ.EnclosingRange)
				// _ = occ
				// fmt.Println("HERE IS THE prd \n", prd)
				// fmt.Println("HERE IS THE c \n", c)
				// fmt.Println("HERE IS THE r \n", r)
				// fmt.Println("HERE IS THE occ \n", occ)
				// fmt.Println("HERE IS THE clippedContent \n", clippedContent)
				// isInside := precise.RangeIntersectsSpan(pprd, int(r.Start.Line), int(r.End.Line))
				// fmt.Println("HERE IS THE isInside \n", isInside)
				// if isInside {
				// 	snpt := extractSnippet(c, r.Start.Line, r.End.Line, r.Start.Character, r.End.Character)
				// 	clippedContent[snpt] = struct{}{}

				// 	e := strings.Split(occ.Symbol, "/")
				// 	key := e[len(e)-1]
				// 	if key == "" || snpt == "" {
				// 		continue
				// 	}
				// 	if _, ok := definitionMap[key]; ok {
				// 		definitionMap[key].text = snpt
				// 	}
				// }
			}
		}
	}

	printOne := textInString
	printTwo := keysInString
	printThree := snippetToPreciseDataMap
	fmt.Println("HERE IS THE definitionMap \n", definitionMap)
	fmt.Println("HERE IS THE clippedContent \n", clippedContent)
	fmt.Println("HERE IS THE snippetToPreciseDataMap \n", snippetToPreciseDataMap)

	fmt.Println("HERE IS THE textInString \n", printOne)
	fmt.Println("HERE IS THE keysInString \n", printTwo)
	fmt.Println("HERE IS THE preciseDataList \n", printThree)

	preciseResponse := []*preciseData{}
	for k, v := range snippetToPreciseDataMap {
		compositeKey := strings.Split(k, "$$$$")
		syntectDescriptor, text := compositeKey[0], compositeKey[1]
		preciseResponse = append(preciseResponse, &preciseData{
			symbolName:        v.symbolName,
			syntectDescriptor: syntectDescriptor,
			repository:        v.repository,
			symbolRole:        v.symbolRole,
			confidence:        v.confidence,
			text:              text,
			location:          v.location,
		})
	}

	printFour := preciseResponse
	fmt.Println("HERE IS THE preciseResponse \n", printFour)

	for _, v := range preciseResponse {
		fmt.Println("HERE IS THE preciseResponse \n", v)
	}

	var allContent string
	for k := range clippedContent {
		// allContent = append(allContent, k)
		allContent += "\n" + k
	}

	return allContent, nil
}

// WORKS GREAT!
// func (s *Service) FindMostRelevantSCIPSymbols(ctx context.Context, args *resolverstubs.FindMostRelevantSCIPSymbolsArgs) (string, error) {
// 	filename := args.Args.EditorState.ActiveFile
// 	content := args.Args.EditorState.ActiveFileContent
// 	commitID := args.Args.CommitID
// 	// repoName := "github.com/Numbers88s/simple-mock-go"

// 	//! TODO: Change repository name to repository ID
// 	repoID, err := strconv.Atoi(args.Args.Repository)
// 	if err != nil {
// 		return "", err
// 	}

// 	// TODO: Get uploads from codenav service
// 	uploads, err := s.codenavSvc.GetClosestDumpsForBlob(ctx, repoID, commitID, filename, true, "")
// 	if err != nil {
// 		return "", err
// 	}
// 	// fmt.Println("HERE is the upload", uploads)

// 	// Step #1: Get scip syntectDocument created by treesitter
// 	syntectDocument, err := s.getSCIPDocumentByContent(ctx, content, filename)
// 	if err != nil {
// 		return "", err
// 	}

// 	// fmt.Println("HERE is the document", document)

// 	usersDescriptorsMap := map[string]struct{}{}
// 	symbolNames := make([]string, 0, len(syntectDocument.Occurrences))
// 	for _, occurrence := range syntectDocument.Occurrences {
// 		symbolNames = append(symbolNames, occurrence.Symbol)
// 		e := strings.Split(occurrence.Symbol, "/")
// 		key := e[len(e)-1]
// 		if key != "" {
// 			usersDescriptorsMap[key] = struct{}{}
// 		}
// 	}

// 	var (
// 		scipDocs  []*scip.Document
// 		uploadIDs []int
// 	)
// 	for _, upload := range uploads {
// 		// TODO: pass in a list of uploadIDs instead of looping thru
// 		// rangeMap, api.RepoName(repoName), api.RepoID(repoID), api.CommitID(commitID), filename
// 		sd, err := s.codenavSvc.GetSCIPDocumentsBySymbolNames(ctx, upload.ID, symbolNames)
// 		if err != nil {
// 			return "", err
// 		}
// 		// fmt.Println("HERE is the sd", sd)
// 		scipDocs = append(scipDocs, sd...)
// 		uploadIDs = append(uploadIDs, upload.ID)
// 	}

// 	// fmt.Println("HERE is the scipDocs", scipDocs)

// 	type occurrenceAndPath struct {
// 		Occurrence   *scip.Occurrence
// 		RelativePath string
// 	}
// 	usersDescriptorToOccurrencesMap := map[string][]occurrenceAndPath{}
// 	for _, sdoc := range scipDocs {
// 		for _, occ := range sdoc.Occurrences {
// 			e := strings.Split(occ.Symbol, "/")
// 			// fmt.Println("HERE IS THE e", e)
// 			key := e[len(e)-1]
// 			// fmt.Println("HERE IS THE key", key)

// 			if _, ok := usersDescriptorsMap[key]; ok {
// 				usersDescriptorToOccurrencesMap[key] = append(usersDescriptorToOccurrencesMap[key], occurrenceAndPath{
// 					Occurrence:   occ,
// 					RelativePath: sdoc.RelativePath,
// 				})
// 			}
// 		}
// 	}
// 	// fmt.Println("HERE IS THE usersDescriptorToOccurrencesMap", usersDescriptorToOccurrencesMap)

// 	repo, err := s.repostore.GetByIDs(ctx, api.RepoID(repoID))
// 	if err != nil {
// 		return "", err
// 	}
// 	if len(repo) == 0 {
// 		return "", fmt.Errorf("repo not found")
// 	}

// 	var defsList []codenavshared.UploadLocation
// 	for _, oap := range usersDescriptorToOccurrencesMap {
// 		for _, occ := range oap {
// 			r := scip.NewRange(occ.Occurrence.Range)
// 			args := codenavtypes.RequestArgs{
// 				RepositoryID: repoID,
// 				Commit:       commitID,
// 				Path:         occ.RelativePath,
// 				Line:         int(r.Start.Line),
// 				Character:    int(r.Start.Character),
// 				Limit:        100, //! MAGIC NUMBER
// 				RawCursor:    "",
// 			}
// 			hunkCache, err := codenav.NewHunkCache(hunkCacheSize)
// 			if err != nil {
// 				return "", err
// 			}
// 			reqState := codenavtypes.NewRequestState(
// 				uploads,
// 				s.repostore,
// 				authz.DefaultSubRepoPermsChecker,
// 				s.gitserverClient,
// 				repo[0],
// 				commitID,
// 				occ.RelativePath,
// 				maximumIndexesPerMonikerSearch,
// 				hunkCache,
// 			)

// 			/// Option #1: Get definitions
// defs, err := s.codenavSvc.GetDefinitions(ctx, args, reqState)
// if err != nil {
// 	return "", err
// }

// 			/// Option #2: Get definitions
// 			// nms, err := s.codenavSvc.GetFullSCIPNameByDescriptor(ctx, uploadIDs, symbolNames)
// 			// if err != nil {
// 			// 	return "", err
// 			// }

// 			// // fmt.Println("HERE IS THE nms", &nms)

// 			// var pqm []precise.QualifiedMonikerData
// 			// for _, nm := range nms {
// 			// 	// fmt.Println("HERE IS THE nm", nm)
// 			// 	pqm = append(pqm, precise.QualifiedMonikerData{
// 			// 		MonikerData: precise.MonikerData{
// 			// 			Kind:       "import",
// 			// 			Scheme:     nm.Scheme,
// 			// 			Identifier: nm.GetIdentifier(),
// 			// 		},
// 			// 		PackageInformationData: precise.PackageInformationData{
// 			// 			Manager: nm.PackageManager,
// 			// 			Name:    nm.PackageName,
// 			// 			Version: nm.PackageVersion,
// 			// 		},
// 			// 	})
// 			// }

// 			// // fmt.Println("PASSSSSES", pqm)

// 			// defs, err := s.codenavSvc.GetDefinitionBySymbolName(ctx, pqm, reqState, args)
// 			// if err != nil {
// 			// 	return "", err
// 			// }

// 			defsList = append(defsList, defs...)
// 		}
// 	}

// 	clippedContent := map[string]struct{}{}
// 	// var syntectDocsList []*scip.Document
// 	for _, def := range defsList {
// 		file, err := s.gitserverClient.ReadFile(
// 			ctx,
// 			authz.DefaultSubRepoPermsChecker,
// 			api.RepoName(def.Dump.RepositoryName),
// 			api.CommitID(def.Dump.Commit),
// 			def.Path,
// 		)
// 		if err != nil {
// 			return "", err
// 		}
// 		c := strings.Split(string(file), "\n")
// 		// fmt.Println("HERE IS THE c", c)

// 		syntectDocs, err := s.getSCIPDocumentByContent(ctx, string(file), def.Path)
// 		if err != nil {
// 			return "", err
// 		}

// 		prd := precise.RangeData{
// 			StartLine:      def.TargetRange.Start.Line,
// 			EndLine:        def.TargetRange.End.Line,
// 			StartCharacter: def.TargetRange.Start.Character,
// 			EndCharacter:   def.TargetRange.End.Character,
// 		}

// 		for _, occ := range syntectDocs.Occurrences {
// 			r := scip.NewRange(occ.EnclosingRange)
// 			if precise.RangeIntersectsSpan(prd, int(r.Start.Line), int(r.End.Line)) {
// 				// z := strings.Join(c[r.Start.Line:r.End.Line], "\n")
// 				// zz := z[r.Start.Character:r.End.Character]
// 				// fmt.Println("HERE IS THE zz \n", zz)
// 				// fmt.Println("HERE IS THE r.Start.Line", r.Start.Line)
// 				// fmt.Println("HERE IS THE r.End.Line", r.End.Line)
// 				// fmt.Println("HERE IS THE r.Start.Character", r.Start.Character)
// 				// fmt.Println("HERE IS THE r.End.Character", r.End.Character)
// 				snpt := extractSnippet(c, r.Start.Line, r.End.Line, r.Start.Character, r.End.Character)
// 				// fmt.Println("HERE IS THE SNIPPETS", snpt)
// 				// cnt := pageContent(c, &r.Start.Line, &r.End.Line)
// 				// clippedContent = append(clippedContent, cnt)
// 				clippedContent[snpt] = struct{}{}
// 			}
// 		}

// 	}

// 	var allContent string
// 	for k := range clippedContent {
// 		// allContent = append(allContent, k)
// 		allContent += "\n" + k
// 	}

// 	return allContent, nil
// }

func extractSnippet(lines []string, startLine, endLine, startChar, endChar int32) string {
	if startLine > endLine || startLine < 0 || endLine >= int32(len(lines)) {
		return ""
	}

	result := make([]string, 0)

	for i := startLine; i <= endLine; i++ {
		line := lines[i]
		if startChar == 0 && endChar == 0 {
			result = append(result, line)
			continue
		}

		if i == startLine {
			if startChar < 0 || startChar >= int32(len(line)) {
				return ""
			}

			if i == endLine {
				if endChar < startChar || endChar > int32(len(line)) {
					return ""
				}
				result = append(result, line[startChar:endChar])
			} else {
				result = append(result, line[startChar:])
			}
		} else if i == endLine {
			if endChar < 0 || endChar > int32(len(line)) {
				return ""
			}
			result = append(result, line[:endChar])
		} else {
			result = append(result, line)
		}
	}

	return strings.Join(result, "\n")
}

func (s *Service) getSCIPDocumentByContent(ctx context.Context, content, fileName string) (*scip.Document, error) {
	q := gosyntect.SymbolsQuery{
		FileName: fileName,
		Content:  content,
	}

	resp, err := s.syntectClient.Symbols(ctx, &q)
	if err != nil {
		return nil, err
	}

	d, err := base64.StdEncoding.DecodeString(resp.Scip)
	if err != nil {
		fmt.Println("ERROR: ", err)
		return nil, err
	}

	var document scip.Document
	if err := proto.Unmarshal(d, &document); err != nil {
		fmt.Println("ERROR: ", err)
		return nil, err
	}

	return &document, nil
}

func (s *Service) SplitIntoEmbeddableChunks(ctx context.Context, text string, fileName string, splitOptions SplitOptions) ([]EmbeddableChunk, error) {
	return SplitIntoEmbeddableChunks(text, fileName, splitOptions), nil
}

// func pageContent(content []string, startLine, endLine *int32, startChar, endChar int32) string {
// 	if len(content) == 0 {
// 		return ""
// 	}

// 	// fmt.Println("CONTENT >>>>> :", content)
// 	totalContentLength := len(content)
// 	startCursor := 0
// 	endCursor := totalContentLength

// 	// Any nil or illegal value for startLine or endLine gets set to either the start or
// 	// end of the file respectively.

// 	// If startLine is set and is a legit value, set the cursor to point to it.
// 	if startLine != nil && *startLine > 0 {
// 		// The left index is inclusive, so we have to shift it back by 1
// 		startCursor = int(*startLine) - 1
// 	}
// 	if startCursor >= totalContentLength {
// 		startCursor = totalContentLength
// 	}

// 	// If endLine is set and is a legit value, set the cursor to point to it.
// 	if endLine != nil && *endLine >= 0 {
// 		endCursor = int(*endLine)
// 	}
// 	if endCursor > totalContentLength {
// 		endCursor = totalContentLength
// 	}

// 	// Final failsafe in case someone is really messing around with this API.
// 	if endCursor < startCursor {
// 		return strings.Join(content[0:totalContentLength], "\n")
// 	}

// 	c := content[startCursor:endCursor]

// 	return strings.Join(c[startChar:endChar-1], "\n")
// }

// func extactDefinitions(document *scip.Document, occurrence *scip.Occurrence) []*scip.Occurrence {
// 	// hoverText               []string
// 	definitionSymbol := occurrence.Symbol // referencesBySymbol      = map[string]struct{}{}
// 	// implementationsBySymbol = map[string]struct{}{}
// 	// prototypeBySymbol       = map[string]struct{}{}

// 	// Extract hover text and relationship data from the symbol information that
// 	// matches the given occurrence. This will give us additional symbol names that
// 	// we should include in reference and implementation searches.

// 	if symbol := scip.FindSymbol(document, occurrence.Symbol); symbol != nil {
// 		// hoverText = symbol.Documentation
// 		for _, rel := range symbol.Relationships {
// 			if rel.IsDefinition {
// 				definitionSymbol = rel.Symbol
// 			}
// 		}
// 	}

// 	// for _, sym := range document.Symbols {
// 	// 	for _, rel := range sym.Relationships {
// 	// 		if rel.IsImplementation {
// 	// 			if rel.Symbol == occurrence.Symbol {
// 	// 				implementationsBySymbol[occurrence.Symbol] = struct{}{}
// 	// 			}
// 	// 		}
// 	// 	}
// 	// }

// 	definitions := []*scip.Occurrence{}

// 	// Include original symbol names for reference search below
// 	// referencesBySymbol[occurrence.Symbol] = struct{}{}

// 	// For each occurrence that references one of the definition, reference, or
// 	// implementation symbol names, extract and aggregate their source positions.

// 	for _, occ := range document.Occurrences {
// 		isDefinition := scip.SymbolRole_Definition.Matches(occ)

// 		// This occurrence defines this symbol
// 		if definitionSymbol == occ.Symbol && isDefinition {
// 			definitions = append(definitions, occ)
// 		}
// 	}

// 	// Override symbol documentation with occurrence documentation, if it exists
// 	// if len(occurrence.OverrideDocumentation) != 0 {
// 	// 	hoverText = occurrence.OverrideDocumentation
// 	// }

// 	return definitions
// }

// func pageContent(content []string, startLine, endLine *int32) string {
// 	totalContentLength := len(content)
// 	startCursor := 0
// 	endCursor := totalContentLength

// 	// Any nil or illegal value for startLine or endLine gets set to either the start or
// 	// end of the file respectively.

// 	// If startLine is set and is a legit value, set the cursor to point to it.
// 	if startLine != nil && *startLine > 0 {
// 		// The left index is inclusive, so we have to shift it back by 1
// 		startCursor = int(*startLine) - 1
// 	}
// 	if startCursor >= totalContentLength {
// 		startCursor = totalContentLength
// 	}

// 	// If endLine is set and is a legit value, set the cursor to point to it.
// 	if endLine != nil && *endLine >= 0 {
// 		endCursor = int(*endLine)
// 	}
// 	if endCursor > totalContentLength {
// 		endCursor = totalContentLength
// 	}

// 	// Final failsafe in case someone is really messing around with this API.
// 	if endCursor < startCursor {
// 		return strings.Join(content[0:totalContentLength+1], "\n")
// 	}

// 	return strings.Join(content[startCursor:endCursor+1], "\n")
// }
