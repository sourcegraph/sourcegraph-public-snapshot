package context

import (
	"context"
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"

	"github.com/sourcegraph/scip/bindings/go/scip"
	"google.golang.org/protobuf/proto"

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
		ident := name.GetIdentifier()

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

		for _, upload := range uploads {
			loc, err := s.codenavSvc.GetLocationByExplodedSymbol(ctx, ident, upload.ID, "definition_ranges")
			if err != nil {
				return nil, err
			}

			ul, err := s.codenavSvc.GetUploadLocations(ctx, args, reqState, loc, true)
			if err != nil {
				return nil, err
			}

			pd := &preciseData{
				symbolName: ident,
				repository: string(repo[0].Name),
				// symbolRole: int32(el.Occurrence.SymbolRoles),
				confidence: "PRECISE",
				location:   ul,
			}

			e := strings.Split(ident, "/")
			key := e[len(e)-1]
			definitionMap[key] = pd

			preciseDataList = append(preciseDataList, pd)
		}
	}

	snippetToPreciseDataMap := map[string]*preciseData{}
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

	c := make([]string, endLine-startLine+1)
	copy(c, lines[startLine:endLine+1])

	n := len(c) - 1
	if n == 0 {
		endChar -= startChar
	}
	c[0] = c[0][startChar:]
	c[n] = c[n][:endChar]

	return strings.Join(c, "\n")
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
