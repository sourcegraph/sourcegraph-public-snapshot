package perforce

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/perforce"
)

func TestParseP4Protects(t *testing.T) {
	protectsOut := []byte(`{"depotFile":"//...","host":"*","line":"1","perm":"list","unmap":"","user":"*"}
{"depotFile":"//test-perms/Frontend/...","host":"*","isgroup":"","line":"6","perm":"write","user":"Frontend"}
{"depotFile":"//integration-test-depot/...","host":"*","isgroup":"","line":"14","perm":"=read","unmap":"","user":"all"}
{"depotFile":"//go/...","host":"*","line":"24","perm":"read","user":"bob"}
{"depotFile":"//go/api/...","host":"*","line":"25","perm":"=read","unmap":"","user":"bob"}
{"depotFile":"//go/*/except.txt","host":"*","isgroup":"","line":"26","perm":"read","user":"Frontend"}
{"depotFile":"//go/...","host":"192.168.10.1/24","line":"27","perm":"=read","unmap":"","user":"bob"}
`)

	protects, err := parseP4Protects(protectsOut)
	require.NoError(t, err)

	want := []*perforce.Protect{
		{
			Level:       "list",
			EntityType:  "user",
			EntityName:  "*",
			Match:       "//...",
			IsExclusion: true,
			Host:        "*",
		},
		{
			Level:       "write",
			EntityType:  "group",
			EntityName:  "Frontend",
			Match:       "//test-perms/Frontend/...",
			IsExclusion: false,
			Host:        "*",
		},
		{
			Level:       "=read",
			EntityType:  "group",
			EntityName:  "all",
			Match:       "//integration-test-depot/...",
			IsExclusion: true,
			Host:        "*",
		},
		{
			Level:       "read",
			EntityType:  "user",
			EntityName:  "bob",
			Match:       "//go/...",
			IsExclusion: false,
			Host:        "*",
		},
		{
			Level:       "=read",
			EntityType:  "user",
			EntityName:  "bob",
			Match:       "//go/api/...",
			IsExclusion: true,
			Host:        "*",
		},
		{
			Level:       "read",
			EntityType:  "group",
			EntityName:  "Frontend",
			Match:       "//go/*/except.txt",
			IsExclusion: false,
			Host:        "*",
		},
		{
			Level:       "=read",
			EntityType:  "user",
			EntityName:  "bob",
			Match:       "//go/...",
			IsExclusion: true,
			Host:        "192.168.10.1/24",
		},
	}

	require.Equal(t, want, protects)
}
