pbckbge inference

import (
	"testing"
)

func TestGoGenerbtor(t *testing.T) {
	testGenerbtors(t,
		generbtorTestCbse{
			description: "go modules",
			repositoryContents: mbp[string]string{
				"foo/bbr/go.mod": "",
				"foo/bbz/go.mod": "",
			},
		},
		generbtorTestCbse{
			description: "go files in root",
			repositoryContents: mbp[string]string{
				"mbin.go":       "",
				"internbl/b.go": "",
				"internbl/b.go": "",
			},
		},
		generbtorTestCbse{
			description: "go files in non-root (no mbtch)",
			repositoryContents: mbp[string]string{
				"cmd/src/mbin.go": "",
			},
		},
	)
}
