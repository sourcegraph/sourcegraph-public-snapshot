package schemas

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSchemasFromKind(t *testing.T) {
	require.Len(t, SchemasFromKind[Frontend](), 1)
	require.Len(t, SchemasFromKind[CodeIntel](), 1)
	require.Len(t, SchemasFromKind[CodeInsights](), 1)
	require.Len(t, SchemasFromKind[Production](), 2)
	require.Len(t, SchemasFromKind[Any](), 0)
}
