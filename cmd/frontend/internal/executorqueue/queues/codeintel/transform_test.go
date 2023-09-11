package codeintel

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/executorqueue/handler"
	uploadsshared "github.com/sourcegraph/sourcegraph/internal/codeintel/uploads/shared"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbmocks"
	apiclient "github.com/sourcegraph/sourcegraph/internal/executor/types"
	srccli "github.com/sourcegraph/sourcegraph/internal/src-cli"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestTransformRecord(t *testing.T) {
	db := dbmocks.NewMockDB()
	db.ExecutorSecretsFunc.SetDefaultReturn(dbmocks.NewMockExecutorSecretStore())

	for _, testCase := range []struct {
		name             string
		resourceMetadata handler.ResourceMetadata
		expected         []string
	}{
		{
			name:             "Default resources",
			resourceMetadata: handler.ResourceMetadata{},
			expected: []string{
				// Default resource variables
				"VM_MEM=12.0 GB", "VM_MEM_GB=12", "VM_MEM_MB=12288", "VM_DISK=20.0 GB", "VM_DISK_GB=20", "VM_DISK_MB=20480",
			},
		},
		{
			name:             "Non-default resources",
			resourceMetadata: handler.ResourceMetadata{NumCPUs: 3, Memory: "3T"},
			expected: []string{
				// Explicitly supplied resource variables
				"VM_CPUS=3", "VM_MEM=3.0 TB", "VM_MEM_GB=3072", "VM_MEM_MB=3145728",
				// Default resource variables
				"VM_DISK=20.0 GB", "VM_DISK_GB=20", "VM_DISK_MB=20480",
			},
		},
		{
			name:             "Unbounded resources",
			resourceMetadata: handler.ResourceMetadata{DiskSpace: "0 KB"},
			expected: []string{
				// Default resource variables (note: no disk)
				"VM_MEM=12.0 GB", "VM_MEM_GB=12", "VM_MEM_MB=12288",
			},
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			index := uploadsshared.Index{
				ID:             42,
				Commit:         "deadbeef",
				RepositoryName: "linux",
				DockerSteps: []uploadsshared.DockerStep{
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

			job, err := transformRecord(context.Background(), db, index, testCase.resourceMetadata, "hunter2")
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
						Env:      testCase.expected,
					},
					{
						Key:      "indexer",
						Image:    "lsif-node",
						Commands: []string{"index -p . -author 'Test User'"},
						Dir:      "web",
						Env:      testCase.expected,
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
							// src-cli-specific variables
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
		})
	}
}

func TestTransformRecordWithoutIndexer(t *testing.T) {
	db := dbmocks.NewMockDB()
	db.ExecutorSecretsFunc.SetDefaultReturn(dbmocks.NewMockExecutorSecretStore())

	index := uploadsshared.Index{
		ID:             42,
		Commit:         "deadbeef",
		RepositoryName: "linux",
		DockerSteps: []uploadsshared.DockerStep{
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

	job, err := transformRecord(context.Background(), db, index, handler.ResourceMetadata{}, "hunter2")
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
				Env: []string{
					// Default resource variables
					"VM_MEM=12.0 GB", "VM_MEM_GB=12", "VM_MEM_MB=12288", "VM_DISK=20.0 GB", "VM_DISK_GB=20", "VM_DISK_MB=20480",
				},
			},
			{
				Key:      "pre-index.1",
				Image:    "lsif-node",
				Commands: []string{"index", "-p", "."},
				Dir:      "web",
				Env: []string{
					// Default resource variables
					"VM_MEM=12.0 GB", "VM_MEM_GB=12", "VM_MEM_MB=12288", "VM_DISK=20.0 GB", "VM_DISK_GB=20", "VM_DISK_MB=20480",
				},
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
					// src-cli-specific variables
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

func TestTransformRecordWithSecrets(t *testing.T) {
	db := dbmocks.NewMockDB()
	secs := dbmocks.NewMockExecutorSecretStore()
	sal := dbmocks.NewMockExecutorSecretAccessLogStore()
	db.ExecutorSecretsFunc.SetDefaultReturn(secs)
	db.ExecutorSecretAccessLogsFunc.SetDefaultReturn(sal)
	secs.ListFunc.SetDefaultHook(func(ctx context.Context, ess database.ExecutorSecretScope, eslo database.ExecutorSecretsListOpts) ([]*database.ExecutorSecret, int, error) {
		if len(eslo.Keys) == 1 && eslo.Keys[0] == "DOCKER_AUTH_CONFIG" {
			return nil, 0, nil
		}
		return []*database.ExecutorSecret{
			database.NewMockExecutorSecret(&database.ExecutorSecret{
				Key:                    "NPM_TOKEN",
				Scope:                  database.ExecutorSecretScopeCodeIntel,
				OverwritesGlobalSecret: false,
			}, "banana"),
		}, 1, nil
	})

	for _, testCase := range []struct {
		name             string
		resourceMetadata handler.ResourceMetadata
		expected         []string
	}{
		{
			name:             "Default resources",
			resourceMetadata: handler.ResourceMetadata{},
			expected: []string{
				// Default resource variables
				"VM_MEM=12.0 GB", "VM_MEM_GB=12", "VM_MEM_MB=12288", "VM_DISK=20.0 GB", "VM_DISK_GB=20", "VM_DISK_MB=20480", "NPM_TOKEN=banana",
			},
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			index := uploadsshared.Index{
				ID:             42,
				Commit:         "deadbeef",
				RepositoryName: "linux",
				DockerSteps: []uploadsshared.DockerStep{
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
				Outfile:          "",
				RequestedEnvVars: []string{"NPM_TOKEN"},
			}
			conf.Mock(&conf.Unified{SiteConfiguration: schema.SiteConfiguration{ExternalURL: "https://test.io"}})
			t.Cleanup(func() {
				conf.Mock(nil)
			})

			job, err := transformRecord(context.Background(), db, index, testCase.resourceMetadata, "hunter2")
			if err != nil {
				t.Fatalf("unexpected error transforming record: %s", err)
			}

			if len(sal.CreateFunc.History()) != 1 {
				t.Errorf("unexpected secrets access log creation count: want=%d got=%d", 1, len(sal.CreateFunc.History()))
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
						Env:      testCase.expected,
					},
					{
						Key:      "indexer",
						Image:    "lsif-node",
						Commands: []string{"index -p . -author 'Test User'"},
						Dir:      "web",
						Env:      testCase.expected,
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
							// src-cli-specific variables
							"SRC_ENDPOINT=https://test.io",
							"SRC_HEADER_AUTHORIZATION=token-executor hunter2",
						},
					},
				},
				RedactedValues: map[string]string{
					"banana":                 "${{ secrets.NPM_TOKEN }}",
					"hunter2":                "PASSWORD_REMOVED",
					"token-executor hunter2": "token-executor REDACTED",
				},
			}
			if diff := cmp.Diff(expected, job); diff != "" {
				t.Errorf("unexpected job (-want +got):\n%s", diff)
			}
		})
	}
}

func TestTransformRecordDockerAuthConfig(t *testing.T) {
	db := dbmocks.NewMockDB()
	secstore := dbmocks.NewMockExecutorSecretStore()
	db.ExecutorSecretsFunc.SetDefaultReturn(secstore)
	secstore.ListFunc.PushReturn([]*database.ExecutorSecret{
		database.NewMockExecutorSecret(&database.ExecutorSecret{
			Key:       "DOCKER_AUTH_CONFIG",
			Scope:     database.ExecutorSecretScopeCodeIntel,
			CreatorID: 1,
		}, `{"auths": { "hub.docker.com": { "auth": "aHVudGVyOmh1bnRlcjI=" }}}`),
	}, 0, nil)
	db.ExecutorSecretAccessLogsFunc.SetDefaultReturn(dbmocks.NewMockExecutorSecretAccessLogStore())

	job, err := transformRecord(context.Background(), db, uploadsshared.Index{ID: 42}, handler.ResourceMetadata{}, "hunter2")
	if err != nil {
		t.Fatal(err)
	}
	expected := apiclient.Job{
		ID:                  42,
		ShallowClone:        true,
		FetchTags:           false,
		VirtualMachineFiles: nil,
		DockerSteps: []apiclient.DockerStep{
			{
				Key:      "upload",
				Image:    fmt.Sprintf("sourcegraph/src-cli:%s", srccli.MinimumVersion),
				Commands: []string{"src lsif upload -no-progress -repo '' -commit '' -root . -upload-route /.executors/lsif/upload -file dump.lsif -associated-index-id 42"},
				Env:      []string{"SRC_ENDPOINT=", "SRC_HEADER_AUTHORIZATION=token-executor hunter2"},
			},
		},
		RedactedValues: map[string]string{
			"hunter2":                "PASSWORD_REMOVED",
			"token-executor hunter2": "token-executor REDACTED",
		},
		DockerAuthConfig: apiclient.DockerAuthConfig{
			Auths: apiclient.DockerAuthConfigAuths{
				"hub.docker.com": apiclient.DockerAuthConfigAuth{
					Auth: []byte("hunter:hunter2"),
				},
			},
		},
	}
	if diff := cmp.Diff(expected, job); diff != "" {
		t.Errorf("unexpected job (-want +got):\n%s", diff)
	}
}
