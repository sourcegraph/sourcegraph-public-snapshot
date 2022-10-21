package codeintel

import (
	"fmt"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/executorqueue/handler"
	apiclient "github.com/sourcegraph/sourcegraph/enterprise/internal/executor"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/shared/types"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	srccli "github.com/sourcegraph/sourcegraph/internal/src-cli"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestTransformRecord(t *testing.T) {
	index := types.Index{
		ID:             42,
		Commit:         "deadbeef",
		RepositoryName: "linux",
		DockerSteps: []types.DockerStep{
			{
				Image:    "alpine",
				Commands: []string{"yarn", "install"},
				Root:     "web",
			},
		},
		Root:    "web",
		Indexer: "lsif-node",
		IndexerArgs: []string{
			"index",
			"-p", ".",
			// Verify args are properly shell quoted.
			"-author", "Test User",
		},
		Outfile: "",
	}
	conf.Mock(&conf.Unified{SiteConfiguration: schema.SiteConfiguration{ExternalURL: "https://test.io"}})
	t.Cleanup(func() {
		conf.Mock(nil)
	})

	job, err := transformRecord(index, handler.ResourceMetadata{}, "hunter2")
	if err != nil {
		t.Fatalf("unexpected error transforming record: %s", err)
	}

	expected := apiclient.Job{
		ID:                  42,
		Commit:              "deadbeef",
		RepositoryName:      "linux",
		ShallowClone:        true,
		FetchTags:           false,
		VirtualMachineFiles: nil,
		DockerSteps: []apiclient.DockerStep{
			{
				Key:      "pre-index.0",
				Image:    "alpine",
				Commands: []string{"yarn", "install"},
				Dir:      "web",
			},
			{
				Key:      "indexer",
				Image:    "lsif-node",
				Commands: []string{"index -p . -author 'Test User'"},
				Dir:      "web",
			},
			{
				Key:   "upload",
				Image: fmt.Sprintf("sourcegraph/src-cli:%s", srccli.MinimumVersion),
				Commands: []string{
					strings.Join(
						[]string{
							"src",
							"lsif", "upload",
							"-no-progress",
							"-repo", "linux",
							"-commit", "deadbeef",
							"-root", "web",
							"-upload-route", "/.executors/lsif/upload",
							"-file", "dump.lsif",
							"-associated-index-id", "42",
						},
						" ",
					),
				},
				Dir: "web",
				Env: []string{
					"SRC_ENDPOINT=https://test.io",
					"SRC_HEADER_AUTHORIZATION=token-executor hunter2",
				},
			},
		},
		RedactedValues: map[string]string{
			"hunter2":                "PASSWORD_REMOVED",
			"token-executor hunter2": "token-executor REDACTED",
		},
	}
	if diff := cmp.Diff(expected, job); diff != "" {
		t.Errorf("unexpected job (-want +got):\n%s", diff)
	}
}

func TestTransformRecordWithoutIndexer(t *testing.T) {
	index := types.Index{
		ID:             42,
		Commit:         "deadbeef",
		RepositoryName: "linux",
		DockerSteps: []types.DockerStep{
			{
				Image:    "alpine",
				Commands: []string{"yarn", "install"},
				Root:     "web",
			},
			{
				Image:    "lsif-node",
				Commands: []string{"index", "-p", "."},
				Root:     "web",
			},
		},
		Root:        "",
		Indexer:     "",
		IndexerArgs: nil,
		Outfile:     "other/path/lsif.dump",
	}
	conf.Mock(&conf.Unified{SiteConfiguration: schema.SiteConfiguration{ExternalURL: "https://test.io"}})
	t.Cleanup(func() {
		conf.Mock(nil)
	})

	job, err := transformRecord(index, handler.ResourceMetadata{}, "hunter2")
	if err != nil {
		t.Fatalf("unexpected error transforming record: %s", err)
	}

	expected := apiclient.Job{
		ID:                  42,
		Commit:              "deadbeef",
		RepositoryName:      "linux",
		ShallowClone:        true,
		FetchTags:           false,
		VirtualMachineFiles: nil,
		DockerSteps: []apiclient.DockerStep{
			{
				Key:      "pre-index.0",
				Image:    "alpine",
				Commands: []string{"yarn", "install"},
				Dir:      "web",
			},
			{
				Key:      "pre-index.1",
				Image:    "lsif-node",
				Commands: []string{"index", "-p", "."},
				Dir:      "web",
			},
			{
				Key:   "upload",
				Image: fmt.Sprintf("sourcegraph/src-cli:%s", srccli.MinimumVersion),
				Commands: []string{
					strings.Join(
						[]string{
							"src",
							"lsif", "upload",
							"-no-progress",
							"-repo", "linux",
							"-commit", "deadbeef",
							"-root", ".",
							"-upload-route", "/.executors/lsif/upload",
							"-file", "other/path/lsif.dump",
							"-associated-index-id", "42",
						},
						" ",
					),
				},
				Dir: "",
				Env: []string{
					"SRC_ENDPOINT=https://test.io",
					"SRC_HEADER_AUTHORIZATION=token-executor hunter2",
				},
			},
		},
		RedactedValues: map[string]string{
			"hunter2":                "PASSWORD_REMOVED",
			"token-executor hunter2": "token-executor REDACTED",
		},
	}
	if diff := cmp.Diff(expected, job); diff != "" {
		t.Errorf("unexpected job (-want +got):\n%s", diff)
	}
}

