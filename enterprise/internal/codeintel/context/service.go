package context

import (
	"context"
	"encoding/base64"
	"fmt"
	"strconv"

	"google.golang.org/protobuf/proto"

	//"github.com/sourcegraph/scip/bindings/go/scip"

	"github.com/sourcegraph/scip/bindings/go/scip"

	"github.com/sourcegraph/sourcegraph/internal/api"
	resolverstubs "github.com/sourcegraph/sourcegraph/internal/codeintel/resolvers"

	scipstore "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/context/internal/scipstore"

	"github.com/sourcegraph/sourcegraph/internal/gosyntect"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type Service struct {
	codenavSvc    CodeNavService
	scipstore     scipstore.ScipStore
	syntectClient *gosyntect.Client
	operations    *operations
}

func newService(
	observationCtx *observation.Context,
	scipstore scipstore.ScipStore,
	syntectClient *gosyntect.Client,
	codenavSvc CodeNavService,
) *Service {
	return &Service{
		codenavSvc:    codenavSvc,
		scipstore:     scipstore,
		syntectClient: syntectClient,
		operations:    newOperations(observationCtx),
	}
}

func (s *Service) FindMostRelevantSCIPSymbols(ctx context.Context, args *resolverstubs.FindMostRelevantSCIPSymbolsArgs) (string, error) {
	filename := args.Args.EditorState.ActiveFile
	content := args.Args.EditorState.ActiveFileContent
	commitID := args.Args.CommitID
	repoName := "github.com/Numbers88s/simple-mock-go"

	//! TODO: Change repository name to repository ID
	repoID, err := strconv.Atoi(args.Args.Repository)
	if err != nil {
		return "", err
	}

	// TODO: Get uploads from codenav service
	uploads, err := s.codenavSvc.GetClosestDumpsForBlob(ctx, repoID, commitID, filename, true, "")
	if err != nil {
		return "", err
	}
	fmt.Println("HERE is the upload", uploads)

	// Step #1: Get scip document created by treesitter
	document, err := s.getSCIPDocumentByContent(ctx, content, filename)
	if err != nil {
		return "", err
	}

	symbolNames := make([]string, 0, len(document.Occurrences))
	for _, occurrence := range document.Occurrences {
		symbolNames = append(symbolNames, occurrence.Symbol)
	}
	// // hack for now
	// uploadID := uploads[0].ID
	// fmt.Println("HERE IT IS a NAME", symbolNames)
	// fmt.Println("HERE IT IS a uploadID", uploadID)

	scipDoc, err := s.codenavSvc.GetSCIPDocumentsBySymbolNames(ctx, uploads, symbolNames, api.RepoName(repoName), api.RepoID(repoID), api.CommitID(commitID), filename)
	if err != nil {
		return "", err
	}
	fmt.Println("HERE IT IS a scipDoc", scipDoc)

	return "", nil
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
