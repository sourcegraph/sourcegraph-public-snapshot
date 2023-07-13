package context

import (
	"context"
	"encoding/base64"
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/sourcegraph/scip/bindings/go/scip"
	"google.golang.org/protobuf/proto"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/codenav"
	codenavtypes "github.com/sourcegraph/sourcegraph/internal/codeintel/codenav"
	codenavshared "github.com/sourcegraph/sourcegraph/internal/codeintel/codenav/shared"
	resolverstubs "github.com/sourcegraph/sourcegraph/internal/codeintel/resolvers"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/shared/symbols"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/types"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/gosyntect"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type Service struct {
	codenavSvc      CodeNavService
	repostore       database.RepoStore
	syntectClient   *gosyntect.Client
	gitserverClient gitserver.Client
	operations      *operations
}

func newService(
	observationCtx *observation.Context,
	repostore database.RepoStore,
	codenavSvc CodeNavService,
	syntectClient *gosyntect.Client,
	gitserverClient gitserver.Client,
) *Service {
	return &Service{
		codenavSvc:      codenavSvc,
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

func (s *Service) GetPreciseContext(ctx context.Context, args *resolverstubs.GetPreciseContextInput) ([]*types.PreciseContext, error) {
	filename := args.Input.ActiveFile
	content := args.Input.ActiveFileContent
	commitID := args.Input.CommitID

	repo, err := s.repostore.GetByName(ctx, api.RepoName(args.Input.Repository))
	if err != nil {
		return nil, err
	}

	uploads, err := s.codenavSvc.GetClosestDumpsForBlob(ctx, int(repo.ID), commitID, filename, true, "")
	if err != nil {
		return nil, err
	}
	if len(uploads) == 0 {
		return nil, nil
	}

	requestArgs := codenavtypes.RequestArgs{
		RepositoryID: int(repo.ID),
		Commit:       commitID,
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
		repo,
		commitID,
		"",
		maximumIndexesPerMonikerSearch,
		hunkCache,
	)

	fmt.Printf("PHASE 1\n")
	// PHASE 1: Run current scope through treesitter

	syntectDocument, err := s.getSCIPDocumentByContent(ctx, content, filename)
	if err != nil {
		return nil, err
	}

	symbolNameMap := map[string]struct{}{}
	for _, occurrence := range syntectDocument.Occurrences {
		symbolNameMap[occurrence.Symbol] = struct{}{}
	}
	symbolNames := make([]string, 0, len(symbolNameMap))
	for symbolName := range symbolNameMap {
		symbolNames = append(symbolNames, symbolName)
	}
	sort.Strings(symbolNames)

	fmt.Printf("PHASE 2\n")
	// PHASE 2: Run treesitter output through a translation layer so we can do
	// the graph navigation in "SCIP-world" using proper identifiers. The following
	// code is pretty sloppy right now since we haven't consolidated on a single way
	// to "match" descriptors together. This should align in the db layer as well.
	//
	// This isn't a deep technical problem, just one of deciding on a thing and
	// conforming to/communicating it in the codebase.

	// Construct a map from syntect name to a list of SCIP names matching the syntect
	// output. I'd like to have the `GetFullSCIPNameByDescriptor` method create this
	// mapping instead. This block should become a single function call after that
	// transformation.

	scipNamesBySyntectName, err := func() (map[string][]*symbols.ExplodedSymbol, error) {
		// TODO: Either pass a slice of uploads or loop thru uploads and pass the ids to fix this hardcoding of the first
		scipNames, err := s.codenavSvc.GetFullSCIPNameByDescriptor(ctx, []int{uploads[0].ID}, symbolNames)
		if err != nil {
			return nil, err
		}

		strip := func(s string) string {
			parts := strings.Split(s, "/")
			return parts[len(parts)-1]
		}
		scipNamesBySyntectName := map[string][]*symbols.ExplodedSymbol{}
		for _, syntectName := range symbolNames {
			ex, _ := symbols.NewExplodedSymbol(syntectName)
			var symbolNames []*symbols.ExplodedSymbol
			for _, scipName := range scipNames {
				// We do a `descriptor ILIKE %syntectName%` in Postgres today, so this
				// is a bit of a less lenient (we do suffix here instead of contains).
				if strippedDescriptor := strip(ex.Descriptor); strippedDescriptor != "" && strip(scipName.Descriptor) == strippedDescriptor {
					symbolNames = append(symbolNames, scipName)
				}
			}

			if len(symbolNames) > 20 {
				fmt.Printf("TOO MANY RESULTS FOR %q\n", syntectName)
				symbolNames = nil
			}

			if len(symbolNames) > 0 {
				scipNamesBySyntectName[syntectName] = symbolNames
			}
		}

		return scipNamesBySyntectName, nil
	}()
	if err != nil {
		return nil, err
	}

	syntectNameSetBySymbol := map[string]map[string]struct{}{}
	for syntectName, explodedSymbols := range scipNamesBySyntectName {
		for _, explodedSymbol := range explodedSymbols {
			symbol := explodedSymbol.Symbol()
			if _, ok := syntectNameSetBySymbol[symbol]; !ok {
				syntectNameSetBySymbol[symbol] = map[string]struct{}{}
			}

			syntectNameSetBySymbol[symbol][syntectName] = struct{}{}
		}
	}

	fmt.Printf("PHASE 3\n")
	// PHASE 3: Gather definitions for each relevant SCIP symbol

	type preciseData struct {
		syntectName string
		symbolName  string
		location    []codenavshared.UploadLocation
	}
	preciseDataList := []*preciseData{}

	for ident, syntectNames := range syntectNameSetBySymbol {
		// TODO - these are duplicated and should also be batched
		fmt.Printf("> Fetching definitions of %q\n", ident)

		// for _, upload := range uploads {
		ul, err := s.codenavSvc.NewGetDefinitionsBySymbolNames(ctx, requestArgs, reqState, []string{ident})
		if err != nil {
			return nil, err
		}

		// TODO - should this ever be non-singleton?
		for syntectName := range syntectNames {
			preciseDataList = append(preciseDataList, &preciseData{
				syntectName: syntectName,
				symbolName:  ident,
				location:    ul,
			})
			// }
		}
	}

	fmt.Printf("PHASE 4\n")
	// PHASE 4: Read the files that contain a definition

	type DocumentAndText struct {
		Content string
		SCIP    *scip.Document
	}
	cache := map[string]DocumentAndText{}
	for _, pd := range preciseDataList {
		for _, l := range pd.location {
			key := fmt.Sprintf("%s@%s:%s", l.Dump.RepositoryName, l.Dump.Commit, filepath.Join(l.Dump.Root, l.Path))
			if _, ok := cache[key]; ok {
				continue
			}

			fmt.Printf("> Parsing file %q\n", key)

			// TODO - archive where possible when we fetch multiple files from the
			// same repo. Cut round trips down from one per file to one per repo,
			// and we'll likely have a lot of shared definition sources.

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

			syntectDocs, err := s.getSCIPDocumentByContent(ctx, string(file), l.Path)
			if err != nil {
				return nil, err
			}
			cache[key] = DocumentAndText{
				Content: string(file),
				SCIP:    syntectDocs,
			}
		}
	}

	fmt.Printf("PHASE 5\n")
	// PHASE 5: Extract the definitions for each of the relevant syntect symbols
	// we originally requested.
	//
	// NOTE: I make an assumption here that the symbols will be equal as
	// they were both generated by the same treesitter process. See the
	// inline note below.

	preciseResponse := []*types.PreciseContext{}
	for _, pd := range preciseDataList {
		for _, l := range pd.location {
			key := fmt.Sprintf("%s@%s:%s", l.Dump.RepositoryName, l.Dump.Commit, filepath.Join(l.Dump.Root, l.Path))
			documentAndText := cache[key]

			for _, occ := range documentAndText.SCIP.Occurrences {
				// NOTE: assumption made; we may want to look at the precise
				// range as an alternate or additional indicator for which
				// syntect occurrences we are interested in
				if occ.Symbol != pd.syntectName {
					continue
				}
				fmt.Println("THIS is the EnclosingRange", occ.EnclosingRange)
				if len(occ.EnclosingRange) > 0 {
					r := scip.NewRange(occ.EnclosingRange)
					fmt.Println("THIS is the CONTENT", documentAndText.Content)
					c := strings.Split(string(documentAndText.Content), "\n")
					snippet := extractSnippet(c, r.Start.Line, r.End.Line, r.Start.Character, r.End.Character)

					preciseResponse = append(preciseResponse, &types.PreciseContext{
						SymbolName:        pd.symbolName,
						SyntectDescriptor: pd.syntectName,
						Repository:        l.Dump.RepositoryName,
						SymbolRole:        0, // TODO
						Confidence:        "PRECISE",
						Text:              snippet,
						FilePath:          l.Path,
					})
				}
			}
		}
	}

	fmt.Printf("DONE\n")
	return preciseResponse, nil
}

// func extractSnippet(file string, startLine, endLine, startChar, endChar int32) string {
// 	lines := strings.Split(file, "\n")

// 	if startLine > endLine || startLine < 0 || endLine >= int32(len(lines)) {
// 		return ""
// 	}

// 	c := make([]string, endLine-startLine+1)
// 	copy(c, lines[startLine:endLine+1])

// 	n := len(c) - 1
// 	if n == 0 {
// 		endChar -= startChar
// 	}
// 	c[0] = c[0][startChar:]
// 	c[n] = c[n][:endChar]

// 	return strings.Join(c, "\n")
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
