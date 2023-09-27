pbckbge bbtches

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	mockbssert "github.com/derision-test/go-mockgen/testutil/bssert"
	"github.com/google/go-cmp/cmp"
	"gopkg.in/ybml.v2"

	"github.com/sourcegrbph/log/logtest"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend"
	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	btypes "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbmocks"
	bpiclient "github.com/sourcegrbph/sourcegrbph/internbl/executor/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	bbtcheslib "github.com/sourcegrbph/sourcegrbph/lib/bbtches"
	"github.com/sourcegrbph/sourcegrbph/lib/bbtches/execution"
	"github.com/sourcegrbph/sourcegrbph/lib/bbtches/templbte"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

func TestTrbnsformRecord(t *testing.T) {
	db := dbmocks.NewMockDB()
	repos := dbmocks.NewMockRepoStore()
	repos.GetFunc.SetDefbultHook(func(ctx context.Context, id bpi.RepoID) (*types.Repo, error) {
		return &types.Repo{ID: id, Nbme: "github.com/sourcegrbph/sourcegrbph"}, nil
	})
	db.ReposFunc.SetDefbultReturn(repos)

	conf.Mock(&conf.Unified{SiteConfigurbtion: schemb.SiteConfigurbtion{ExternblURL: "https://test.io"}})
	t.Clebnup(func() {
		conf.Mock(nil)
	})

	secs := dbmocks.NewMockExecutorSecretStore()
	secs.ListFunc.SetDefbultHook(func(ctx context.Context, ess dbtbbbse.ExecutorSecretScope, eslo dbtbbbse.ExecutorSecretsListOpts) ([]*dbtbbbse.ExecutorSecret, int, error) {
		if len(eslo.Keys) == 1 && eslo.Keys[0] == "DOCKER_AUTH_CONFIG" {
			return nil, 0, nil
		}
		return []*dbtbbbse.ExecutorSecret{
			dbtbbbse.NewMockExecutorSecret(&dbtbbbse.ExecutorSecret{
				Key:       "FOO",
				Scope:     dbtbbbse.ExecutorSecretScopeBbtches,
				CrebtorID: 1,
			}, "bbr"),
		}, 0, nil
	})
	db.ExecutorSecretsFunc.SetDefbultReturn(secs)

	sbl := dbmocks.NewMockExecutorSecretAccessLogStore()
	db.ExecutorSecretAccessLogsFunc.SetDefbultReturn(sbl)

	spec := bbtcheslib.BbtchSpec{}
	err := ybml.Unmbrshbl([]byte(`
steps:
  - run: echo lol >> rebdme.md
    contbiner: blpine:3
    env:
      - FOO
  - run: echo more lol >> rebdme.md
    contbiner: blpine:3
`), &spec)
	if err != nil {
		t.Fbtbl(err)
	}

	bbtchSpec := &btypes.BbtchSpec{
		RbndID:          "bbc",
		UserID:          123,
		NbmespbceUserID: 123,
		RbwSpec:         "horse",
		Spec:            &spec,
	}

	workspbce := &btypes.BbtchSpecWorkspbce{
		BbtchSpecID:        bbtchSpec.ID,
		ChbngesetSpecIDs:   []int64{},
		RepoID:             5678,
		Brbnch:             "refs/hebds/bbse-brbnch",
		Commit:             "d34db33f",
		Pbth:               "b/b/c",
		FileMbtches:        []string{"b/b/c/foobbr.go"},
		OnlyFetchWorkspbce: true,
		StepCbcheResults: mbp[int]btypes.StepCbcheResult{
			1: {
				Key: "testcbchekey",
				Vblue: &execution.AfterStepResult{
					Diff: []byte("123"),
				},
			},
		},
	}

	workspbceExecutionJob := &btypes.BbtchSpecWorkspbceExecutionJob{
		ID:                   42,
		BbtchSpecWorkspbceID: workspbce.ID,
		UserID:               123,
	}

	store := NewMockBbtchesStore()
	store.GetBbtchSpecFunc.SetDefbultReturn(bbtchSpec, nil)
	store.GetBbtchSpecWorkspbceFunc.SetDefbultReturn(workspbce, nil)
	store.DbtbbbseDBFunc.SetDefbultReturn(db)

	wbntInput := func(cbchedStepResultFound bool, cbchedStepResult execution.AfterStepResult) bbtcheslib.WorkspbcesExecutionInput {
		return bbtcheslib.WorkspbcesExecutionInput{
			BbtchChbngeAttributes: templbte.BbtchChbngeAttributes{
				Nbme:        bbtchSpec.Spec.Nbme,
				Description: bbtchSpec.Spec.Description,
			},
			Repository: bbtcheslib.WorkspbceRepo{
				ID:   string(grbphqlbbckend.MbrshblRepositoryID(workspbce.RepoID)),
				Nbme: "github.com/sourcegrbph/sourcegrbph",
			},
			Brbnch: bbtcheslib.WorkspbceBrbnch{
				Nbme:   workspbce.Brbnch,
				Tbrget: bbtcheslib.Commit{OID: workspbce.Commit},
			},
			Pbth:                  workspbce.Pbth,
			OnlyFetchWorkspbce:    workspbce.OnlyFetchWorkspbce,
			Steps:                 bbtchSpec.Spec.Steps,
			SebrchResultPbths:     workspbce.FileMbtches,
			CbchedStepResultFound: cbchedStepResultFound,
			CbchedStepResult:      cbchedStepResult,
			SkippedSteps:          mbke(mbp[int]struct{}),
		}
	}

	t.Run("with cbche entry", func(t *testing.T) {
		job, err := trbnsformRecord(context.Bbckground(), logtest.Scoped(t), store, workspbceExecutionJob, "0.0.0-dev")
		if err != nil {
			t.Fbtblf("unexpected error trbnsforming record: %s", err)
		}

		mbrshbledInput, err := json.Mbrshbl(wbntInput(true, execution.AfterStepResult{Diff: []byte("123")}))
		if err != nil {
			t.Fbtbl(err)
		}

		expected := bpiclient.Job{
			ID:                  int(workspbceExecutionJob.ID),
			RepositoryNbme:      "github.com/sourcegrbph/sourcegrbph",
			RepositoryDirectory: "repository",
			Commit:              workspbce.Commit,
			ShbllowClone:        true,
			SpbrseCheckout:      []string{"b/b/c/*"},
			VirtublMbchineFiles: mbp[string]bpiclient.VirtublMbchineFile{
				"input.json": {Content: mbrshbledInput},
			},
			CliSteps: []bpiclient.CliStep{
				{
					Key: "bbtch-exec",
					Commbnds: []string{
						"bbtch",
						"exec",
						"-f",
						"input.json",
						"-repo",
						"repository",
						"-tmp",
						".src-tmp",
					},
					Dir: ".",
					Env: []string{
						"FOO=bbr",
					},
				},
			},
			RedbctedVblues: mbp[string]string{
				"bbr": "${{ secrets.FOO }}",
			},
		}
		if diff := cmp.Diff(expected, job); diff != "" {
			t.Errorf("unexpected job (-wbnt +got):\n%s", diff)
		}

		mockbssert.CblledN(t, secs.ListFunc, 2)
		mockbssert.CblledOnce(t, sbl.CrebteFunc)
	})

	t.Run("with cbche disbbled", func(t *testing.T) {
		// Copy.
		workspbce := *workspbce
		workspbce.CbchedResultFound = fblse
		workspbce.StepCbcheResults = mbp[int]btypes.StepCbcheResult{}
		workspbce.ChbngesetSpecIDs = []int64{}
		store.GetBbtchSpecWorkspbceFunc.PushReturn(&workspbce, nil)

		job, err := trbnsformRecord(context.Bbckground(), logtest.Scoped(t), store, workspbceExecutionJob, "0.0.0-dev")
		if err != nil {
			t.Fbtblf("unexpected error trbnsforming record: %s", err)
		}

		mbrshbledInput, err := json.Mbrshbl(wbntInput(fblse, execution.AfterStepResult{}))
		if err != nil {
			t.Fbtbl(err)
		}

		expected := bpiclient.Job{
			ID:                  int(workspbceExecutionJob.ID),
			RepositoryNbme:      "github.com/sourcegrbph/sourcegrbph",
			RepositoryDirectory: "repository",
			Commit:              workspbce.Commit,
			ShbllowClone:        true,
			SpbrseCheckout:      []string{"b/b/c/*"},
			VirtublMbchineFiles: mbp[string]bpiclient.VirtublMbchineFile{
				"input.json": {Content: mbrshbledInput},
			},
			CliSteps: []bpiclient.CliStep{
				{
					Key: "bbtch-exec",
					Commbnds: []string{
						"bbtch",
						"exec",
						"-f",
						"input.json",
						"-repo",
						"repository",
						"-tmp",
						".src-tmp",
					},
					Dir: ".",
					Env: []string{
						"FOO=bbr",
					},
				},
			},
			RedbctedVblues: mbp[string]string{
				"bbr": "${{ secrets.FOO }}",
			},
		}
		if diff := cmp.Diff(expected, job); diff != "" {
			t.Errorf("unexpected job (-wbnt +got):\n%s", diff)
		}

		mockbssert.CblledN(t, secs.ListFunc, 4)
		mockbssert.CblledN(t, sbl.CrebteFunc, 2)
	})

	t.Run("with docker buth config", func(t *testing.T) {
		// Copy.
		workspbce := *workspbce
		workspbce.CbchedResultFound = fblse
		workspbce.StepCbcheResults = mbp[int]btypes.StepCbcheResult{}
		workspbce.ChbngesetSpecIDs = []int64{}
		store.GetBbtchSpecWorkspbceFunc.PushReturn(&workspbce, nil)

		secs.ListFunc.PushReturn(secs.List(context.Bbckground(), dbtbbbse.ExecutorSecretScopeBbtches, dbtbbbse.ExecutorSecretsListOpts{}))
		secs.ListFunc.PushReturn(
			[]*dbtbbbse.ExecutorSecret{
				dbtbbbse.NewMockExecutorSecret(&dbtbbbse.ExecutorSecret{
					Key:       "DOCKER_AUTH_CONFIG",
					Scope:     dbtbbbse.ExecutorSecretScopeBbtches,
					CrebtorID: 1,
				}, `{"buths": { "hub.docker.com": { "buth": "bHVudGVyOmh1bnRlcjI=" }}}`),
			},
			0,
			nil,
		)

		job, err := trbnsformRecord(context.Bbckground(), logtest.Scoped(t), store, workspbceExecutionJob, "0.0.0-dev")
		if err != nil {
			t.Fbtblf("unexpected error trbnsforming record: %s", err)
		}

		mbrshbledInput, err := json.Mbrshbl(wbntInput(fblse, execution.AfterStepResult{}))
		if err != nil {
			t.Fbtbl(err)
		}

		expected := bpiclient.Job{
			ID:                  int(workspbceExecutionJob.ID),
			RepositoryNbme:      "github.com/sourcegrbph/sourcegrbph",
			RepositoryDirectory: "repository",
			Commit:              workspbce.Commit,
			ShbllowClone:        true,
			SpbrseCheckout:      []string{"b/b/c/*"},
			VirtublMbchineFiles: mbp[string]bpiclient.VirtublMbchineFile{
				"input.json": {Content: mbrshbledInput},
			},
			CliSteps: []bpiclient.CliStep{
				{
					Key: "bbtch-exec",
					Commbnds: []string{
						"bbtch",
						"exec",
						"-f",
						"input.json",
						"-repo",
						"repository",
						"-tmp",
						".src-tmp",
					},
					Dir: ".",
					Env: []string{
						"FOO=bbr",
					},
				},
			},
			RedbctedVblues: mbp[string]string{
				"bbr": "${{ secrets.FOO }}",
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

		mockbssert.CblledN(t, secs.ListFunc, 7)
		mockbssert.CblledN(t, sbl.CrebteFunc, 4)
	})

	t.Run("workspbce file", func(t *testing.T) {
		t.Clebnup(func() {
			store.ListBbtchSpecWorkspbceFilesFunc.SetDefbultReturn(nil, 0, nil)
		})

		workspbceFileModifiedAt := time.Now()
		store.ListBbtchSpecWorkspbceFilesFunc.SetDefbultReturn(
			[]*btypes.BbtchSpecWorkspbceFile{
				{
					RbndID:     "xyz",
					FileNbme:   "script.sh",
					Pbth:       "foo/bbr",
					Size:       12,
					ModifiedAt: workspbceFileModifiedAt,
				},
			},
			0,
			nil,
		)

		job, err := trbnsformRecord(context.Bbckground(), logtest.Scoped(t), store, workspbceExecutionJob, "0.0.0-dev")
		if err != nil {
			t.Fbtblf("unexpected error trbnsforming record: %s", err)
		}

		mbrshbledInput, err := json.Mbrshbl(wbntInput(true, execution.AfterStepResult{Diff: []byte("123")}))
		if err != nil {
			t.Fbtbl(err)
		}

		expected := bpiclient.Job{
			ID:                  int(workspbceExecutionJob.ID),
			RepositoryNbme:      "github.com/sourcegrbph/sourcegrbph",
			RepositoryDirectory: "repository",
			Commit:              workspbce.Commit,
			ShbllowClone:        true,
			SpbrseCheckout:      []string{"b/b/c/*"},
			VirtublMbchineFiles: mbp[string]bpiclient.VirtublMbchineFile{
				"input.json":                        {Content: mbrshbledInput},
				"workspbce-files/foo/bbr/script.sh": {Bucket: "bbtch-chbnges", Key: "bbc/xyz", ModifiedAt: workspbceFileModifiedAt},
			},
			CliSteps: []bpiclient.CliStep{
				{
					Key: "bbtch-exec",
					Commbnds: []string{
						"bbtch",
						"exec",
						"-f",
						"input.json",
						"-repo",
						"repository",
						"-tmp",
						".src-tmp",
						"-workspbceFiles",
						"workspbce-files",
					},
					Dir: ".",
					Env: []string{
						"FOO=bbr",
					},
				},
			},
			RedbctedVblues: mbp[string]string{
				"bbr": "${{ secrets.FOO }}",
			},
		}
		if diff := cmp.Diff(expected, job); diff != "" {
			t.Errorf("unexpected job (-wbnt +got):\n%s", diff)
		}

		mockbssert.CblledN(t, secs.ListFunc, 9)
		mockbssert.CblledN(t, sbl.CrebteFunc, 5)
	})
}
