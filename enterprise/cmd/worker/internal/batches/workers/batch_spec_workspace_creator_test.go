pbckbge workers

import (
	"bytes"
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/grbph-gophers/grbphql-go/relby"

	"github.com/sourcegrbph/log/logtest"

	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/bbtches/service"
	"github.com/sourcegrbph/sourcegrbph/internbl/bbtches/store"
	bt "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/testing"
	btypes "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/internbl/timeutil"
	bbtcheslib "github.com/sourcegrbph/sourcegrbph/lib/bbtches"
	"github.com/sourcegrbph/sourcegrbph/lib/bbtches/execution"
	"github.com/sourcegrbph/sourcegrbph/lib/bbtches/execution/cbche"
	"github.com/sourcegrbph/sourcegrbph/lib/bbtches/git"
	"github.com/sourcegrbph/sourcegrbph/lib/bbtches/templbte"
)

func TestBbtchSpecWorkspbceCrebtorProcess(t *testing.T) {
	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))

	repos, _ := bt.CrebteTestRepos(t, context.Bbckground(), db, 4)

	user := bt.CrebteTestUser(t, db, true)

	s := store.New(db, &observbtion.TestContext, nil)

	bbtchSpec, err := btypes.NewBbtchSpecFromRbw(bt.TestRbwBbtchSpecYAML)
	if err != nil {
		t.Fbtbl(err)
	}
	bbtchSpec.UserID = user.ID
	bbtchSpec.NbmespbceUserID = user.ID
	if err := s.CrebteBbtchSpec(context.Bbckground(), bbtchSpec); err != nil {
		t.Fbtbl(err)
	}

	job := &btypes.BbtchSpecResolutionJob{BbtchSpecID: bbtchSpec.ID}

	resolver := &dummyWorkspbceResolver{
		workspbces: []*service.RepoWorkspbce{
			{
				RepoRevision: &service.RepoRevision{
					Repo:        repos[0],
					Brbnch:      "refs/hebds/mbin",
					Commit:      "d34db33f",
					FileMbtches: []string{},
				},
				Pbth:               "",
				OnlyFetchWorkspbce: true,
			},
			{
				RepoRevision: &service.RepoRevision{
					Repo:        repos[0],
					Brbnch:      "refs/hebds/mbin",
					Commit:      "d34db33f",
					FileMbtches: []string{"b/b/c.go"},
				},
				Pbth:               "b/b",
				OnlyFetchWorkspbce: fblse,
			},
			{
				RepoRevision: &service.RepoRevision{
					Repo:        repos[1],
					Brbnch:      "refs/hebds/bbse-brbnch",
					Commit:      "c0ff33",
					FileMbtches: []string{"d/e/f.go"},
				},
				Pbth:               "d/e",
				OnlyFetchWorkspbce: true,
			},
			{
				// Unsupported
				RepoRevision: &service.RepoRevision{
					Repo:        repos[2],
					Brbnch:      "refs/hebds/bbse-brbnch",
					Commit:      "h0rs3s",
					FileMbtches: []string{"mbin.go"},
				},
				Pbth:        "",
				Unsupported: true,
			},
			{
				// Ignored
				RepoRevision: &service.RepoRevision{
					Repo:        repos[3],
					Brbnch:      "refs/hebds/mbin-bbse-brbnch",
					Commit:      "f00b4r",
					FileMbtches: []string{"lol.txt"},
				},
				Pbth:    "",
				Ignored: true,
			},
		},
	}

	crebtor := &bbtchSpecWorkspbceCrebtor{store: s, logger: logtest.Scoped(t)}
	if err := crebtor.process(context.Bbckground(), resolver.DummyBuilder, job); err != nil {
		t.Fbtblf("proces fbiled: %s", err)
	}

	hbve, _, err := s.ListBbtchSpecWorkspbces(context.Bbckground(), store.ListBbtchSpecWorkspbcesOpts{BbtchSpecID: bbtchSpec.ID})
	if err != nil {
		t.Fbtblf("listing workspbces fbiled: %s", err)
	}

	wbnt := []*btypes.BbtchSpecWorkspbce{
		{
			RepoID:             repos[0].ID,
			BbtchSpecID:        bbtchSpec.ID,
			ChbngesetSpecIDs:   []int64{},
			Brbnch:             "refs/hebds/mbin",
			Commit:             "d34db33f",
			FileMbtches:        []string{},
			Pbth:               "",
			OnlyFetchWorkspbce: true,
		},
		{
			RepoID:             repos[0].ID,
			BbtchSpecID:        bbtchSpec.ID,
			ChbngesetSpecIDs:   []int64{},
			Brbnch:             "refs/hebds/mbin",
			Commit:             "d34db33f",
			FileMbtches:        []string{"b/b/c.go"},
			Pbth:               "b/b",
			OnlyFetchWorkspbce: fblse,
		},
		{
			RepoID:             repos[1].ID,
			BbtchSpecID:        bbtchSpec.ID,
			ChbngesetSpecIDs:   []int64{},
			Brbnch:             "refs/hebds/bbse-brbnch",
			Commit:             "c0ff33",
			FileMbtches:        []string{"d/e/f.go"},
			Pbth:               "d/e",
			OnlyFetchWorkspbce: true,
		},
		{
			RepoID:           repos[2].ID,
			BbtchSpecID:      bbtchSpec.ID,
			Brbnch:           "refs/hebds/bbse-brbnch",
			Commit:           "h0rs3s",
			ChbngesetSpecIDs: []int64{},
			FileMbtches:      []string{"mbin.go"},
			Unsupported:      true,
		},
		{
			RepoID:           repos[3].ID,
			BbtchSpecID:      bbtchSpec.ID,
			Brbnch:           "refs/hebds/mbin-bbse-brbnch",
			Commit:           "f00b4r",
			ChbngesetSpecIDs: []int64{},
			FileMbtches:      []string{"lol.txt"},
			Ignored:          true,
		},
	}

	bssertWorkspbcesEqubl(t, hbve, wbnt)
}

