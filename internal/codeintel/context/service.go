package context

import (
	"archive/tar"
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/sourcegraph/scip/bindings/go/scip"
	"go.opentelemetry.io/otel/attribute"
	"google.golang.org/protobuf/proto"

	"github.com/sourcegraph/sourcegraph/lib/codeintel/precise"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/codenav"
	codenavtypes "github.com/sourcegraph/sourcegraph/internal/codeintel/codenav"
	codenavshared "github.com/sourcegraph/sourcegraph/internal/codeintel/codenav/shared"
	resolverstubs "github.com/sourcegraph/sourcegraph/internal/codeintel/resolvers"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/shared/symbols"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/types"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/uploads/shared"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
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
	enableSyntect                  = true
)

func (s *Service) GetPreciseContext(ctx context.Context, args *resolverstubs.GetPreciseContextInput) (_ []*types.PreciseContext, err error) {
	ctx, trace, endObservation := s.operations.getPreciseContext.With(ctx, &err, observation.Args{Attrs: []attribute.KeyValue{
		attribute.String("repositoryName", args.Input.RepositoryName),
		attribute.String("content", args.Input.ActiveFileContent),
		attribute.String("closestRemoteCommitSHA", args.Input.ClosestRemoteCommitSHA),
	}})
	defer endObservation(1, observation.Args{})

	filename := args.Input.ActiveFile
	content := args.Input.ActiveFileContent
	repoName := api.RepoName(args.Input.RepositoryName)
	activeRangeSelection := translateToScipRange(args.Input.ActiveFileSelectionRange)

	repo, err := s.repostore.GetByName(ctx, repoName)
	if err != nil {
		return nil, err
	}

	commitID, err := s.gitserverClient.ResolveRevision(ctx, repoName, args.Input.ClosestRemoteCommitSHA, gitserver.ResolveRevisionOptions{})
	if err != nil {
		return nil, err
	}
	closestRemoteCommitSHA := string(commitID)

	uploads, err := s.codenavSvc.GetClosestDumpsForBlob(ctx, int(repo.ID), closestRemoteCommitSHA, filename, true, "")
	if err != nil {
		return nil, err
	}
	trace.AddEvent("codenavSvc.GetClosestDumpsForBlob", attribute.Int("numDumps", len(uploads)))
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
	fuzzyNameSetBySymbol := map[string]map[string]struct{}{}
	if enableSyntect {
		// PHASE 1: Run current scope through treesitter
		syntectDocument, err := s.getSCIPDocumentByContent(ctx, content, filename)
		if err != nil {
			return nil, err
		}
		trace.AddEvent("contextSvc.getSCIPDocumentByContent", attribute.String("filename", filename))

		fuzzySymbolNameSet := precise.NewSet[string]()
		for _, occurrence := range syntectDocument.Occurrences {
			if activeRangeSelection == nil || precise.IsOccurrenceWithinRange(activeRangeSelection, occurrence) {
				fuzzySymbolNameSet.Add(occurrence.Symbol)
			}
		}

		fuzzySymbolNames := fuzzySymbolNameSet.ToSlice()
		sort.Strings(fuzzySymbolNames)
		trace.AddEvent("fuzzySymbolNames", attribute.StringSlice("fuzzyNames", fuzzySymbolNames))

		uploadIDs := make([]int, len(uploads))
		for i, upload := range uploads {
			uploadIDs[i] = upload.ID
		}
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
			explodedScipNames, err := s.codenavSvc.GetFullSCIPNameByDescriptor(ctx, uploadIDs, fuzzySymbolNames)
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
					ex, err := symbols.NewExplodedSymbol(fuzzyName)
					if err != nil {
						trace.AddEvent("NewExplodedSymbol error in scipNamesByFuzzyName", attribute.String("exploded symbol err", err.Error()))
						continue
					}
					fmt.Printf("\tSYNTECT: %q -> %q\n", fuzzyName, ex.DescriptorSuffix)
				}
			}
			fmt.Printf("\n\n")

			explodedScipSymbolsByFuzzyName := map[string][]*symbols.ExplodedSymbol{}
			for _, fuzzyName := range fuzzySymbolNames {
				ex, err := symbols.NewExplodedSymbol(fuzzyName)
				if err != nil {
					trace.AddEvent("NewExplodedSymbol error", attribute.String("exploded symbol err", err.Error()))
					continue
				}
				var explodedScipSymbols []*symbols.ExplodedSymbol
				for _, esn := range explodedScipNames {
					// N.B. this matches what we search against in fuzzyDescriptorSuffixConditions
					if !strings.HasSuffix(esn.DescriptorSuffix, ex.DescriptorSuffix) {
						continue
					}
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

			trace.AddEvent(
				"num of explodedScipSymbolsByFuzzyName",
				attribute.Int(
					"length of explodedScipSymbolsByFuzzyName",
					len(explodedScipSymbolsByFuzzyName),
				),
			)
			// DEBUGGING
			fmt.Printf("\n\n")

			return explodedScipSymbolsByFuzzyName, nil
		}()
		if err != nil {
			return nil, err
		}

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
	} else {
		// TODO: we might have to strip the root active file
		// TODO: if the file name starts with the root remove it
		symbolsNames, err := s.codenavSvc.GetSymbolNamesByRange(ctx, requestArgs, filename, reqState, activeRangeSelection)
		if err != nil {
			return nil, err
		}

		for _, symbolName := range symbolsNames {
			fuzzyNameSetBySymbol[symbolName] = map[string]struct{}{"": {}}
		}
	}

	// PHASE 3: Gather definitions for each relevant SCIP symbol

	type preciseData struct {
		fuzzyName      string
		scipSymbolName string
		location       []codenavshared.UploadLocation
	}
	preciseDataList := []*preciseData{}

	symbolNames := make([]string, 0, len(fuzzyNameSetBySymbol))
	for symbolName := range fuzzyNameSetBySymbol {
		symbolNames = append(symbolNames, symbolName)
	}

	ul, err := s.codenavSvc.NewGetDefinitionsBySymbolNames(ctx, requestArgs, reqState, symbolNames)
	if err != nil {
		return nil, err
	}

	for _, location := range ul {
		for fzn := range fuzzyNameSetBySymbol[location.SymbolName] {
			preciseDataList = append(preciseDataList, &preciseData{
				fuzzyName:      fzn,
				scipSymbolName: location.SymbolName,
				location:       []codenavshared.UploadLocation{location},
			})
		}
	}

	// for symbol, fuzzyNames := range fuzzyNameSetBySymbol {
	// 	// TODO - these are duplicated and should also be batched
	// 	fmt.Printf("> Fetching definitions of %q\n", symbol)

	// 	ul, err := s.codenavSvc.NewGetDefinitionsBySymbolNames(ctx, requestArgs, reqState, []string{symbol})
	// 	if err != nil {
	// 		return nil, err
	// 	}

	// 	// TODO - should this ever be non-singleton? (this will go away when we batch)
	// 	for fzn := range fuzzyNames {
	// 		preciseDataList = append(preciseDataList, &preciseData{
	// 			fuzzyName:      fzn,
	// 			scipSymbolName: symbol,
	// 			location:       ul,
	// 		})
	// 	}
	// }

	trace.AddEvent("preciseDataList", attribute.Int("fuzzyName", len(preciseDataList)))

	// DEBUGGING
	lap("PHASE 3: %d matching precise symbols with len %d \n", len(fuzzyNameSetBySymbol), len(preciseDataList))

	// PHASE 4: Read the files that contain a definition
	filesByRepo := map[string]struct {
		paths map[gitdomain.Pathspec]struct{}
		dump  shared.Dump
	}{}

	cache := map[string]DocumentAndText{}
	for _, pd := range preciseDataList {
		for _, l := range pd.location {
			// key := fmt.Sprintf("%s@%s:%s", l.Dump.RepositoryName, l.Dump.Commit, filepath.Join(l.Dump.Root, l.Path))
			// if _, ok := cache[key]; ok {
			// 	continue
			// }
			// fmt.Printf("> Parsing file %q\n", key)

			repoCommitKey := fmt.Sprintf("%s@%s", l.Dump.RepositoryName, l.Dump.Commit)

			px := filesByRepo[repoCommitKey].paths
			if px == nil {
				px = map[gitdomain.Pathspec]struct{}{}
			}
			px[gitdomain.Pathspec(l.Path)] = struct{}{}
			filesByRepo[repoCommitKey] = struct {
				paths map[gitdomain.Pathspec]struct{}
				dump  shared.Dump
			}{
				dump:  l.Dump,
				paths: px,
			}

			// // TODO - archive where possible when we fetch multiple files from the
			// // same repo. Cut round trips down from one per file to one per repo,
			// // and we'll likely have a lot of shared definition sources.
			// // TODO: Group by repo, then fetch the archive containing those files

			// file, err := s.gitserverClient.ReadFile(
			// 	ctx,
			// 	authz.DefaultSubRepoPermsChecker,
			// 	api.RepoName(l.Dump.RepositoryName),
			// 	api.CommitID(l.Dump.Commit),
			// 	l.Path,
			// )
			// if err != nil {
			// 	return nil, err
			// }

			// syntectDocs, err := s.getSCIPDocumentByContent(ctx, string(file), l.Path)
			// if err != nil {
			// 	return nil, err
			// }
			// cache[key] = NewDocumentAndText(string(file), syntectDocs)
		}
	}

	for repoCommitKey, path := range filesByRepo {
		parts := strings.Split(repoCommitKey, "@")
		repo := api.RepoName(parts[0])
		pathspec := []gitdomain.Pathspec{}
		for key := range path.paths {
			pathspec = append(pathspec, key)
		}
		opts := gitserver.ArchiveOptions{
			Treeish:   parts[1],
			Format:    gitserver.ArchiveFormatTar,
			Pathspecs: pathspec,
		}
		rc, err := s.gitserverClient.ArchiveReader(ctx, authz.DefaultSubRepoPermsChecker, repo, opts)
		if err != nil {
			return nil, err
		}
		defer rc.Close()

		tr := tar.NewReader(rc)
		for {
			header, err := tr.Next()
			if err != nil {
				if err != io.EOF {
					return nil, err
				}

				break
			}

			var buf bytes.Buffer
			if _, err := io.CopyN(&buf, tr, header.Size); err != nil {
				return nil, err
			}

			// Since we quoted all literal path specs on entry, we need to remove it from
			// the returned filepaths.
			file := buf.String()
			p := strings.TrimPrefix(header.Name, ":(literal)")
			syntectDocs, err := s.getSCIPDocumentByContent(ctx, file, p)
			if err != nil {
				return nil, err
			}
			key := fmt.Sprintf("%s@%s:%s", path.dump.RepositoryName, path.dump.Commit, filepath.Join(path.dump.Root, p))
			cache[key] = NewDocumentAndText(file, syntectDocs)
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
				if occ.Symbol != pd.fuzzyName && pd.fuzzyName != "" {
					continue
				}
				if len(occ.EnclosingRange) > 0 {
					ex, err := symbols.NewExplodedSymbol(pd.scipSymbolName)
					if err != nil {
						return nil, err
					}
					fex, err := symbols.NewExplodedSymbol(pd.fuzzyName)
					if err != nil {
						return nil, err
					}
					preciseResponse = append(preciseResponse, &types.PreciseContext{
						ScipSymbolName:        pd.scipSymbolName,
						FuzzySymbolName:       pd.fuzzyName,
						ScipDescriptorSuffix:  ex.DescriptorSuffix,
						FuzzyDescriptorSuffix: fex.FuzzyDescriptorSuffix,
						RepositoryName:        l.Dump.RepositoryName,
						Text:                  documentAndText.Extract(scip.NewRange(occ.EnclosingRange)),
						FilePath:              l.Path,
					})
				}
			}
		}
	}
	trace.AddEvent("preciseResponse", attribute.Int("length of preciseResponse", len(preciseResponse)))

	// DEBUGGING
	lap("PHASE 5: generated %s context items\n", len(preciseResponse))
	return preciseResponse, nil
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

func translateToScipRange(ar *resolverstubs.ActiveFileSelectionRangeInput) (r *scip.Range) {
	if ar == nil {
		return nil
	}

	return &scip.Range{
		Start: scip.Position{Line: int32(ar.StartLine), Character: int32(ar.StartCharacter)},
		End:   scip.Position{Line: int32(ar.EndLine), Character: int32(ar.EndCharacter)},
	}
}
