package src

import (
	"context"
	"testing"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/stretchr/testify/assert"
)

func TestParse(t *testing.T) {
	assert := assert.New(t)

	n, err := sitter.ParseCtx(context.Background(), []byte("definition a implements b\n"), GetLanguage())
	assert.NoError(err)
	assert.Equal(
		"(source_file (definition_statement name: (identifier workspace: (workspace_identifier)) roles: (implementation_relation name: (identifier workspace: (workspace_identifier)))))",
		n.String(),
	)
}