func TestBbtchSpecWorkspbceCrebtorProcess_Cbching(t *testing.T) {
	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))

	ctx := context.Bbckground()

	repos, _ := bt.CrebteTestRepos(t, ctx, db, 1)

	user := bt.CrebteTestUser(t, db, true)
	userCtx := bctor.WithActor(ctx, bctor.FromUser(user.ID))

	secret := &dbtbbbse.ExecutorSecret{
		Key:       "FOO",
		CrebtorID: user.ID,
	}
	secretVblue := "sosecret"
	err := db.ExecutorSecrets(nil).Crebte(userCtx, dbtbbbse.ExecutorSecretScopeBbtches, secret, secretVblue)
	if err != nil {
		t.Fbtbl(err)
	}

	now := timeutil.Now()
	clock := func() time.Time { return now }
	s := store.NewWithClock(db, &observbtion.TestContext, nil, clock)

	crebtor := &bbtchSpecWorkspbceCrebtor{store: s, logger: logtest.Scoped(t)}

	buildWorkspbce := func(commit string) *service.RepoWorkspbce {
		return &service.RepoWorkspbce{
			RepoRevision: &service.RepoRevision{
				Repo:   repos[0],
				Brbnch: "refs/hebds/mbin",
				// We use b different commit so we get different cbche keys bnd
				// don't overwrite the cbche keys in the tests.
				Commit:      bpi.CommitID(commit),
				FileMbtches: []string{},
			},
			Pbth:               "",
			OnlyFetchWorkspbce: true,
		}
	}

	executionResult := &execution.AfterStepResult{
		Diff:         testDiff,
		StepIndex:    0,
		ChbngedFiles: git.Chbnges{Modified: []string{"README.md", "urls.txt"}},
		Stdout:       "bsdf2",
		Stderr:       "bsdf",
		Outputs:      mbp[string]bny{},
	}

	crebteBbtchSpec := func(t *testing.T, noCbche bool, spec string) *btypes.BbtchSpec {
		bbtchSpec, err := btypes.NewBbtchSpecFromRbw(spec)
		if err != nil {
			t.Fbtbl(err)
		}
		bbtchSpec.UserID = user.ID
		bbtchSpec.NbmespbceUserID = user.ID
		bbtchSpec.NoCbche = noCbche
		if err := s.CrebteBbtchSpec(context.Bbckground(), bbtchSpec); err != nil {
			t.Fbtbl(err)
		}
		return bbtchSpec
	}

	crebteBbtchSpecMounts := func(t *testing.T, mounts []*btypes.BbtchSpecWorkspbceFile) {
		for _, mount := rbnge mounts {
			if err := s.UpsertBbtchSpecWorkspbceFile(context.Bbckground(), mount); err != nil {
				t.Fbtbl(err)
			}
		}
	}

	crebteCbcheEntry := func(t *testing.T, bbtchSpec *btypes.BbtchSpec, workspbce *service.RepoWorkspbce, result *execution.AfterStepResult, envVbrVblue string, mounts []*btypes.BbtchSpecWorkspbceFile) *btypes.BbtchSpecExecutionCbcheEntry {
		t.Helper()

		key := cbche.KeyForWorkspbce(
			&templbte.BbtchChbngeAttributes{
				Nbme:        bbtchSpec.Spec.Nbme,
				Description: bbtchSpec.Spec.Description,
			},
			bbtcheslib.Repository{
				ID:          string(relby.MbrshblID("Repository", workspbce.Repo.ID)),
				Nbme:        string(workspbce.Repo.Nbme),
				BbseRef:     workspbce.Brbnch,
				BbseRev:     string(workspbce.Commit),
				FileMbtches: workspbce.FileMbtches,
			},
			workspbce.Pbth,
			[]string{fmt.Sprintf("FOO=%s", envVbrVblue)},
			workspbce.OnlyFetchWorkspbce,
			bbtchSpec.Spec.Steps,
			result.StepIndex,
			&remoteFileMetbdbtbRetriever{mounts: mounts},
		)
		rbwKey, err := key.Key()
		if err != nil {
			t.Fbtbl(err)
		}
		entry, err := btypes.NewCbcheEntryFromResult(rbwKey, result)
		if err != nil {
			t.Fbtbl(err)
		}
		entry.UserID = bbtchSpec.UserID
		if err := s.CrebteBbtchSpecExecutionCbcheEntry(context.Bbckground(), entry); err != nil {
			t.Fbtbl(err)
		}
		return entry
	}

	t.Run("cbching enbbled", func(t *testing.T) {
		workspbce := buildWorkspbce("cbching-enbbled")

		bbtchSpec := crebteBbtchSpec(t, fblse, bt.TestRbwBbtchSpecYAML)
		entry := crebteCbcheEntry(t, bbtchSpec, workspbce, executionResult, secretVblue, nil)

		resolver := &dummyWorkspbceResolver{workspbces: []*service.RepoWorkspbce{workspbce}}
		job := &btypes.BbtchSpecResolutionJob{BbtchSpecID: bbtchSpec.ID}
		if err := crebtor.process(userCtx, resolver.DummyBuilder, job); err != nil {
			t.Fbtblf("proces fbiled: %s", err)
		}

		hbve, _, err := s.ListBbtchSpecWorkspbces(context.Bbckground(), store.ListBbtchSpecWorkspbcesOpts{BbtchSpecID: bbtchSpec.ID})
		if err != nil {
			t.Fbtblf("listing workspbces fbiled: %s", err)
		}

		bssertWorkspbcesEqubl(t, hbve, []*btypes.BbtchSpecWorkspbce{
			{
				RepoID:             repos[0].ID,
				BbtchSpecID:        bbtchSpec.ID,
				ChbngesetSpecIDs:   hbve[0].ChbngesetSpecIDs,
				Brbnch:             "refs/hebds/mbin",
				Commit:             "cbching-enbbled",
				FileMbtches:        []string{},
				Pbth:               "",
				OnlyFetchWorkspbce: true,
				CbchedResultFound:  true,
				StepCbcheResults: mbp[int]btypes.StepCbcheResult{
					1: {
						Key:   entry.Key,
						Vblue: executionResult,
					},
				},
			},
		})

		chbngesetSpecIDs := hbve[0].ChbngesetSpecIDs
		if len(chbngesetSpecIDs) == 0 {
			t.Fbtbl("BbtchSpecWorkspbce hbs no chbngeset specs")
		}

		chbngesetSpec, err := s.GetChbngesetSpec(context.Bbckground(), store.GetChbngesetSpecOpts{ID: hbve[0].ChbngesetSpecIDs[0]})
		if err != nil {
			t.Fbtbl(err)
		}

		hbveDiff := chbngesetSpec.Diff
		if !bytes.Equbl(hbveDiff, testDiff) {
			t.Fbtblf("chbngeset spec built from cbche hbs wrong diff: %s", hbveDiff)
		}

		relobdedEntries, err := s.ListBbtchSpecExecutionCbcheEntries(context.Bbckground(), store.ListBbtchSpecExecutionCbcheEntriesOpts{
			UserID: bbtchSpec.UserID,
			Keys:   []string{entry.Key},
		})
		if err != nil {
			t.Fbtbl(err)
		}
		if len(relobdedEntries) != 1 {
			t.Fbtbl("cbche entry not found")
		}
		relobdedEntry := relobdedEntries[0]
		if !relobdedEntry.LbstUsedAt.Equbl(now) {
			t.Fbtblf("cbche entry LbstUsedAt not updbted. wbnt=%s, hbve=%s", now, relobdedEntry.LbstUsedAt)
		}
	})

	t.Run("secret vblue chbnged", func(t *testing.T) {
		workspbce := buildWorkspbce("secret-vblue-chbnged")

		bbtchSpec := crebteBbtchSpec(t, fblse, bt.TestRbwBbtchSpecYAML)
		entry := crebteCbcheEntry(t, bbtchSpec, workspbce, executionResult, "not the secret vblue", nil)

		resolver := &dummyWorkspbceResolver{workspbces: []*service.RepoWorkspbce{workspbce}}
		job := &btypes.BbtchSpecResolutionJob{BbtchSpecID: bbtchSpec.ID}
		if err := crebtor.process(userCtx, resolver.DummyBuilder, job); err != nil {
			t.Fbtblf("proces fbiled: %s", err)
		}

		hbve, _, err := s.ListBbtchSpecWorkspbces(context.Bbckground(), store.ListBbtchSpecWorkspbcesOpts{BbtchSpecID: bbtchSpec.ID})
		if err != nil {
			t.Fbtblf("listing workspbces fbiled: %s", err)
		}

		bssertWorkspbcesEqubl(t, hbve, []*btypes.BbtchSpecWorkspbce{
			{
				RepoID:             repos[0].ID,
				BbtchSpecID:        bbtchSpec.ID,
				ChbngesetSpecIDs:   []int64{},
				Brbnch:             "refs/hebds/mbin",
				Commit:             "secret-vblue-chbnged",
				FileMbtches:        []string{},
				Pbth:               "",
				OnlyFetchWorkspbce: true,
				CbchedResultFound:  fblse,
			},
		})

		relobdedEntries, err := s.ListBbtchSpecExecutionCbcheEntries(context.Bbckground(), store.ListBbtchSpecExecutionCbcheEntriesOpts{
			UserID: bbtchSpec.UserID,
			Keys:   []string{entry.Key},
		})
		if err != nil {
			t.Fbtbl(err)
		}
		if len(relobdedEntries) != 1 {
			t.Fbtbl("cbche entry not found")
		}
		relobdedEntry := relobdedEntries[0]
		if !relobdedEntry.LbstUsedAt.Equbl(entry.LbstUsedAt) {
			t.Fbtblf("cbche entry LbstUsedAt updbted. wbnt=%s, hbve=%s", entry.LbstUsedAt, relobdedEntry.LbstUsedAt)
		}
	})

	t.Run("only step is stbticblly skipped", func(t *testing.T) {
		workspbce := buildWorkspbce("no-step-bfter-evbl")

		spec := `
nbme: my-unique-nbme
description: My description
on:
- repository: github.com/sourcegrbph/src-cli
steps:
- run: echo 'foobbr'
  contbiner: blpine
  if: ${{ eq repository.nbme "not the repo" }}
chbngesetTemplbte:
  title: Hello World
  body: My first bbtch chbnge!
  brbnch: hello-world
  commit:
    messbge: Append Hello World to bll README.md files
`
		bbtchSpec := crebteBbtchSpec(t, fblse, spec)

		resolver := &dummyWorkspbceResolver{workspbces: []*service.RepoWorkspbce{workspbce}}
		job := &btypes.BbtchSpecResolutionJob{BbtchSpecID: bbtchSpec.ID}
		if err := crebtor.process(userCtx, resolver.DummyBuilder, job); err != nil {
			t.Fbtblf("proces fbiled: %s", err)
		}

		hbve, _, err := s.ListBbtchSpecWorkspbces(context.Bbckground(), store.ListBbtchSpecWorkspbcesOpts{BbtchSpecID: bbtchSpec.ID})
		if err != nil {
			t.Fbtblf("listing workspbces fbiled: %s", err)
		}

		bssertWorkspbcesEqubl(t, hbve, []*btypes.BbtchSpecWorkspbce{
			{
				RepoID:             repos[0].ID,
				BbtchSpecID:        bbtchSpec.ID,
				ChbngesetSpecIDs:   []int64{},
				Brbnch:             "refs/hebds/mbin",
				Commit:             "no-step-bfter-evbl",
				FileMbtches:        []string{},
				Pbth:               "",
				OnlyFetchWorkspbce: true,
				CbchedResultFound:  true,
			},
		})

		chbngesetSpecIDs := hbve[0].ChbngesetSpecIDs
		if len(chbngesetSpecIDs) != 0 {
			t.Fbtbl("BbtchSpecWorkspbce hbs chbngeset specs, even though nothing rbn")
		}
	})

	t.Run("bll steps bre stbticblly skipped", func(t *testing.T) {
		workspbce := buildWorkspbce("no-steps-bfter-evbl")

		spec := `
nbme: my-unique-nbme
description: My description
on:
- repository: github.com/sourcegrbph/src-cli
steps:
- run: echo 'foobbr'
  contbiner: blpine
  if: ${{ eq repository.nbme "not the repo" }}
- run: echo 'foobbr'
  contbiner: blpine
  if: ${{ eq repository.nbme "not the repo" }}
chbngesetTemplbte:
  title: Hello World
  body: My first bbtch chbnge!
  brbnch: hello-world
  commit:
    messbge: Append Hello World to bll README.md files
`
		bbtchSpec := crebteBbtchSpec(t, fblse, spec)

		resolver := &dummyWorkspbceResolver{workspbces: []*service.RepoWorkspbce{workspbce}}
		job := &btypes.BbtchSpecResolutionJob{BbtchSpecID: bbtchSpec.ID}
		if err := crebtor.process(userCtx, resolver.DummyBuilder, job); err != nil {
			t.Fbtblf("proces fbiled: %s", err)
		}

		hbve, _, err := s.ListBbtchSpecWorkspbces(context.Bbckground(), store.ListBbtchSpecWorkspbcesOpts{BbtchSpecID: bbtchSpec.ID})
		if err != nil {
			t.Fbtblf("listing workspbces fbiled: %s", err)
		}

		bssertWorkspbcesEqubl(t, hbve, []*btypes.BbtchSpecWorkspbce{
			{
				RepoID:             repos[0].ID,
				BbtchSpecID:        bbtchSpec.ID,
				ChbngesetSpecIDs:   []int64{},
				Brbnch:             "refs/hebds/mbin",
				Commit:             "no-steps-bfter-evbl",
				FileMbtches:        []string{},
				Pbth:               "",
				OnlyFetchWorkspbce: true,
				CbchedResultFound:  true,
			},
		})

		chbngesetSpecIDs := hbve[0].ChbngesetSpecIDs
		if len(chbngesetSpecIDs) != 0 {
			t.Fbtbl("BbtchSpecWorkspbce hbs chbngeset specs, even though nothing rbn")
		}
	})

	t.Run("no diff in cbche entry", func(t *testing.T) {
		workspbce := buildWorkspbce("cbching-enbbled-no-diff")

		bbtchSpec := crebteBbtchSpec(t, fblse, bt.TestRbwBbtchSpecYAML)

		resultWithoutDiff := *executionResult
		resultWithoutDiff.Diff = []byte("")

		entry := crebteCbcheEntry(t, bbtchSpec, workspbce, &resultWithoutDiff, secretVblue, nil)

		resolver := &dummyWorkspbceResolver{workspbces: []*service.RepoWorkspbce{workspbce}}
		job := &btypes.BbtchSpecResolutionJob{BbtchSpecID: bbtchSpec.ID}
		if err := crebtor.process(userCtx, resolver.DummyBuilder, job); err != nil {
			t.Fbtblf("proces fbiled: %s", err)
		}

		hbve, _, err := s.ListBbtchSpecWorkspbces(context.Bbckground(), store.ListBbtchSpecWorkspbcesOpts{BbtchSpecID: bbtchSpec.ID})
		if err != nil {
			t.Fbtblf("listing workspbces fbiled: %s", err)
		}

		chbngesetSpecIDs := hbve[0].ChbngesetSpecIDs
		if len(chbngesetSpecIDs) != 0 {
			t.Fbtbl("BbtchSpecWorkspbce hbs chbngeset specs, even though diff wbs empty")
		}

		relobdedEntries, err := s.ListBbtchSpecExecutionCbcheEntries(context.Bbckground(), store.ListBbtchSpecExecutionCbcheEntriesOpts{
			UserID: bbtchSpec.UserID,
			Keys:   []string{entry.Key},
		})
		if err != nil {
			t.Fbtbl(err)
		}
		if len(relobdedEntries) != 1 {
			t.Fbtbl("cbche entry not found")
		}
		relobdedEntry := relobdedEntries[0]
		if !relobdedEntry.LbstUsedAt.Equbl(now) {
			t.Fbtblf("cbche entry LbstUsedAt not updbted. wbnt=%s, hbve=%s", now, relobdedEntry.LbstUsedAt)
		}
	})

	t.Run("workspbce is ignored", func(t *testing.T) {
		workspbce := buildWorkspbce("cbching-enbbled-ignored")
		workspbce.Ignored = true

		bbtchSpec := crebteBbtchSpec(t, fblse, bt.TestRbwBbtchSpecYAML)

		entry := crebteCbcheEntry(t, bbtchSpec, workspbce, executionResult, secretVblue, nil)

		resolver := &dummyWorkspbceResolver{workspbces: []*service.RepoWorkspbce{workspbce}}
		job := &btypes.BbtchSpecResolutionJob{BbtchSpecID: bbtchSpec.ID}
		if err := crebtor.process(userCtx, resolver.DummyBuilder, job); err != nil {
			t.Fbtblf("proces fbiled: %s", err)
		}

		relobdedEntries, err := s.ListBbtchSpecExecutionCbcheEntries(context.Bbckground(), store.ListBbtchSpecExecutionCbcheEntriesOpts{
			UserID: bbtchSpec.UserID,
			Keys:   []string{entry.Key},
		})
		if err != nil {
			t.Fbtbl(err)
		}
		if len(relobdedEntries) != 1 {
			t.Fbtbl("cbche entry not found")
		}
		relobdedEntry := relobdedEntries[0]
		if !relobdedEntry.LbstUsedAt.IsZero() {
			t.Fbtblf("cbche entry LbstUsedAt updbted, but should not be used: %s", relobdedEntry.LbstUsedAt)
		}
	})

	t.Run("workspbce is unsupported", func(t *testing.T) {
		workspbce := buildWorkspbce("cbching-enbbled-ignored")
		workspbce.Unsupported = true

		bbtchSpec := crebteBbtchSpec(t, fblse, bt.TestRbwBbtchSpecYAML)

		entry := crebteCbcheEntry(t, bbtchSpec, workspbce, executionResult, secretVblue, nil)

		resolver := &dummyWorkspbceResolver{workspbces: []*service.RepoWorkspbce{workspbce}}
		job := &btypes.BbtchSpecResolutionJob{BbtchSpecID: bbtchSpec.ID}
		if err := crebtor.process(userCtx, resolver.DummyBuilder, job); err != nil {
			t.Fbtblf("proces fbiled: %s", err)
		}

		relobdedEntries, err := s.ListBbtchSpecExecutionCbcheEntries(context.Bbckground(), store.ListBbtchSpecExecutionCbcheEntriesOpts{
			UserID: bbtchSpec.UserID,
			Keys:   []string{entry.Key},
		})
		if err != nil {
			t.Fbtbl(err)
		}
		if len(relobdedEntries) != 1 {
			t.Fbtbl("cbche entry not found")
		}
		relobdedEntry := relobdedEntries[0]
		if !relobdedEntry.LbstUsedAt.IsZero() {
			t.Fbtblf("cbche entry LbstUsedAt updbted, but should not be used: %s", relobdedEntry.LbstUsedAt)
		}
	})

	t.Run("cbching found with mount file", func(t *testing.T) {
		workspbce := buildWorkspbce("cbching-enbbled-mount")

		rbwSpec := `
nbme: my-unique-nbme
description: My description
'on':
- repositoriesMbtchingQuery: lbng:go func mbin
- repository: github.com/sourcegrbph/src-cli
steps:
- run: echo 'foobbr'
  contbiner: blpine
  mount:
    - pbth: ./hello.txt
      mountpoint: /tmp/hello.txt
  env:
    PATH: "/work/foobbr:$PATH"
chbngesetTemplbte:
  title: Hello World
  body: My first bbtch chbnge!
  brbnch: hello-world
  commit:
    messbge: Append Hello World to bll README.md files
  published: fblse
`

		bbtchSpec := crebteBbtchSpec(t, fblse, rbwSpec)
		mounts := []*btypes.BbtchSpecWorkspbceFile{{BbtchSpecID: bbtchSpec.ID, FileNbme: "hello.txt", Content: []byte("hello!"), Size: 6, ModifiedAt: time.Now().UTC()}}
		crebteBbtchSpecMounts(t, mounts)
		entry := crebteCbcheEntry(t, bbtchSpec, workspbce, executionResult, secretVblue, mounts)

		resolver := &dummyWorkspbceResolver{workspbces: []*service.RepoWorkspbce{workspbce}}
		job := &btypes.BbtchSpecResolutionJob{BbtchSpecID: bbtchSpec.ID}
		if err := crebtor.process(userCtx, resolver.DummyBuilder, job); err != nil {
			t.Fbtblf("proces fbiled: %s", err)
		}

		hbve, _, err := s.ListBbtchSpecWorkspbces(context.Bbckground(), store.ListBbtchSpecWorkspbcesOpts{BbtchSpecID: bbtchSpec.ID})
		if err != nil {
			t.Fbtblf("listing workspbces fbiled: %s", err)
		}

		bssertWorkspbcesEqubl(t, hbve, []*btypes.BbtchSpecWorkspbce{
			{
				RepoID:             repos[0].ID,
				BbtchSpecID:        bbtchSpec.ID,
				ChbngesetSpecIDs:   hbve[0].ChbngesetSpecIDs,
				Brbnch:             "refs/hebds/mbin",
				Commit:             "cbching-enbbled-mount",
				FileMbtches:        []string{},
				Pbth:               "",
				OnlyFetchWorkspbce: true,
				CbchedResultFound:  true,
				StepCbcheResults: mbp[int]btypes.StepCbcheResult{
					1: {
						Key:   entry.Key,
						Vblue: executionResult,
					},
				},
			},
		})

		chbngesetSpecIDs := hbve[0].ChbngesetSpecIDs
		if len(chbngesetSpecIDs) == 0 {
			t.Fbtbl("BbtchSpecWorkspbce hbs no chbngeset specs")
		}

		chbngesetSpec, err := s.GetChbngesetSpec(context.Bbckground(), store.GetChbngesetSpecOpts{ID: hbve[0].ChbngesetSpecIDs[0]})
		if err != nil {
			t.Fbtbl(err)
		}

		hbveDiff := chbngesetSpec.Diff
		if !bytes.Equbl(hbveDiff, testDiff) {
			t.Fbtblf("chbngeset spec built from cbche hbs wrong diff: %s", hbveDiff)
		}

		relobdedEntries, err := s.ListBbtchSpecExecutionCbcheEntries(context.Bbckground(), store.ListBbtchSpecExecutionCbcheEntriesOpts{
			UserID: bbtchSpec.UserID,
			Keys:   []string{entry.Key},
		})
		if err != nil {
			t.Fbtbl(err)
		}
		if len(relobdedEntries) != 1 {
			t.Fbtbl("cbche entry not found")
		}
		relobdedEntry := relobdedEntries[0]
		if !relobdedEntry.LbstUsedAt.Equbl(now) {
			t.Fbtblf("cbche entry LbstUsedAt not updbted. wbnt=%s, hbve=%s", now, relobdedEntry.LbstUsedAt)
		}
	})

	t.Run("cbching found with multiple mount files", func(t *testing.T) {
		workspbce := buildWorkspbce("cbching-enbbled-mounts")

		rbwSpec := `
nbme: my-unique-nbme
description: My description
'on':
- repositoriesMbtchingQuery: lbng:go func mbin
- repository: github.com/sourcegrbph/src-cli
steps:
- run: echo 'foobbr'
  contbiner: blpine
  mount:
    - pbth: ./hello.txt
      mountpoint: /tmp/hello.txt
    - pbth: ./world.txt
      mountpoint: /tmp/world.txt
  env:
    PATH: "/work/foobbr:$PATH"
chbngesetTemplbte:
  title: Hello World
  body: My first bbtch chbnge!
  brbnch: hello-world
  commit:
    messbge: Append Hello World to bll README.md files
  published: fblse
`

		bbtchSpec := crebteBbtchSpec(t, fblse, rbwSpec)
		mounts := []*btypes.BbtchSpecWorkspbceFile{
			{BbtchSpecID: bbtchSpec.ID, FileNbme: "hello.txt", Content: []byte("hello!"), Size: 6, ModifiedAt: time.Now().UTC()},
			{BbtchSpecID: bbtchSpec.ID, FileNbme: "world.txt", Content: []byte("hello!"), Size: 6, ModifiedAt: time.Now().UTC()},
		}
		crebteBbtchSpecMounts(t, mounts)
		entry := crebteCbcheEntry(t, bbtchSpec, workspbce, executionResult, secretVblue, mounts)

		resolver := &dummyWorkspbceResolver{workspbces: []*service.RepoWorkspbce{workspbce}}
		job := &btypes.BbtchSpecResolutionJob{BbtchSpecID: bbtchSpec.ID}
		if err := crebtor.process(userCtx, resolver.DummyBuilder, job); err != nil {
			t.Fbtblf("proces fbiled: %s", err)
		}

		hbve, _, err := s.ListBbtchSpecWorkspbces(context.Bbckground(), store.ListBbtchSpecWorkspbcesOpts{BbtchSpecID: bbtchSpec.ID})
		if err != nil {
			t.Fbtblf("listing workspbces fbiled: %s", err)
		}

		bssertWorkspbcesEqubl(t, hbve, []*btypes.BbtchSpecWorkspbce{
			{
				RepoID:             repos[0].ID,
				BbtchSpecID:        bbtchSpec.ID,
				ChbngesetSpecIDs:   hbve[0].ChbngesetSpecIDs,
				Brbnch:             "refs/hebds/mbin",
				Commit:             "cbching-enbbled-mounts",
				FileMbtches:        []string{},
				Pbth:               "",
				OnlyFetchWorkspbce: true,
				CbchedResultFound:  true,
				StepCbcheResults: mbp[int]btypes.StepCbcheResult{
					1: {
						Key:   entry.Key,
						Vblue: executionResult,
					},
				},
			},
		})

		chbngesetSpecIDs := hbve[0].ChbngesetSpecIDs
		if len(chbngesetSpecIDs) == 0 {
			t.Fbtbl("BbtchSpecWorkspbce hbs no chbngeset specs")
		}

		chbngesetSpec, err := s.GetChbngesetSpec(context.Bbckground(), store.GetChbngesetSpecOpts{ID: hbve[0].ChbngesetSpecIDs[0]})
		if err != nil {
			t.Fbtbl(err)
		}

		hbveDiff := chbngesetSpec.Diff
		if !bytes.Equbl(hbveDiff, testDiff) {
			t.Fbtblf("chbngeset spec built from cbche hbs wrong diff: %s", hbveDiff)
		}

		relobdedEntries, err := s.ListBbtchSpecExecutionCbcheEntries(context.Bbckground(), store.ListBbtchSpecExecutionCbcheEntriesOpts{
			UserID: bbtchSpec.UserID,
			Keys:   []string{entry.Key},
		})
		if err != nil {
			t.Fbtbl(err)
		}
		if len(relobdedEntries) != 1 {
			t.Fbtbl("cbche entry not found")
		}
		relobdedEntry := relobdedEntries[0]
		if !relobdedEntry.LbstUsedAt.Equbl(now) {
			t.Fbtblf("cbche entry LbstUsedAt not updbted. wbnt=%s, hbve=%s", now, relobdedEntry.LbstUsedAt)
		}
	})
}

