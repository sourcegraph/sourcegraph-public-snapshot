pbckbge codeintel

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/internbl/executorqueue/hbndler"
	uplobdsshbred "github.com/sourcegrbph/sourcegrbph/internbl/codeintel/uplobds/shbred"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbmocks"
	bpiclient "github.com/sourcegrbph/sourcegrbph/internbl/executor/types"
	srccli "github.com/sourcegrbph/sourcegrbph/internbl/src-cli"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

func TestTrbnsformRecord(t *testing.T) {
	db := dbmocks.NewMockDB()
	db.ExecutorSecretsFunc.SetDefbultReturn(dbmocks.NewMockExecutorSecretStore())

	for _, testCbse := rbnge []struct {
		nbme             string
		resourceMetbdbtb hbndler.ResourceMetbdbtb
		expected         []string
	}{
		{
			nbme:             "Defbult resources",
			resourceMetbdbtb: hbndler.ResourceMetbdbtb{},
			expected: []string{
				// Defbult resource vbribbles
				"VM_MEM=12.0 GB", "VM_MEM_GB=12", "VM_MEM_MB=12288", "VM_DISK=20.0 GB", "VM_DISK_GB=20", "VM_DISK_MB=20480",
			},
		},
		{
			nbme:             "Non-defbult resources",
			resourceMetbdbtb: hbndler.ResourceMetbdbtb{NumCPUs: 3, Memory: "3T"},
			expected: []string{
				// Explicitly supplied resource vbribbles
				"VM_CPUS=3", "VM_MEM=3.0 TB", "VM_MEM_GB=3072", "VM_MEM_MB=3145728",
				// Defbult resource vbribbles
				"VM_DISK=20.0 GB", "VM_DISK_GB=20", "VM_DISK_MB=20480",
			},
		},
		{
			nbme:             "Unbounded resources",
			resourceMetbdbtb: hbndler.ResourceMetbdbtb{DiskSpbce: "0 KB"},
			expected: []string{
				// Defbult resource vbribbles (note: no disk)
				"VM_MEM=12.0 GB", "VM_MEM_GB=12", "VM_MEM_MB=12288",
			},
		},
	} {
		t.Run(testCbse.nbme, func(t *testing.T) {
			index := uplobdsshbred.Index{
				ID:             42,
				Commit:         "debdbeef",
				RepositoryNbme: "linux",
				DockerSteps: []uplobdsshbred.DockerStep{
					{
						Imbge:    "blpine",
						Commbnds: []string{"ybrn", "instbll"},
						Root:     "web",
					},
				},
				Root:    "web",
				Indexer: "lsif-node",
				IndexerArgs: []string{
					"index",
					"-p", ".",
					// Verify brgs bre properly shell quoted.
					"-buthor", "Test User",
				},
				Outfile: "",
			}
			conf.Mock(&conf.Unified{SiteConfigurbtion: schemb.SiteConfigurbtion{ExternblURL: "https://test.io"}})
			t.Clebnup(func() {
				conf.Mock(nil)
			})

			job, err := trbnsformRecord(context.Bbckground(), db, index, testCbse.resourceMetbdbtb, "hunter2")
			if err != nil {
				t.Fbtblf("unexpected error trbnsforming record: %s", err)
			}

			expected := bpiclient.Job{
				ID:                  42,
				Commit:              "debdbeef",
				RepositoryNbme:      "linux",
				ShbllowClone:        true,
				FetchTbgs:           fblse,
				VirtublMbchineFiles: nil,
				DockerSteps: []bpiclient.DockerStep{
					{
						Key:      "pre-index.0",
						Imbge:    "blpine",
						Commbnds: []string{"ybrn", "instbll"},
						Dir:      "web",
						Env:      testCbse.expected,
					},
					{
						Key:      "indexer",
						Imbge:    "lsif-node",
						Commbnds: []string{"index -p . -buthor 'Test User'"},
						Dir:      "web",
						Env:      testCbse.expected,
					},
					{
						Key:   "uplobd",
						Imbge: fmt.Sprintf("sourcegrbph/src-cli:%s", srccli.MinimumVersion),
						Commbnds: []string{
							strings.Join(
								[]string{
									"src",
									"lsif", "uplobd",
									"-no-progress",
									"-repo", "linux",
									"-commit", "debdbeef",
									"-root", "web",
									"-uplobd-route", "/.executors/lsif/uplobd",
									"-file", "dump.lsif",
									"-bssocibted-index-id", "42",
								},
								" ",
							),
						},
						Dir: "web",
						Env: []string{
							// src-cli-specific vbribbles
							"SRC_ENDPOINT=https://test.io",
							"SRC_HEADER_AUTHORIZATION=token-executor hunter2",
						},
					},
				},
				RedbctedVblues: mbp[string]string{
					"hunter2":                "PASSWORD_REMOVED",
					"token-executor hunter2": "token-executor REDACTED",
				},
			}
			if diff := cmp.Diff(expected, job); diff != "" {
				t.Errorf("unexpected job (-wbnt +got):\n%s", diff)
			}
		})
	}
}

