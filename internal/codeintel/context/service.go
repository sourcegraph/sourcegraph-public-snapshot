package context

import (
	"context"
	"encoding/base64"
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/sourcegraph/scip/bindings/go/scip"
	"go.opentelemetry.io/otel/attribute"
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

func (s *Service) GetPreciseContext(ctx context.Context, args *resolverstubs.GetPreciseContextInput) (_ []*types.PreciseContext, err error) {
	ctx, trace, endObservation := s.operations.getPreciseContext.With(ctx, &err, observation.Args{Attrs: []attribute.KeyValue{
		attribute.String("filename", args.Input.Repository),
		attribute.String("content", args.Input.ActiveFileContent),
		attribute.String("closestRemoteCommitSHA", args.Input.ClosestRemoteCommitSHA),
	}})
	defer endObservation(1, observation.Args{})

	// TODO: s.operations.getPreciseContext.With(ctx, ...)

	filename := args.Input.ActiveFile
	content := args.Input.ActiveFileContent
	closestRemoteCommitSHA := args.Input.ClosestRemoteCommitSHA

	repo, err := s.repostore.GetByName(ctx, api.RepoName(args.Input.Repository))
	if err != nil {
		return nil, err
	}

	uploads, err := s.codenavSvc.GetClosestDumpsForBlob(ctx, int(repo.ID), closestRemoteCommitSHA, filename, true, "")
	if err != nil {
		return nil, err
	}
	trace.AddEvent("codenavSvc.GetClosestDumpsForBlob",
		attribute.Int("numDumps", len(uploads)),
	)
	if len(uploads) == 0 {
		return nil, nil
	}

	requestArgs := codenavtypes.RequestArgs{
		RepositoryID: int(repo.ID),
		Commit:       closestRemoteCommitSHA,
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
		closestRemoteCommitSHA,
		"",
		maximumIndexesPerMonikerSearch,
		hunkCache,
	)

	// DEBUGGING
	start := time.Now()
	phaseStart := start
	lap := func(format string, args ...any) {
		n := time.Now()
		delta := n.Sub(phaseStart)
		phaseStart = n
		fmt.Printf("\t[%s]: %s\n", delta, fmt.Sprintf(format, args...))
	}
	fmt.Printf("> CONTEXT API\n")
	defer func() { fmt.Printf("< CONTEXT API done in %s (%s)\n", time.Since(start), err) }()

	// PHASE 1: Run current scope through treesitter

	syntectDocument, err := s.getSCIPDocumentByContent(ctx, content, filename)
	if err != nil {
		return nil, err
	}
	trace.AddEvent("contextSvc.getSCIPDocumentByContent",
		attribute.String("filename", filename),
	)

	fuzzySymbolNameMap := map[string]struct{}{}
	for _, occurrence := range syntectDocument.Occurrences {
		fuzzySymbolNameMap[occurrence.Symbol] = struct{}{}
	}
	fuzzySymbolNames := make([]string, 0, len(fuzzySymbolNameMap))
	for fuzzyName := range fuzzySymbolNameMap {
		trace.AddEvent("symbolNameMap", attribute.String("symbolName", fuzzyName)) // TODO - batch
		fuzzySymbolNames = append(fuzzySymbolNames, fuzzyName)
	}
	sort.Strings(fuzzySymbolNames)

	// DEBUGGING
	lap("PHASE 1: %d symbols from %s: %v\n", len(fuzzySymbolNames), filename, fuzzySymbolNames)

	// PHASE 2: Run treesitter output through a translation layer so we can do
	// the graph navigation in "SCIP-world" using proper identifiers. The following
	// code is pretty sloppy right now since we haven't consolidated on a single way
	// to "match" descriptors together. This should align in the db layer as well.
	//
	// This isn't a deep technical problem, just one of deciding on a thing and
	// conforming to/communicating it in the codebase.

	// Construct a map from syntect (fuzzy) name to a list of SCIP names matching the syntect
	// output. I'd like to have the `GetFullSCIPNameByDescriptor` method create this
	// mapping instead. This block should become a single function call after that
	// transformation.

	scipNamesByFuzzyName, err := func() (map[string][]*symbols.ExplodedSymbol, error) {
		// TODO: Either pass a slice of uploads or loop thru uploads and pass the ids to fix this hardcoding of the first
		// TODO: does it make more sense to return a map to avoid having to loop thru them on line 189?
		explodedScipNames, err := s.codenavSvc.GetFullSCIPNameByDescriptor(ctx, []int{uploads[0].ID}, fuzzySymbolNames)
		if err != nil {
			return nil, err
		}

		// DEBUGGING
		for _, scipName := range explodedScipNames {
			if strings.Contains(scipName.DescriptorSuffix, "Runner") {
				fmt.Printf("\tSCIP:%q\n", scipName.DescriptorSuffix)
			}
		}
		for _, fuzzyName := range fuzzySymbolNames {
			if strings.Contains(fuzzyName, "Runner") {
				ex, _ := symbols.NewExplodedSymbol(fuzzyName)
				fmt.Printf("\tSYNTECT: %q -> %q\n", fuzzyName, ex.DescriptorSuffix)
			}
		}
		fmt.Printf("\n\n")

		explodedScipSymbolsByFuzzyName := map[string][]*symbols.ExplodedSymbol{}
		for _, fuzzyName := range fuzzySymbolNames {
			// TODO: Don't swallow error
			ex, _ := symbols.NewExplodedSymbol(fuzzyName)
			var explodedScipSymbols []*symbols.ExplodedSymbol
			for _, esn := range explodedScipNames {
				// N.B. this matches what we search against in formatSymbolNamesToLikeClause
				if !strings.HasSuffix(esn.DescriptorSuffix, ex.DescriptorSuffix) {
					continue
				}
				// TODO - batch
				trace.AddEvent(
					"scipNames DescriptorSuffix or DescriptorSuffix",
					attribute.String("DescriptorSuffix", ex.DescriptorSuffix),
					attribute.String("DescriptorSuffix", esn.DescriptorSuffix),
				)

				explodedScipSymbols = append(explodedScipSymbols, esn)
			}

			// DEBUGGING
			if len(explodedScipSymbols) == 0 {
				ex, _ := symbols.NewExplodedSymbol(fuzzyName)
				if strings.Contains(fuzzyName, "Runner") {
					fmt.Printf("> NO MATCHES FOR %q (%q)??\n", fuzzyName, ex.DescriptorSuffix)
				}
			}

			if len(explodedScipSymbols) > 20 {
				// DEBUGGING
				fmt.Printf("TOO MANY RESULTS FOR %q\n", fuzzyName)
				trace.AddEvent("TOO MANY RESULTS", attribute.String("syntectName", fuzzyName))
				explodedScipSymbols = nil
			}

			if len(explodedScipSymbols) > 0 {
				explodedScipSymbolsByFuzzyName[fuzzyName] = explodedScipSymbols
			}
		}
		// DEBUGGING
		fmt.Printf("\n\n")

		return explodedScipSymbolsByFuzzyName, nil
	}()
	if err != nil {
		return nil, err
	}

	fuzzyNameSetBySymbol := map[string]map[string]struct{}{}
	for fuzzyName, explodedSymbols := range scipNamesByFuzzyName {
		for _, explodedSymbol := range explodedSymbols {
			symbol := explodedSymbol.Symbol()
			if _, ok := fuzzyNameSetBySymbol[symbol]; !ok {
				fuzzyNameSetBySymbol[symbol] = map[string]struct{}{}
			}

			fuzzyNameSetBySymbol[symbol][fuzzyName] = struct{}{}
		}
	}

	// DEBUGGING
	lap("PHASE 2: %d matching precise symbols\n", len(fuzzyNameSetBySymbol))

	// PHASE 3: Gather definitions for each relevant SCIP symbol

	type preciseData struct {
		fuzzyName      string
		scipSymbolName string
		location       []codenavshared.UploadLocation
	}
	preciseDataList := []*preciseData{}

	for symbol, fuzzyNames := range fuzzyNameSetBySymbol {
		// TODO - these are duplicated and should also be batched
		fmt.Printf("> Fetching definitions of %q\n", symbol)

		ul, err := s.codenavSvc.NewGetDefinitionsBySymbolNames(ctx, requestArgs, reqState, []string{symbol})
		if err != nil {
			return nil, err
		}

		// TODO - should this ever be non-singleton?
		for fzn := range fuzzyNames {
			preciseDataList = append(preciseDataList, &preciseData{
				fuzzyName:      fzn,
				scipSymbolName: symbol,
				location:       ul,
			})
			// TODO - batch
			trace.AddEvent(
				"preciseDataList",
				attribute.String("fuzzyName", fzn),
				attribute.String("symbolName", symbol),
			)
		}
	}

	// DEBUGGING
	lap("PHASE 3: %d matching precise symbols\n", len(fuzzyNameSetBySymbol))

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

	// DEBUGGING
	lap("PHASE 4: read %d files\n", len(cache))

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
				if occ.Symbol != pd.fuzzyName {
					continue
				}
				fmt.Println("THIS is the EnclosingRange", occ.EnclosingRange)
				if len(occ.EnclosingRange) > 0 {
					r := scip.NewRange(occ.EnclosingRange)
					fmt.Println("THIS is the CONTENT", documentAndText.Content)
					c := strings.Split(string(documentAndText.Content), "\n")
					snippet := extractSnippet(c, r.Start.Line, r.End.Line, r.Start.Character, r.End.Character)

					preciseResponse = append(preciseResponse, &types.PreciseContext{
						ScipSymbolName:  pd.scipSymbolName,
						FuzzySymbolName: pd.fuzzyName,
						RepositoryName:  l.Dump.RepositoryName,
						SymbolRole:      0, // TODO
						Confidence:      "PRECISE",
						Text:            snippet,
						FilePath:        l.Path,
					})

					// TODO - batch?
					trace.AddEvent(
						"preciseResponse",
						attribute.String("symbolName", pd.scipSymbolName),
						attribute.String("fuzzyName", pd.fuzzyName),
						attribute.String("repository", l.Dump.RepositoryName),
						attribute.String("filePath", l.Path),
					)
				}
			}
		}
	}

	// DEBUGGING
	lap("PHASE 5: generated %s context items\n", len(preciseResponse))
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