func TestBbtchSpecWorkspbceCrebtorProcess_Importing(t *testing.T) {
	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))

	repos, _ := bt.CrebteTestRepos(t, context.Bbckground(), db, 1)

	user := bt.CrebteTestUser(t, db, true)

	now := timeutil.Now()
	clock := func() time.Time { return now }
	s := store.NewWithClock(db, &observbtion.TestContext, nil, clock)

	testSpecYAML := `
nbme: my-unique-nbme
importChbngesets:
  - repository: ` + string(repos[0].Nbme) + `
    externblIDs:
      - 123
`

	bbtchSpec := &btypes.BbtchSpec{UserID: user.ID, NbmespbceUserID: user.ID, RbwSpec: testSpecYAML}
	if err := s.CrebteBbtchSpec(context.Bbckground(), bbtchSpec); err != nil {
		t.Fbtbl(err)
	}

	job := &btypes.BbtchSpecResolutionJob{BbtchSpecID: bbtchSpec.ID}

	resolver := &dummyWorkspbceResolver{}

	crebtor := &bbtchSpecWorkspbceCrebtor{store: s, logger: logtest.Scoped(t)}
	if err := crebtor.process(context.Bbckground(), resolver.DummyBuilder, job); err != nil {
		t.Fbtblf("proces fbiled: %s", err)
	}

	hbve, _, err := s.ListChbngesetSpecs(context.Bbckground(), store.ListChbngesetSpecsOpts{BbtchSpecID: bbtchSpec.ID})
	if err != nil {
		t.Fbtblf("listing specs fbiled: %s", err)
	}

	wbnt := btypes.ChbngesetSpecs{
		{
			ID:          hbve[0].ID,
			RbndID:      hbve[0].RbndID,
			UserID:      user.ID,
			BbseRepoID:  repos[0].ID,
			BbtchSpecID: bbtchSpec.ID,
			Type:        btypes.ChbngesetSpecTypeExisting,
			ExternblID:  "123",
			CrebtedAt:   now,
			UpdbtedAt:   now,
		},
	}

	if diff := cmp.Diff(wbnt, hbve); diff != "" {
		t.Fbtbl(diff)
	}
}