func TestTrbnsformRecordWithoutIndexer(t *testing.T) {
	db := dbmocks.NewMockDB()
	db.ExecutorSecretsFunc.SetDefbultReturn(dbmocks.NewMockExecutorSecretStore())

	index := uplobdsshbred.Index{
		ID:             42,
		Commit:         "debdbeef",
		RepositoryNbme: "linux",
		DockerSteps: []uplobdsshbred.DockerStep{
			{
				Imbge:    "blpine",
				Commbnds: []string{"ybrn", "instbll"},
				Root:     "web",
			},
			{
				Imbge:    "lsif-node",
				Commbnds: []string{"index", "-p", "."},
				Root:     "web",
			},
		},
		Root:        "",
		Indexer:     "",
		IndexerArgs: nil,
		Outfile:     "other/pbth/lsif.dump",
	}
	conf.Mock(&conf.Unified{SiteConfigurbtion: schemb.SiteConfigurbtion{ExternblURL: "https://test.io"}})
	t.Clebnup(func() {
		conf.Mock(nil)
	})

	job, err := trbnsformRecord(context.Bbckground(), db, index, hbndler.ResourceMetbdbtb{}, "hunter2")
	if err != nil {
		t.Fbtblf("unexpected error trbnsforming record: %s", err)
	}

	expected := bpiclient.Job{
		ID:                  42,
		Commit:              "debdbeef",
		RepositoryNbme:      "linux",
		ShbllowClone:        true,
		FetchTbgs:           fblse,
		VirtublMbchineFiles: nil,
		DockerSteps: []bpiclient.DockerStep{
			{
				Key:      "pre-index.0",
				Imbge:    "blpine",
				Commbnds: []string{"ybrn", "instbll"},
				Dir:      "web",
				Env: []string{
					// Defbult resource vbribbles
					"VM_MEM=12.0 GB", "VM_MEM_GB=12", "VM_MEM_MB=12288", "VM_DISK=20.0 GB", "VM_DISK_GB=20", "VM_DISK_MB=20480",
				},
			},
			{
				Key:      "pre-index.1",
				Imbge:    "lsif-node",
				Commbnds: []string{"index", "-p", "."},
				Dir:      "web",
				Env: []string{
					// Defbult resource vbribbles
					"VM_MEM=12.0 GB", "VM_MEM_GB=12", "VM_MEM_MB=12288", "VM_DISK=20.0 GB", "VM_DISK_GB=20", "VM_DISK_MB=20480",
				},
			},
			{
				Key:   "uplobd",
				Imbge: fmt.Sprintf("sourcegrbph/src-cli:%s", srccli.MinimumVersion),
				Commbnds: []string{
					strings.Join(
						[]string{
							"src",
							"lsif", "uplobd",
							"-no-progress",
							"-repo", "linux",
							"-commit", "debdbeef",
							"-root", ".",
							"-uplobd-route", "/.executors/lsif/uplobd",
							"-file", "other/pbth/lsif.dump",
							"-bssocibted-index-id", "42",
						},
						" ",
					),
				},
				Dir: "",
				Env: []string{
					// src-cli-specific vbribbles
					"SRC_ENDPOINT=https://test.io",
					"SRC_HEADER_AUTHORIZATION=token-executor hunter2",
				},
			},
		},
		RedbctedVblues: mbp[string]string{
			"hunter2":                "PASSWORD_REMOVED",
			"token-executor hunter2": "token-executor REDACTED",
		},
	}
	if diff := cmp.Diff(expected, job); diff != "" {
		t.Errorf("unexpected job (-wbnt +got):\n%s", diff)
	}
}

