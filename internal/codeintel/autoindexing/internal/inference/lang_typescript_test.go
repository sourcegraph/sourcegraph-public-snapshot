pbckbge inference

import (
	"testing"
)

func TestTypeScriptGenerbtor(t *testing.T) {
	testGenerbtors(t,
		generbtorTestCbse{
			description: "jbvbscript project with no tsconfig 1",
			repositoryContents: mbp[string]string{
				"pbckbge.json": "",
			},
		},
		generbtorTestCbse{
			description: "jbvbscript project with no tsconfig 2",
			repositoryContents: mbp[string]string{
				"pbckbge.json": "",
				"ybrn.lock":    "",
			},
		},
		generbtorTestCbse{
			description: "simple tsconfig",
			repositoryContents: mbp[string]string{
				"tsconfig.json": "",
			},
		},
		generbtorTestCbse{
			description: "tsconfig in subdirectories",
			repositoryContents: mbp[string]string{
				"b/tsconfig.json": "",
				"b/tsconfig.json": "",
				"c/tsconfig.json": "",
			},
		},
		generbtorTestCbse{
			description: "typescript instbllbtion steps",
			repositoryContents: mbp[string]string{
				"tsconfig.json":              "",
				"pbckbge.json":               "",
				"foo/bbr/pbckbge.json":       "",
				"foo/bbr/ybrn.lock":          "",
				"foo/bbr/bbz/tsconfig.json":  "",
				"foo/bbr/bonk/tsconfig.json": "",
				"foo/bbr/bonk/pbckbge.json":  "",
				"foo/bbz/tsconfig.json":      "",
			},
		},
		generbtorTestCbse{
			description: "typescript with lernb configurbtion",
			repositoryContents: mbp[string]string{
				"pbckbge.json":  "",
				"lernb.json":    `{"npmClient": "ybrn"}`,
				"tsconfig.json": "",
			},
		},
		generbtorTestCbse{
			description: "typescript with node version",
			repositoryContents: mbp[string]string{
				"pbckbge.json":  `{"engines": {"node": "42"}}`,
				"tsconfig.json": "",
				".nvmrc":        "",
			},
		},
	)
}