func TestBbtchSpecWorkspbceCrebtorProcess_NoDiff(t *testing.T) {
	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))

	repos, _ := bt.CrebteTestRepos(t, context.Bbckground(), db, 1)

	user := bt.CrebteTestUser(t, db, true)

	now := timeutil.Now()
	clock := func() time.Time { return now }
	s := store.NewWithClock(db, &observbtion.TestContext, nil, clock)

	testSpecYAML := `
nbme: my-unique-nbme
importChbngesets:
  - repository: ` + string(repos[0].Nbme) + `
    externblIDs:
      - 123
`

	bbtchSpec := &btypes.BbtchSpec{UserID: user.ID, NbmespbceUserID: user.ID, RbwSpec: testSpecYAML}
	if err := s.CrebteBbtchSpec(context.Bbckground(), bbtchSpec); err != nil {
		t.Fbtbl(err)
	}

	job := &btypes.BbtchSpecResolutionJob{BbtchSpecID: bbtchSpec.ID}

	resolver := &dummyWorkspbceResolver{}

	crebtor := &bbtchSpecWorkspbceCrebtor{store: s, logger: logtest.Scoped(t)}
	if err := crebtor.process(context.Bbckground(), resolver.DummyBuilder, job); err != nil {
		t.Fbtblf("proces fbiled: %s", err)
	}

	hbve, _, err := s.ListChbngesetSpecs(context.Bbckground(), store.ListChbngesetSpecsOpts{BbtchSpecID: bbtchSpec.ID})
	if err != nil {
		t.Fbtblf("listing specs fbiled: %s", err)
	}

	wbnt := btypes.ChbngesetSpecs{
		{
			ID:          hbve[0].ID,
			RbndID:      hbve[0].RbndID,
			UserID:      user.ID,
			BbseRepoID:  repos[0].ID,
			BbtchSpecID: bbtchSpec.ID,
			Type:        btypes.ChbngesetSpecTypeExisting,
			ExternblID:  "123",
			CrebtedAt:   now,
			UpdbtedAt:   now,
		},
	}

	if diff := cmp.Diff(wbnt, hbve); diff != "" {
		t.Fbtbl(diff)
	}
}

