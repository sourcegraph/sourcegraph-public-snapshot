package context

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"

	"google.golang.org/protobuf/proto"

	"github.com/sourcegraph/scip/bindings/go/scip"

	resolverstubs "github.com/sourcegraph/sourcegraph/internal/codeintel/resolvers"

	scipstore "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/context/internal/scipstore"

	"github.com/sourcegraph/sourcegraph/internal/gosyntect"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type Service struct {
	scipstore  scipstore.ScipStore
	syntect    *gosyntect.Client
	operations *operations
}

// var syntectClient *gosyntect.Client
var contextConfigInst = &config{}

func newService(
	observationCtx *observation.Context,
	scipstore scipstore.ScipStore,
) *Service {
	// config := env.BaseConfig.Get()
	// syntectServer := env.Get("SRC_SYNTECT_SERVER", "http://syntect-server:9238", "syntect_server HTTP(s) address")
	// contextConfigInst.Load()
	// syntectServer := contextConfigInst.syntectServer
	// syntectServer := conf.Get()
	syntectServer := os.Getenv("SRC_SYNTECT_SERVER")
	fmt.Println("SYNTECT SERVER: ", syntectServer)
	if syntectServer == "" {
		syntectServer = "http://syntect-server:9238"
	}

	return &Service{
		scipstore:  scipstore,
		syntect:    gosyntect.New(syntectServer),
		operations: newOperations(observationCtx),
	}
}

type SyntecSymbolQueryArgs struct {
	fileName string
	content  string
}

func (s *Service) FindMostRelevantSCIPSymbols(ctx context.Context, args *resolverstubs.FindMostRelevantSCIPSymbolsArgs) (string, error) {
	fmt.Println("CALLED ")
	filename := args.Args.EditorState.ActiveFile
	content := args.Args.EditorState.ActiveFileContent
	fmt.Println("FILENAME: ", filename)
	fmt.Println("CONTENT: ", content)
	// filetypeQuery := highlight.DetectSyntaxHighlightingLanguage(filename, content)

	// engine := highlight.GetEngineParameter(filetypeQuery.Engine)
	q := gosyntect.SymbolsQuery{
		FileName: filename,
		Content:  content,
		// Engine:   engine,
	}

	resp, err := s.syntect.Symbols(ctx, &q)
	// resp, err := highlight.Symbols(ctx, &q)
	if err != nil {
		return "", err
	}

	fmt.Println("RESPONSE: ", resp)

	d, err := base64.StdEncoding.DecodeString(resp.Scip)
	if err != nil {
		fmt.Println("ERROR: ", err)
		return "", err
	}

	var document scip.Document
	if err := proto.Unmarshal(d, &document); err != nil {
		fmt.Println("ERROR: ", err)
		return "", err
	}

	fmt.Println("HERE IT IS", document)

	_ = resp

	return "", nil
}

func (s *Service) SplitIntoEmbeddableChunks(ctx context.Context, text string, fileName string, splitOptions SplitOptions) ([]EmbeddableChunk, error) {
	return SplitIntoEmbeddableChunks(text, fileName, splitOptions), nil
}
