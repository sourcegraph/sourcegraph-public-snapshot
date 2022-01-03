package java

import (
	"context"

	javaGrammar "github.com/smacker/go-tree-sitter/java"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/lsif_typed"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/search-based/api"
)

type Indexer struct{}

var _ api.Indexer = Indexer{}

func (_ Indexer) Name() string {
	return "java"
}

func (_ Indexer) FileExtensions() []string {
	return []string{".java"}
}

func (_ Indexer) Index(ctx context.Context, input *api.Input) (*lsif_typed.Document, error) {
	return api.Index(ctx, input, javaGrammar.GetLanguage(), api.LocalIntelGrammar{
		Identifiers: map[string]struct{}{"identifier": {}},
		Fingerprints: []api.DefinitionFingerprint{
			{
				ParentTypes:      []string{"identifier", "variable_declarator", "local_variable_declaration", "block"},
				ParentFieldNames: []string{"name", "declarator"},
			},
			{
				ParentTypes:      []string{"identifier", "formal_parameter", "formal_parameters", "method_declaration"},
				ParentFieldNames: []string{"name", "", "parameters"},
			},
		},
	})
}