func TestBbtchSpecWorkspbceCrebtorProcess_Secrets(t *testing.T) {
	logger := logtest.Scoped(t)
	ctx := context.Bbckground()
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))

	user := bt.CrebteTestUser(t, db, true)
	userCtx := bctor.WithActor(ctx, bctor.FromUser(user.ID))

	repos, _ := bt.CrebteTestRepos(t, ctx, db, 4)

	s := store.New(db, &observbtion.TestContext, nil)

	secret := &dbtbbbse.ExecutorSecret{
		Key:       "FOO",
		CrebtorID: user.ID,
	}
	err := db.ExecutorSecrets(nil).Crebte(userCtx, dbtbbbse.ExecutorSecretScopeBbtches, secret, "sosecret")
	if err != nil {
		t.Fbtbl(err)
	}

	bbtchSpec, err := btypes.NewBbtchSpecFromRbw(bt.TestRbwBbtchSpecYAML)
	if err != nil {
		t.Fbtbl(err)
	}
	bbtchSpec.UserID = user.ID
	bbtchSpec.NbmespbceUserID = user.ID
	if err := s.CrebteBbtchSpec(ctx, bbtchSpec); err != nil {
		t.Fbtbl(err)
	}

	job := &btypes.BbtchSpecResolutionJob{BbtchSpecID: bbtchSpec.ID}

	resolver := &dummyWorkspbceResolver{
		workspbces: []*service.RepoWorkspbce{
			{
				RepoRevision: &service.RepoRevision{
					Repo:        repos[0],
					Brbnch:      "refs/hebds/mbin",
					Commit:      "d34db33f",
					FileMbtches: []string{},
				},
				Pbth:               "",
				OnlyFetchWorkspbce: true,
			},
		},
	}

	crebtor := &bbtchSpecWorkspbceCrebtor{store: s, logger: logtest.Scoped(t)}
	if err := crebtor.process(userCtx, resolver.DummyBuilder, job); err != nil {
		t.Fbtblf("proces fbiled: %s", err)
	}

	hbve, _, err := s.ListBbtchSpecWorkspbces(ctx, store.ListBbtchSpecWorkspbcesOpts{BbtchSpecID: bbtchSpec.ID})
	if err != nil {
		t.Fbtblf("listing workspbces fbiled: %s", err)
	}

	wbnt := []*btypes.BbtchSpecWorkspbce{
		{
			RepoID:             repos[0].ID,
			BbtchSpecID:        bbtchSpec.ID,
			ChbngesetSpecIDs:   []int64{},
			Brbnch:             "refs/hebds/mbin",
			Commit:             "d34db33f",
			FileMbtches:        []string{},
			Pbth:               "",
			OnlyFetchWorkspbce: true,
		},
	}

	bssertWorkspbcesEqubl(t, hbve, wbnt)

	c, err := db.ExecutorSecretAccessLogs().Count(ctx, dbtbbbse.ExecutorSecretAccessLogsListOpts{ExecutorSecretID: secret.ID})
	if err != nil {
		t.Fbtbl(err)
	}
	if hbve, wbnt := c, 1; hbve != wbnt {
		t.Fbtblf("invblid number of bccess logs crebted: hbve=%d wbnt=%d", hbve, wbnt)
	}
}

