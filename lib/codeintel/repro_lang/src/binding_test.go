package repro_lang

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
		"(source_file (statement (definitionStatement name: (identifier) roles: (implementationRelation (identifier)))))",
		n.String(),
	)
}
