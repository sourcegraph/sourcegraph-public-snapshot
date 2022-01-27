package inference

import (
	"strconv"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/sourcegraph/lib/codeintel/autoindex/config"
)

func TestTypeScriptPatterns(t *testing.T) {
	testLangPatterns(t, TypeScriptPatterns(), []PathTestCase{
		{"tsconfig.json", true},
		{"tsconfig.json/subdir", false},
		{".nvmrc", true},
		{"subdir/package.json", true},
		{"subdir/yarn.lock", true},
	})
}

func TestInferTypeScriptIndexJobsMissingTsConfig(t *testing.T) {
	testCases := []struct {
		paths   []string
		command string
	}{
		{[]string{"package.json"}, "npm install --ignore-scripts"},
		{[]string{"yarn.lock", "package.json"}, "yarn --ignore-engines --ignore-scripts"},
	}

	for _, testCase := range testCases {
		expectedIndexJobs := []config.IndexJob{
			{
				Steps:       []config.DockerStep{{Image: "sourcegraph/lsif-node:autoindex", Commands: []string{testCase.command}}},
				Root:        "",
				Indexer:     lsifTscImage,
				IndexerArgs: []string{"lsif-tsc", "-p", ".", "--inferTSConfig"},
				Outfile:     "",
			},
		}
		if diff := cmp.Diff(expectedIndexJobs, InferTypeScriptIndexJobs(NewMockGitClient(), testCase.paths)); diff != "" {
			t.Errorf("unexpected index jobs (-want +got):\n%s", diff)
		}
	}
}

func TestInferTypeScriptIndexJobsTsConfigRoot(t *testing.T) {
	paths := []string{
		"tsconfig.json",
	}

	expectedIndexJobs := []config.IndexJob{
		{
			Steps:       nil,
			Root:        "",
			Indexer:     lsifTscImage,
			IndexerArgs: []string{"lsif-tsc", "-p", "."},
			Outfile:     "",
		},
	}
	if diff := cmp.Diff(expectedIndexJobs, InferTypeScriptIndexJobs(NewMockGitClient(), paths)); diff != "" {
		t.Errorf("unexpected index jobs (-want +got):\n%s", diff)
	}
}

func TestInferTypeScriptIndexJobsTsConfigSubdirs(t *testing.T) {
	paths := []string{
		"a/tsconfig.json",
		"b/tsconfig.json",
		"c/tsconfig.json",
	}

	expectedIndexJobs := []config.IndexJob{
		{
			Steps:       nil,
			Root:        "a",
			Indexer:     lsifTscImage,
			IndexerArgs: []string{"lsif-tsc", "-p", "."},
			Outfile:     "",
		},
		{
			Steps:       nil,
			Root:        "b",
			Indexer:     lsifTscImage,
			IndexerArgs: []string{"lsif-tsc", "-p", "."},
			Outfile:     "",
		},
		{
			Steps:       nil,
			Root:        "c",
			Indexer:     lsifTscImage,
			IndexerArgs: []string{"lsif-tsc", "-p", "."},
			Outfile:     "",
		},
	}
	if diff := cmp.Diff(expectedIndexJobs, InferTypeScriptIndexJobs(NewMockGitClient(), paths)); diff != "" {
		t.Errorf("unexpected index jobs (-want +got):\n%s", diff)
	}
}

func TestInferTypeScriptIndexJobsInstallSteps(t *testing.T) {
	paths := []string{
		"tsconfig.json",
		"package.json",
		"foo/baz/tsconfig.json",
		"foo/bar/baz/tsconfig.json",
		"foo/bar/bonk/tsconfig.json",
		"foo/bar/bonk/package.json",
		"foo/bar/package.json",
		"foo/bar/yarn.lock",
	}

	expectedIndexJobs := []config.IndexJob{
		{
			Steps: []config.DockerStep{
				{
					Root:     "",
					Image:    lsifTscImage,
					Commands: []string{"npm install"},
				},
			},
			Root:        "",
			Indexer:     lsifTscImage,
			IndexerArgs: []string{"lsif-tsc", "-p", "."},
			Outfile:     "",
		},
		{
			Steps: []config.DockerStep{
				{
					Root:     "",
					Image:    lsifTscImage,
					Commands: []string{"npm install"},
				},
			},
			Root:        "foo/baz",
			Indexer:     lsifTscImage,
			IndexerArgs: []string{"lsif-tsc", "-p", "."},
			Outfile:     "",
		},
		{
			Steps: []config.DockerStep{
				{
					Root:     "",
					Image:    lsifTscImage,
					Commands: []string{"npm install"},
				},
				{
					Root:     "foo/bar",
					Image:    lsifTscImage,
					Commands: []string{"yarn --ignore-engines"},
				},
			},
			Root:        "foo/bar/baz",
			Indexer:     lsifTscImage,
			IndexerArgs: []string{"lsif-tsc", "-p", "."},
			Outfile:     "",
		},
		{
			Steps: []config.DockerStep{
				{
					Root:     "",
					Image:    lsifTscImage,
					Commands: []string{"npm install"},
				},
				{
					Root:     "foo/bar",
					Image:    lsifTscImage,
					Commands: []string{"yarn --ignore-engines"},
				},
				{
					Root:     "foo/bar/bonk",
					Image:    lsifTscImage,
					Commands: []string{"npm install"},
				},
			},
			Root:        "foo/bar/bonk",
			Indexer:     lsifTscImage,
			IndexerArgs: []string{"lsif-tsc", "-p", "."},
			Outfile:     "",
		},
	}
	if diff := cmp.Diff(expectedIndexJobs, InferTypeScriptIndexJobs(NewMockGitClient(), paths)); diff != "" {
		t.Errorf("unexpected index jobs (-want +got):\n%s", diff)
	}
}