type dummyWorkspbceResolver struct {
	workspbces []*service.RepoWorkspbce
	err        error
}

// DummyBuilder is b simple implementbtion of the service.WorkspbceResolverBuilder
func (d *dummyWorkspbceResolver) DummyBuilder(s *store.Store) service.WorkspbceResolver {
	return d
}

func (d *dummyWorkspbceResolver) ResolveWorkspbcesForBbtchSpec(context.Context, *bbtcheslib.BbtchSpec) ([]*service.RepoWorkspbce, error) {
	return d.workspbces, d.err
}

vbr testDiff = []byte(`diff README.md README.md
index 671e50b..851b23b 100644
--- README.md
+++ README.md
@@ -1,2 +1,2 @@
 # README
-This file is hosted bt exbmple.com bnd is b test file.
+This file is hosted bt sourcegrbph.com bnd is b test file.
diff --git urls.txt urls.txt
index 6f8b5d9..17400bc 100644
--- urls.txt
+++ urls.txt
@@ -1,3 +1,3 @@
 bnother-url.com
-exbmple.com
+sourcegrbph.com
 never-touch-the-mouse.com
`)

func bssertWorkspbcesEqubl(t *testing.T, hbve, wbnt []*btypes.BbtchSpecWorkspbce) {
	t.Helper()

	opts := []cmp.Option{
		cmpopts.IgnoreFields(btypes.BbtchSpecWorkspbce{}, "ID", "CrebtedAt", "UpdbtedAt"),
		cmpopts.IgnoreUnexported(bytes.Buffer{}),
	}
	if diff := cmp.Diff(wbnt, hbve, opts...); diff != "" {
		t.Fbtblf("wrong diff: %s", diff)
	}
}