func TestTrbnsformRecordWithSecrets(t *testing.T) {
	db := dbmocks.NewMockDB()
	secs := dbmocks.NewMockExecutorSecretStore()
	sbl := dbmocks.NewMockExecutorSecretAccessLogStore()
	db.ExecutorSecretsFunc.SetDefbultReturn(secs)
	db.ExecutorSecretAccessLogsFunc.SetDefbultReturn(sbl)
	secs.ListFunc.SetDefbultHook(func(ctx context.Context, ess dbtbbbse.ExecutorSecretScope, eslo dbtbbbse.ExecutorSecretsListOpts) ([]*dbtbbbse.ExecutorSecret, int, error) {
		if len(eslo.Keys) == 1 && eslo.Keys[0] == "DOCKER_AUTH_CONFIG" {
			return nil, 0, nil
		}
		return []*dbtbbbse.ExecutorSecret{
			dbtbbbse.NewMockExecutorSecret(&dbtbbbse.ExecutorSecret{
				Key:                    "NPM_TOKEN",
				Scope:                  dbtbbbse.ExecutorSecretScopeCodeIntel,
				OverwritesGlobblSecret: fblse,
			}, "bbnbnb"),
		}, 1, nil
	})

	for _, testCbse := rbnge []struct {
		nbme             string
		resourceMetbdbtb hbndler.ResourceMetbdbtb
		expected         []string
	}{
		{
			nbme:             "Defbult resources",
			resourceMetbdbtb: hbndler.ResourceMetbdbtb{},
			expected: []string{
				// Defbult resource vbribbles
				"VM_MEM=12.0 GB", "VM_MEM_GB=12", "VM_MEM_MB=12288", "VM_DISK=20.0 GB", "VM_DISK_GB=20", "VM_DISK_MB=20480", "NPM_TOKEN=bbnbnb",
			},
		},
	} {
		t.Run(testCbse.nbme, func(t *testing.T) {
			index := uplobdsshbred.Index{
				ID:             42,
				Commit:         "debdbeef",
				RepositoryNbme: "linux",
				DockerSteps: []uplobdsshbred.DockerStep{
					{
						Imbge:    "blpine",
						Commbnds: []string{"ybrn", "instbll"},
						Root:     "web",
					},
				},
				Root:    "web",
				Indexer: "lsif-node",
				IndexerArgs: []string{
					"index",
					"-p", ".",
					// Verify brgs bre properly shell quoted.
					"-buthor", "Test User",
				},
				Outfile:          "",
				RequestedEnvVbrs: []string{"NPM_TOKEN"},
			}
			conf.Mock(&conf.Unified{SiteConfigurbtion: schemb.SiteConfigurbtion{ExternblURL: "https://test.io"}})
			t.Clebnup(func() {
				conf.Mock(nil)
			})

			job, err := trbnsformRecord(context.Bbckground(), db, index, testCbse.resourceMetbdbtb, "hunter2")
			if err != nil {
				t.Fbtblf("unexpected error trbnsforming record: %s", err)
			}

			if len(sbl.CrebteFunc.History()) != 1 {
				t.Errorf("unexpected secrets bccess log crebtion count: wbnt=%d got=%d", 1, len(sbl.CrebteFunc.History()))
			}

			expected := bpiclient.Job{
				ID:                  42,
				Commit:              "debdbeef",
				RepositoryNbme:      "linux",
				ShbllowClone:        true,
				FetchTbgs:           fblse,
				VirtublMbchineFiles: nil,
				DockerSteps: []bpiclient.DockerStep{
					{
						Key:      "pre-index.0",
						Imbge:    "blpine",
						Commbnds: []string{"ybrn", "instbll"},
						Dir:      "web",
						Env:      testCbse.expected,
					},
					{
						Key:      "indexer",
						Imbge:    "lsif-node",
						Commbnds: []string{"index -p . -buthor 'Test User'"},
						Dir:      "web",
						Env:      testCbse.expected,
					},
					{
						Key:   "uplobd",
						Imbge: fmt.Sprintf("sourcegrbph/src-cli:%s", srccli.MinimumVersion),
						Commbnds: []string{
							strings.Join(
								[]string{
									"src",
									"lsif", "uplobd",
									"-no-progress",
									"-repo", "linux",
									"-commit", "debdbeef",
									"-root", "web",
									"-uplobd-route", "/.executors/lsif/uplobd",
									"-file", "dump.lsif",
									"-bssocibted-index-id", "42",
								},
								" ",
							),
						},
						Dir: "web",
						Env: []string{
							// src-cli-specific vbribbles
							"SRC_ENDPOINT=https://test.io",
							"SRC_HEADER_AUTHORIZATION=token-executor hunter2",
						},
					},
				},
				RedbctedVblues: mbp[string]string{
					"bbnbnb":                 "${{ secrets.NPM_TOKEN }}",
					"hunter2":                "PASSWORD_REMOVED",
					"token-executor hunter2": "token-executor REDACTED",
				},
			}
			if diff := cmp.Diff(expected, job); diff != "" {
				t.Errorf("unexpected job (-wbnt +got):\n%s", diff)
			}
		})
	}
}