func TestTransformRecordWithResourceReplacement(t *testing.T) {
	index := types.Index{
		ID:             42,
		Commit:         "deadbeef",
		RepositoryName: "linux",
		DockerSteps: []types.DockerStep{
			{
				Image:    "alpine",
				Commands: []string{"env", "CUSTOM_MAX_MEM=$VM_MEM_GB", "yarn", "install"},
				Root:     "web",
			},
		},
		Root:        "web",
		Indexer:     "lsif-node",
		IndexerArgs: []string{"env", "MAX_DISK=$VM_DISK_MB", "DEBUG=$VM_DISK", "index"},
		Outfile:     "",
	}
	conf.Mock(&conf.Unified{SiteConfiguration: schema.SiteConfiguration{ExternalURL: "https://test.io"}})
	t.Cleanup(func() {
		conf.Mock(nil)
	})

	job, err := transformRecord(index, handler.ResourceMetadata{
		NumCPUs: 3,
		Memory:  "3T", // big honker
	}, "hunter2")
	if err != nil {
		t.Fatalf("unexpected error transforming record: %s", err)
	}

	expected := apiclient.Job{
		ID:                  42,
		Commit:              "deadbeef",
		RepositoryName:      "linux",
		ShallowClone:        true,
		FetchTags:           false,
		VirtualMachineFiles: nil,
		DockerSteps: []apiclient.DockerStep{
			{
				Key:      "pre-index.0",
				Image:    "alpine",
				Commands: []string{"env", "CUSTOM_MAX_MEM=3072", "yarn", "install"},
				Dir:      "web",
			},
			{
				Key:      "indexer",
				Image:    "lsif-node",
				Commands: []string{"env MAX_DISK=20480 'DEBUG=20.0 GB' index"},
				Dir:      "web",
			},
			{
				Key:   "upload",
				Image: fmt.Sprintf("sourcegraph/src-cli:%s", srccli.MinimumVersion),
				Commands: []string{
					strings.Join(
						[]string{
							"src",
							"lsif", "upload",
							"-no-progress",
							"-repo", "linux",
							"-commit", "deadbeef",
							"-root", "web",
							"-upload-route", "/.executors/lsif/upload",
							"-file", "dump.lsif",
							"-associated-index-id", "42",
						},
						" ",
					),
				},
				Dir: "web",
				Env: []string{
					"SRC_ENDPOINT=https://test.io",
					"SRC_HEADER_AUTHORIZATION=token-executor hunter2",
				},
			},
		},
		RedactedValues: map[string]string{
			"hunter2":                "PASSWORD_REMOVED",
			"token-executor hunter2": "token-executor REDACTED",
		},
	}
	if diff := cmp.Diff(expected, job); diff != "" {
		t.Errorf("unexpected job (-want +got):\n%s", diff)
	}
}

func TestTransformRecordWithResourceReplacementUnbounded(t *testing.T) {
	index := types.Index{
		ID:             42,
		Commit:         "deadbeef",
		RepositoryName: "linux",
		DockerSteps:    nil,
		Root:           "web",
		Indexer:        "lsif-node",
		IndexerArgs:    []string{"bash", "-c", `[ "$VM_DISK_MB" = "" ] && index`},
		Outfile:        "",
	}
	conf.Mock(&conf.Unified{SiteConfiguration: schema.SiteConfiguration{ExternalURL: "https://test.io"}})
	t.Cleanup(func() {
		conf.Mock(nil)
	})

	job, err := transformRecord(index, handler.ResourceMetadata{
		NumCPUs:   3,
		DiskSpace: "0 KB",
	}, "hunter2")
	if err != nil {
		t.Fatalf("unexpected error transforming record: %s", err)
	}

	expected := apiclient.Job{
		ID:                  42,
		Commit:              "deadbeef",
		RepositoryName:      "linux",
		ShallowClone:        true,
		FetchTags:           false,
		VirtualMachineFiles: nil,
		DockerSteps: []apiclient.DockerStep{
			{
				Key:      "indexer",
				Image:    "lsif-node",
				Commands: []string{`bash -c '[ "" = "" ] && index'`},
				Dir:      "web",
			},
			{
				Key:   "upload",
				Image: fmt.Sprintf("sourcegraph/src-cli:%s", srccli.MinimumVersion),
				Commands: []string{
					strings.Join(
						[]string{
							"src",
							"lsif", "upload",
							"-no-progress",
							"-repo", "linux",
							"-commit", "deadbeef",
							"-root", "web",
							"-upload-route", "/.executors/lsif/upload",
							"-file", "dump.lsif",
							"-associated-index-id", "42",
						},
						" ",
					),
				},
				Dir: "web",
				Env: []string{
					"SRC_ENDPOINT=https://test.io",
					"SRC_HEADER_AUTHORIZATION=token-executor hunter2",
				},
			},
		},
		RedactedValues: map[string]string{
			"hunter2":                "PASSWORD_REMOVED",
			"token-executor hunter2": "token-executor REDACTED",
		},
	}
	if diff := cmp.Diff(expected, job); diff != "" {
		t.Errorf("unexpected job (-want +got):\n%s", diff)
	}
}