func TestInferTypeScriptIndexJobsTscLernaConfig(t *testing.T) {
	mockGit := NewMockGitClient()
	// this is kinda tied to the order in which the impl calls them :(
	{
		mockGit.RawContentsFunc.PushReturn([]byte(`{"npmClient": "yarn"}`), nil)
		mockGit.RawContentsFunc.PushReturn([]byte(`{}`), nil)
	}
	{
		mockGit.RawContentsFunc.PushReturn([]byte(`{}`), nil)
		mockGit.RawContentsFunc.PushReturn([]byte(`{"npmClient": "npm"}`), nil)
		mockGit.RawContentsFunc.PushReturn([]byte(`{}`), nil)
	}
	{
		mockGit.RawContentsFunc.PushReturn([]byte(`{"npmClient": "yarn"}`), nil)
		mockGit.RawContentsFunc.PushReturn([]byte(`{}`), nil)
	}

	paths := [][]string{
		{
			"package.json",
			"lerna.json",
			"tsconfig.json",
		},
		{
			"package.json",
			"lerna.json",
			"tsconfig.json",
		},
		{
			"package.json",
			"tsconfig.json",
		},
		{
			"foo/package.json",
			"yarn.lock",
			"lerna.json",
			"package.json",
			"foo/bar/tsconfig.json",
		},
	}

	expectedJobs := [][]config.IndexJob{
		{
			{
				Steps: []config.DockerStep{
					{
						Root:     "",
						Image:    lsifTscImage,
						Commands: []string{"yarn --ignore-engines"},
					},
				},
				LocalSteps:  nil,
				Root:        "",
				Indexer:     lsifTscImage,
				IndexerArgs: []string{"lsif-tsc", "-p", "."},
				Outfile:     "",
			},
		},
		{
			{
				Steps: []config.DockerStep{
					{
						Root:     "",
						Image:    lsifTscImage,
						Commands: []string{"npm install"},
					},
				},
				LocalSteps:  nil,
				Root:        "",
				Indexer:     lsifTscImage,
				IndexerArgs: []string{"lsif-tsc", "-p", "."},
				Outfile:     "",
			},
		},
		{
			{
				Steps: []config.DockerStep{
					{
						Root:     "",
						Image:    lsifTscImage,
						Commands: []string{"npm install"},
					},
				},
				LocalSteps:  nil,
				Root:        "",
				Indexer:     lsifTscImage,
				IndexerArgs: []string{"lsif-tsc", "-p", "."},
				Outfile:     "",
			},
		},
		{
			{
				Steps: []config.DockerStep{
					{
						Root:     "",
						Image:    lsifTscImage,
						Commands: []string{"yarn --ignore-engines"},
					},
					{
						Root:     "foo",
						Image:    lsifTscImage,
						Commands: []string{"yarn --ignore-engines"},
					},
				},
				LocalSteps:  nil,
				Root:        "foo/bar",
				Indexer:     lsifTscImage,
				IndexerArgs: []string{"lsif-tsc", "-p", "."},
				Outfile:     "",
			},
		},
	}

	for i, paths := range paths {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			if diff := cmp.Diff(expectedJobs[i], InferTypeScriptIndexJobs(mockGit, paths)); diff != "" {
				t.Errorf("unexpected index jobs (-want +got):\n%s", diff)
			}
		})
	}
}

func TestInferTypeScriptIndexJobsNodeVersionInferrence(t *testing.T) {
	mockGit := NewMockGitClient()
	mockGit.RawContentsFunc.PushReturn([]byte(""), nil)
	mockGit.RawContentsFunc.PushReturn([]byte(`{"engines":{"node":"420"}}`), nil)

	paths := [][]string{
		{
			"package.json",
			"tsconfig.json",
			".nvmrc",
		},
		{
			"tsconfig.json",
			"package.json",
		},
	}

	expectedJobs := [][]config.IndexJob{
		{
			{
				Steps: []config.DockerStep{
					{
						Root:     "",
						Image:    lsifTscImage,
						Commands: []string{nMuslCommand, "npm install"},
					},
				},
				LocalSteps: []string{
					nMuslCommand,
				},
				Root:        "",
				Indexer:     lsifTscImage,
				IndexerArgs: []string{"lsif-tsc", "-p", "."},
				Outfile:     "",
			},
		},
		{
			{
				Steps: []config.DockerStep{
					{
						Root:     "",
						Image:    lsifTscImage,
						Commands: []string{nMuslCommand, "npm install"},
					},
				},
				LocalSteps: []string{
					nMuslCommand,
				},
				Root:        "",
				Indexer:     lsifTscImage,
				IndexerArgs: []string{"lsif-tsc", "-p", "."},
				Outfile:     "",
			},
		},
	}

	for i, paths := range paths {
		if diff := cmp.Diff(expectedJobs[i], InferTypeScriptIndexJobs(mockGit, paths)); diff != "" {
			t.Errorf("unexpected index jobs (-want +got):\n%s", diff)
		}
	}
}