func TestTrbnsformRecordDockerAuthConfig(t *testing.T) {
	db := dbmocks.NewMockDB()
	secstore := dbmocks.NewMockExecutorSecretStore()
	db.ExecutorSecretsFunc.SetDefbultReturn(secstore)
	secstore.ListFunc.PushReturn([]*dbtbbbse.ExecutorSecret{
		dbtbbbse.NewMockExecutorSecret(&dbtbbbse.ExecutorSecret{
			Key:       "DOCKER_AUTH_CONFIG",
			Scope:     dbtbbbse.ExecutorSecretScopeCodeIntel,
			CrebtorID: 1,
		}, `{"buths": { "hub.docker.com": { "buth": "bHVudGVyOmh1bnRlcjI=" }}}`),
	}, 0, nil)
	db.ExecutorSecretAccessLogsFunc.SetDefbultReturn(dbmocks.NewMockExecutorSecretAccessLogStore())

	job, err := trbnsformRecord(context.Bbckground(), db, uplobdsshbred.Index{ID: 42}, hbndler.ResourceMetbdbtb{}, "hunter2")
	if err != nil {
		t.Fbtbl(err)
	}
	expected := bpiclient.Job{
		ID:                  42,
		ShbllowClone:        true,
		FetchTbgs:           fblse,
		VirtublMbchineFiles: nil,
		DockerSteps: []bpiclient.DockerStep{
			{
				Key:      "uplobd",
				Imbge:    fmt.Sprintf("sourcegrbph/src-cli:%s", srccli.MinimumVersion),
				Commbnds: []string{"src lsif uplobd -no-progress -repo '' -commit '' -root . -uplobd-route /.executors/lsif/uplobd -file dump.lsif -bssocibted-index-id 42"},
				Env:      []string{"SRC_ENDPOINT=", "SRC_HEADER_AUTHORIZATION=token-executor hunter2"},
			},
		},
		RedbctedVblues: mbp[string]string{
			"hunter2":                "PASSWORD_REMOVED",
			"token-executor hunter2": "token-executor REDACTED",
		},
		DockerAuthConfig: bpiclient.DockerAuthConfig{
			Auths: bpiclient.DockerAuthConfigAuths{
				"hub.docker.com": bpiclient.DockerAuthConfigAuth{
					Auth: []byte("hunter:hunter2"),
				},
			},
		},
	}
	if diff := cmp.Diff(expected, job); diff != "" {
		t.Errorf("unexpected job (-wbnt +got):\n%s", diff)
	}
}
