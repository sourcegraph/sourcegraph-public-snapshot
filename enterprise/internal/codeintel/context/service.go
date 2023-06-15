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
	"github.com/sourcegraph/sourcegraph/internal/codeintel/types"
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

func (s *Service) GetPreciseContext(ctx context.Context, args *resolverstubs.GetPreciseContextInput) ([]*types.PreciseData, error) {
	// TODO: validate args whether here or at the resolver level
	filename := args.Input.ActiveFile
	content := args.Input.ActiveFileContent
	commitID := args.Input.CommitID

	repoID, err := strconv.Atoi(args.Input.Repository)
	if err != nil {
		return nil, err
	}

	uploads, err := s.codenavSvc.GetClosestDumpsForBlob(ctx, repoID, commitID, filename, true, "")
	if err != nil {
		return nil, err
	}

	syntectDocument, err := s.getSCIPDocumentByContent(ctx, content, filename)
	if err != nil {
		return nil, err
	}

	symbolNames := make([]string, 0, len(syntectDocument.Occurrences))
	for _, occurrence := range syntectDocument.Occurrences {
		symbolNames = append(symbolNames, occurrence.Symbol)
	}

	// TODO: Either pass a slice of uploads or loop thru uploads and pass the ids to fix this hardcoding of the first
	scipNames, err := s.codenavSvc.GetFullSCIPNameByDescriptor(ctx, []int{uploads[0].ID}, symbolNames)
	if err != nil {
		return nil, err
	}

	repo, err := s.repostore.GetByIDs(ctx, api.RepoID(repoID))
	if err != nil {
		return nil, err
	}
	if len(repo) == 0 {
		return nil, fmt.Errorf("repo not found")
	}

	type preciseData struct {
		symbolName string
		// syntectDescriptor string
		repository string
		symbolRole int32
		confidence string
		// text              string
		location []codenavshared.UploadLocation
	}

	preciseDataList := []*preciseData{}
	definitionMap := map[string]*preciseData{}
	for _, name := range scipNames {
		args := codenavtypes.RequestArgs{
			RepositoryID: repoID,
			Commit:       commitID,
			Path:         "",
			Line:         0,
			Character:    0,
			Limit:        100, //! MAGIC NUMBER
			RawCursor:    "",
		}
		hunkCache, err := codenav.NewHunkCache(hunkCacheSize)
		if err != nil {
			return nil, err
		}
		reqState := codenavtypes.NewRequestState(
			uploads,
			s.repostore,
			authz.DefaultSubRepoPermsChecker,
			s.gitserverClient,
			repo[0],
			commitID,
			"",
			maximumIndexesPerMonikerSearch,
			hunkCache,
		)

		// if _, ok := seenOccurrences[name.GetIdentifier()]; ok {
		// 	continue
		// }
		// seenOccurrences[name.GetIdentifier()] = struct{}{}

		for _, upload := range uploads {
			loc, err := s.codenavSvc.GetLocationByExplodedSymbol(ctx, name.GetIdentifier(), upload.ID, "definition_ranges")
			if err != nil {
				return nil, err
			}

			ul, err := s.codenavSvc.GetUploadLocations(ctx, args, reqState, loc, true)
			if err != nil {
				return nil, err
			}

			pd := &preciseData{
				symbolName: name.GetIdentifier(),
				repository: string(repo[0].Name),
				// symbolRole: int32(el.Occurrence.SymbolRoles),
				confidence: "PRECISE",
				location:   ul,
			}

			e := strings.Split(name.GetIdentifier(), "/")
			key := e[len(e)-1]
			definitionMap[key] = pd

			preciseDataList = append(preciseDataList, pd)
		}
	}

	snippetToPreciseDataMap := map[string]*preciseData{}
	clippedContent := map[string]struct{}{}
	// var syntectDocsList []*scip.Document
	// for _, pd := range definitionMap {
	for _, pd := range preciseDataList {
		for _, l := range pd.location {
			file, err := s.gitserverClient.ReadFile(
				ctx,
				authz.DefaultSubRepoPermsChecker,
				api.RepoName(l.Dump.RepositoryName),
				api.CommitID(l.Dump.Commit),
				l.Path,
			)
			if err != nil {
				return nil, err
			}
			c := strings.Split(string(file), "\n")

			syntectDocs, err := s.getSCIPDocumentByContent(ctx, string(file), l.Path)
			if err != nil {
				return nil, err
			}

			for _, occ := range syntectDocs.Occurrences {
				r := scip.NewRange(occ.EnclosingRange)
				snpt := extractSnippet(c, r.Start.Line, r.End.Line, r.Start.Character, r.End.Character)
				clippedContent[snpt] = struct{}{}

				e := strings.Split(occ.Symbol, "/")
				key := e[len(e)-1]

				keyLookup := fmt.Sprintf("%s$$$$%s", key, snpt)

				data := &preciseData{
					confidence: "SEARCH",
				}
				if _, ok := definitionMap[key]; ok {
					data = definitionMap[key]
				}
				snippetToPreciseDataMap[keyLookup] = data
			}
		}
	}

	preciseResponse := []*types.PreciseData{}
	for k, v := range snippetToPreciseDataMap {
		compositeKey := strings.Split(k, "$$$$")
		syntectDescriptor, text := compositeKey[0], compositeKey[1]
		preciseResponse = append(preciseResponse, &types.PreciseData{
			SymbolName:        v.symbolName,
			SyntectDescriptor: syntectDescriptor,
			Repository:        v.repository,
			SymbolRole:        v.symbolRole,
			Confidence:        v.confidence,
			Text:              text,
		})
	}

	return preciseResponse, nil
}

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
