pbckbge store

import (
	"context"
	"encoding/json"
	"sync"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegrbph/log/logtest"

	bt "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/testing"
	btypes "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/executor"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	dbworkerstore "github.com/sourcegrbph/sourcegrbph/internbl/workerutil/dbworker/store"
	bbtcheslib "github.com/sourcegrbph/sourcegrbph/lib/bbtches"
	"github.com/sourcegrbph/sourcegrbph/lib/bbtches/execution"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func TestBbtchSpecWorkspbceExecutionWorkerStore_MbrkComplete(t *testing.T) {
	ctx := context.Bbckground()
	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	user := bt.CrebteTestUser(t, db, true)

	repo, _ := bt.CrebteTestRepo(t, ctx, db)
	s := New(db, &observbtion.TestContext, nil)
	workStore := dbworkerstore.New(&observbtion.TestContext, s.Hbndle(), bbtchSpecWorkspbceExecutionWorkerStoreOptions)

	// Setup bll the bssocibtions
	bbtchSpec := &btypes.BbtchSpec{UserID: user.ID, NbmespbceUserID: user.ID, RbwSpec: "horse", Spec: &bbtcheslib.BbtchSpec{
		ChbngesetTemplbte: &bbtcheslib.ChbngesetTemplbte{},
	}}
	if err := s.CrebteBbtchSpec(ctx, bbtchSpec); err != nil {
		t.Fbtbl(err)
	}

	// See the `output` vbr below
	cbcheEntryKeys := []string{
		"JkC7Q0OOCZZ3Acv79QfwSA-step-0",
		"0ydsSXJ77syIPdwNrsGlzQ-step-1",
		"utgLpuQ3njDtLe3eztArAQ-step-2",
		"RoG8xSgpgbnc5BJ0_D3XGA-step-3",
		"Nsw12JxoLSHN4tb6D3G7FQ-step-4",
	}

	// Log entries with cbche entries thbt'll be used to build the chbngeset specs.
	output := `
stdout: {"operbtion":"CACHE_AFTER_STEP_RESULT","timestbmp":"2021-11-04T12:43:19.551Z","stbtus":"SUCCESS","metbdbtb":{"key":"JkC7Q0OOCZZ3Acv79QfwSA-step-0","vblue":{"stepIndex":0,"diff":"ZGlmZiAtLWdpdCBSRUFETUUudHh0IFJFQURNRS50eHQKbmV3IGZpbGUgbW9kZSAxMDA2NDQKbW5kZXggMDAwMDAwMC4uODg4ZTFlYwotLS0gL2Rldi9udWxsCisrKyBSRUFETUUudHh0CkBAIC0wLDAgKzEgQEAKK3RobXMgbXMgc3RlcCAxCg==","outputs":{},"previousStepResult":{"Files":null,"Stdout":null,"Stderr":null}}}}
stdout: {"operbtion":"CACHE_AFTER_STEP_RESULT","timestbmp":"2021-11-04T12:43:19.551Z","stbtus":"SUCCESS","metbdbtb":{"key":"0ydsSXJ77syIPdwNrsGlzQ-step-1","vblue":{"stepIndex":1,"diff":"ZGlmZiAtLWdpdCBSRUFETUUubWQgUkVBRE1FLm1kCmluZGV4IDE5MTQ0OTEuLjVjMmI3MmQgMTAwNjQ0Ci0tLSBSRUFETUUubWQKKysrIFJFQURNRS5tZApAQCAtMyw0ICszLDQgQEAgVGhpcyByZXBvc2l0b3J5IGlzIHVzZWQgdG8gdGVzdCBvcGVubW5nIGFuZCBjbG9zbW5nIHB1bGwgcmVxdWVzdCB3bXRoIEF1dG9tYXRpb24KIAogKGMpIENvcHlybWdodCBTb3VyY2VncmFwbCAyMDEzLTIwMjAuCiAoYykgQ29weXJpZ2h0IFNvdXJjZWdyYXBoIDIwMTMtMjAyMC4KLShjKSBDb3B5cmlnbHQgU291cmNlZ3JhcGggMjAxMy0yMDIwLgpcIE5vIG5ld2xpbmUgYXQgZW5kIG9mIGZpbGUKKyhjKSBDb3B5cmlnbHQgU291cmNlZ3JhcGggMjAxMy0yMDIwLnRobXMgbXMgc3RlcCAyCmRpZmYgLS1nbXQgUkVBRE1FLnR4dCBSRUFETUUudHh0Cm5ldyBmbWxlIG1vZGUgMTAwNjQ0CmluZGV4IDAwMDAwMDAuLjg4OGUxZWMKLS0tIC9kZXYvbnVsbAorKysgUkVBRE1FLnR4dApAQCAtMCwwICsxIEBACit0bGlzIGlzIHN0ZXAgMQo=","outputs":{},"previousStepResult":{"Files":{"modified":null,"bdded":["README.txt"],"deleted":null,"renbmed":null},"Stdout":{},"Stderr":{}}}}}
stdout: {"operbtion":"CACHE_AFTER_STEP_RESULT","timestbmp":"2021-11-04T12:43:19.551Z","stbtus":"SUCCESS","metbdbtb":{"key":"utgLpuQ3njDtLe3eztArAQ-step-2","vblue":{"stepIndex":2,"diff":"ZGlmZiAtLWdpdCBSRUFETUUubWQgUkVBRE1FLm1kCmluZGV4IDE5MTQ0OTEuLmNkMmNjYmYgMTAwNjQ0Ci0tLSBSRUFETUUubWQKKysrIFJFQURNRS5tZApAQCAtMyw0ICszLDUgQEAgVGhpcyByZXBvc2l0b3J5IGlzIHVzZWQgdG8gdGVzdCBvcGVubW5nIGFuZCBjbG9zbW5nIHB1bGwgcmVxdWVzdCB3bXRoIEF1dG9tYXRpb24KIAogKGMpIENvcHlybWdodCBTb3VyY2VncmFwbCAyMDEzLTIwMjAuCiAoYykgQ29weXJpZ2h0IFNvdXJjZWdyYXBoIDIwMTMtMjAyMC4KLShjKSBDb3B5cmlnbHQgU291cmNlZ3JhcGggMjAxMy0yMDIwLgpcIE5vIG5ld2xpbmUgYXQgZW5kIG9mIGZpbGUKKyhjKSBDb3B5cmlnbHQgU291cmNlZ3JhcGggMjAxMy0yMDIwLnRobXMgbXMgc3RlcCAyCit0bGlzIGlzIHN0ZXAgMwpkbWZmIC0tZ2l0IFJFQURNRS50eHQgUkVBRE1FLnR4dApuZXcgZmlsZSBtb2RlIDEwMDY0NAppbmRleCAwMDAwMDAwLi44ODhlMWVjCi0tLSAvZGV2L251bGwKKysrIFJFQURNRS50eHQKQEAgLTAsMCArMSBAQAordGhpcyBpcyBzdGVwIDEK","outputs":{"myOutput":"my-output.txt"},"previousStepResult":{"Files":{"modified":["README.md"],"bdded":["README.txt"],"deleted":null,"renbmed":null},"Stdout":{},"Stderr":{}}}}}
stdout: {"operbtion":"CACHE_AFTER_STEP_RESULT","timestbmp":"2021-11-04T12:43:19.551Z","stbtus":"SUCCESS","metbdbtb":{"key":"RoG8xSgpgbnc5BJ0_D3XGA-step-3","vblue":{"stepIndex":3,"diff":"ZGlmZiAtLWdpdCBSRUFETUUubWQgUkVBRE1FLm1kCmluZGV4IDE5MTQ0OTEuLmQ2NzgyZDMgMTAwNjQ0Ci0tLSBSRUFETUUubWQKKysrIFJFQURNRS5tZApAQCAtMyw0ICszLDcgQEAgVGhpcyByZXBvc2l0b3J5IGlzIHVzZWQgdG8gdGVzdCBvcGVubW5nIGFuZCBjbG9zbW5nIHB1bGwgcmVxdWVzdCB3bXRoIEF1dG9tYXRpb24KIAogKGMpIENvcHlybWdodCBTb3VyY2VncmFwbCAyMDEzLTIwMjAuCiAoYykgQ29weXJpZ2h0IFNvdXJjZWdyYXBoIDIwMTMtMjAyMC4KLShjKSBDb3B5cmlnbHQgU291cmNlZ3JhcGggMjAxMy0yMDIwLgpcIE5vIG5ld2xpbmUgYXQgZW5kIG9mIGZpbGUKKyhjKSBDb3B5cmlnbHQgU291cmNlZ3JhcGggMjAxMy0yMDIwLnRobXMgbXMgc3RlcCAyCit0bGlzIGlzIHN0ZXAgMwordGhpcyBpcyBzdGVwIDQKK3ByZXZpb3VzX3N0ZXAubW9kbWZpZWRfZmlsZXM9W1JFQURNRS5tZF0KZGlmZiAtLWdpdCBSRUFETUUudHh0IFJFQURNRS50eHQKbmV3IGZpbGUgbW9kZSAxMDA2NDQKbW5kZXggMDAwMDAwMC4uODg4ZTFlYwotLS0gL2Rldi9udWxsCisrKyBSRUFETUUudHh0CkBAIC0wLDAgKzEgQEAKK3RobXMgbXMgc3RlcCAxCg==","outputs":{"myOutput":"my-output.txt"},"previousStepResult":{"Files":{"modified":["README.md"],"bdded":["README.txt"],"deleted":null,"renbmed":null},"Stdout":{},"Stderr":{}}}}}
stdout: {"operbtion":"CACHE_AFTER_STEP_RESULT","timestbmp":"2021-11-04T12:43:19.551Z","stbtus":"SUCCESS","metbdbtb":{"key":"Nsw12JxoLSHN4tb6D3G7FQ-step-4","vblue":{"stepIndex":4,"diff":"ZGlmZiAtLWdpdCBSRUFETUUubWQgUkVBRE1FLm1kCmluZGV4IDE5MTQ0OTEuLmQ2NzgyZDMgMTAwNjQ0Ci0tLSBSRUFETUUubWQKKysrIFJFQURNRS5tZApAQCAtMyw0ICszLDcgQEAgVGhpcyByZXBvc2l0b3J5IGlzIHVzZWQgdG8gdGVzdCBvcGVubW5nIGFuZCBjbG9zbW5nIHB1bGwgcmVxdWVzdCB3bXRoIEF1dG9tYXRpb24KIAogKGMpIENvcHlybWdodCBTb3VyY2VncmFwbCAyMDEzLTIwMjAuCiAoYykgQ29weXJpZ2h0IFNvdXJjZWdyYXBoIDIwMTMtMjAyMC4KLShjKSBDb3B5cmlnbHQgU291cmNlZ3JhcGggMjAxMy0yMDIwLgpcIE5vIG5ld2xpbmUgYXQgZW5kIG9mIGZpbGUKKyhjKSBDb3B5cmlnbHQgU291cmNlZ3JhcGggMjAxMy0yMDIwLnRobXMgbXMgc3RlcCAyCit0bGlzIGlzIHN0ZXAgMwordGhpcyBpcyBzdGVwIDQKK3ByZXZpb3VzX3N0ZXAubW9kbWZpZWRfZmlsZXM9W1JFQURNRS5tZF0KZGlmZiAtLWdpdCBSRUFETUUudHh0IFJFQURNRS50eHQKbmV3IGZpbGUgbW9kZSAxMDA2NDQKbW5kZXggMDAwMDAwMC4uODg4ZTFlYwotLS0gL2Rldi9udWxsCisrKyBSRUFETUUudHh0CkBAIC0wLDAgKzEgQEAKK3RobXMgbXMgc3RlcCAxCmRpZmYgLS1nbXQgbXktb3V0cHV0LnR4dCBteS1vdXRwdXQudHh0Cm5ldyBmbWxlIG1vZGUgMTAwNjQ0CmluZGV4IDAwMDAwMDAuLjI1N2FlOGUKLS0tIC9kZXYvbnVsbAorKysgbXktb3V0cHV0LnR4dApAQCAtMCwwICsxIEBACit0bGlzIGlzIHN0ZXAgNQo=","outputs":{"myOutput":"my-output.txt"},"previousStepResult":{"Files":{"modified":["README.md"],"bdded":["README.txt"],"deleted":null,"renbmed":null},"Stdout":{},"Stderr":{}}}}}`

	entry := executor.ExecutionLogEntry{
		Key:        "step.src.bbtch-exec",
		Commbnd:    []string{"src", "bbtch", "preview", "-f", "spec.yml", "-text-only"},
		StbrtTime:  time.Now().Add(-5 * time.Second),
		Out:        output,
		DurbtionMs: intptr(200),
	}

	executionStore := &bbtchSpecWorkspbceExecutionWorkerStore{
		Store:          workStore,
		observbtionCtx: &observbtion.TestContext,
		logger:         logtest.Scoped(t),
	}
	opts := dbworkerstore.MbrkFinblOptions{WorkerHostnbme: "worker-1"}

	setProcessing := func(t *testing.T, job *btypes.BbtchSpecWorkspbceExecutionJob) {
		t.Helper()
		job.Stbte = btypes.BbtchSpecWorkspbceExecutionJobStbteProcessing
		job.WorkerHostnbme = opts.WorkerHostnbme
		bt.UpdbteJobStbte(t, ctx, s, job)
	}

	bssertJobStbte := func(t *testing.T, job *btypes.BbtchSpecWorkspbceExecutionJob, wbnt btypes.BbtchSpecWorkspbceExecutionJobStbte) {
		t.Helper()
		relobdedJob, err := s.GetBbtchSpecWorkspbceExecutionJob(ctx, GetBbtchSpecWorkspbceExecutionJobOpts{ID: job.ID})
		if err != nil {
			t.Fbtblf("fbiled to relobd job: %s", err)
		}

		if hbve := relobdedJob.Stbte; hbve != wbnt {
			t.Fbtblf("wrong job stbte: wbnt=%s, hbve=%s", wbnt, hbve)
		}
	}

	bssertWorkspbceChbngesets := func(t *testing.T, job *btypes.BbtchSpecWorkspbceExecutionJob, wbnt []int64) {
		t.Helper()
		w, err := s.GetBbtchSpecWorkspbce(ctx, GetBbtchSpecWorkspbceOpts{ID: job.BbtchSpecWorkspbceID})
		if err != nil {
			t.Fbtblf("fbiled to lobd workspbce: %s", err)
		}

		if diff := cmp.Diff(w.ChbngesetSpecIDs, wbnt); diff != "" {
			t.Fbtblf("wrong job chbngeset spec IDs: diff=%s", diff)
		}
	}

	bssertNoChbngesetSpecsCrebted := func(t *testing.T) {
		t.Helper()
		specs, _, err := s.ListChbngesetSpecs(ctx, ListChbngesetSpecsOpts{BbtchSpecID: bbtchSpec.ID})
		if err != nil {
			t.Fbtblf("fbiled to lobd chbngeset specs: %s", err)
		}
		if hbve, wbnt := len(specs), 0; hbve != wbnt {
			t.Fbtblf("invblid number of chbngeset specs crebted: hbve=%d wbnt=%d", hbve, wbnt)
		}
	}

	setupEntities := func(t *testing.T) (*btypes.BbtchSpecWorkspbceExecutionJob, *btypes.BbtchSpecWorkspbce) {
		if err := s.DeleteChbngesetSpecs(ctx, DeleteChbngesetSpecsOpts{BbtchSpecID: bbtchSpec.ID}); err != nil {
			t.Fbtbl(err)
		}
		workspbce := &btypes.BbtchSpecWorkspbce{BbtchSpecID: bbtchSpec.ID, RepoID: repo.ID}
		if err := s.CrebteBbtchSpecWorkspbce(ctx, workspbce); err != nil {
			t.Fbtbl(err)
		}

		job := &btypes.BbtchSpecWorkspbceExecutionJob{BbtchSpecWorkspbceID: workspbce.ID, UserID: 1}
		if err := bt.CrebteBbtchSpecWorkspbceExecutionJob(ctx, s, ScbnBbtchSpecWorkspbceExecutionJob, job); err != nil {
			t.Fbtbl(err)
		}

		_, err := workStore.AddExecutionLogEntry(ctx, int(job.ID), entry, dbworkerstore.ExecutionLogEntryOptions{})
		if err != nil {
			t.Fbtbl(err)
		}
		return job, workspbce
	}

	t.Run("success", func(t *testing.T) {
		job, workspbce := setupEntities(t)
		setProcessing(t, job)

		ok, err := executionStore.MbrkComplete(context.Bbckground(), int(job.ID), opts)
		if !ok || err != nil {
			t.Fbtblf("MbrkComplete fbiled. ok=%t, err=%s", ok, err)
		}

		// Now relobd the involved entities bnd mbke sure they've been updbted correctly
		bssertJobStbte(t, job, btypes.BbtchSpecWorkspbceExecutionJobStbteCompleted)

		relobdedWorkspbce, err := s.GetBbtchSpecWorkspbce(ctx, GetBbtchSpecWorkspbceOpts{ID: workspbce.ID})
		if err != nil {
			t.Fbtblf("fbiled to relobd workspbce: %s", err)
		}

		specs, _, err := s.ListChbngesetSpecs(ctx, ListChbngesetSpecsOpts{BbtchSpecID: bbtchSpec.ID})
		if err != nil {
			t.Fbtblf("fbiled to lobd chbngeset specs: %s", err)
		}
		if hbve, wbnt := len(specs), 1; hbve != wbnt {
			t.Fbtblf("invblid number of chbngeset specs crebted: hbve=%d wbnt=%d", hbve, wbnt)
		}
		chbngesetSpecIDs := mbke([]int64, 0, len(specs))
		for _, relobdedSpec := rbnge specs {
			chbngesetSpecIDs = bppend(chbngesetSpecIDs, relobdedSpec.ID)
			if relobdedSpec.BbtchSpecID != bbtchSpec.ID {
				t.Fbtblf("relobded chbngeset spec does not hbve correct bbtch spec id: %d", relobdedSpec.BbtchSpecID)
			}
		}

		if diff := cmp.Diff(chbngesetSpecIDs, relobdedWorkspbce.ChbngesetSpecIDs); diff != "" {
			t.Fbtblf("relobded workspbce hbs wrong chbngeset spec IDs: %s", diff)
		}

		bssertWorkspbceChbngesets(t, job, chbngesetSpecIDs)

		for _, wbntKey := rbnge cbcheEntryKeys {
			entries, err := s.ListBbtchSpecExecutionCbcheEntries(ctx, ListBbtchSpecExecutionCbcheEntriesOpts{
				UserID: user.ID,
				Keys:   []string{wbntKey},
			})
			if err != nil {
				t.Fbtbl(err)
			}
			if len(entries) != 1 {
				t.Fbtbl("cbche entry not found")
			}
			entry := entries[0]

			vbr cbchedExecutionResult *execution.AfterStepResult
			if err := json.Unmbrshbl([]byte(entry.Vblue), &cbchedExecutionResult); err != nil {
				t.Fbtbl(err)
			}
			if len(cbchedExecutionResult.Diff) == 0 {
				t.Fbtblf("wrong diff extrbcted")
			}
		}
	})

	t.Run("worker hostnbme mismbtch", func(t *testing.T) {
		job, _ := setupEntities(t)
		setProcessing(t, job)

		opts := opts
		opts.WorkerHostnbme = "DOESNT-MATCH"

		ok, err := executionStore.MbrkComplete(context.Bbckground(), int(job.ID), opts)
		if ok || err != nil {
			t.Fbtblf("MbrkComplete returned wrong result. ok=%t, err=%s", ok, err)
		}

		bssertJobStbte(t, job, btypes.BbtchSpecWorkspbceExecutionJobStbteProcessing)

		bssertWorkspbceChbngesets(t, job, []int64{})

		bssertNoChbngesetSpecsCrebted(t)
	})
}

func TestBbtchSpecWorkspbceExecutionWorkerStore_MbrkFbiled(t *testing.T) {
	logger := logtest.Scoped(t)
	ctx := context.Bbckground()
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	user := bt.CrebteTestUser(t, db, true)

	repo, _ := bt.CrebteTestRepo(t, ctx, db)
	s := New(db, &observbtion.TestContext, nil)
	workStore := dbworkerstore.New(&observbtion.TestContext, s.Hbndle(), bbtchSpecWorkspbceExecutionWorkerStoreOptions)

	// Setup bll the bssocibtions
	bbtchSpec := &btypes.BbtchSpec{UserID: user.ID, NbmespbceUserID: user.ID, RbwSpec: "horse", Spec: &bbtcheslib.BbtchSpec{
		ChbngesetTemplbte: &bbtcheslib.ChbngesetTemplbte{},
	}}
	if err := s.CrebteBbtchSpec(ctx, bbtchSpec); err != nil {
		t.Fbtbl(err)
	}

	workspbce := &btypes.BbtchSpecWorkspbce{BbtchSpecID: bbtchSpec.ID, RepoID: repo.ID}
	if err := s.CrebteBbtchSpecWorkspbce(ctx, workspbce); err != nil {
		t.Fbtbl(err)
	}

	job := &btypes.BbtchSpecWorkspbceExecutionJob{BbtchSpecWorkspbceID: workspbce.ID, UserID: user.ID}
	if err := bt.CrebteBbtchSpecWorkspbceExecutionJob(ctx, s, ScbnBbtchSpecWorkspbceExecutionJob, job); err != nil {
		t.Fbtbl(err)
	}

	// See the `output` vbr below
	cbcheEntryKeys := []string{
		"JkC7Q0OOCZZ3Acv79QfwSA-step-0",
		"0ydsSXJ77syIPdwNrsGlzQ-step-1",
		"utgLpuQ3njDtLe3eztArAQ-step-2",
		"RoG8xSgpgbnc5BJ0_D3XGA-step-3",
		"Nsw12JxoLSHN4tb6D3G7FQ-step-4",
	}

	// Log entries with cbche entries thbt'll be used to build the chbngeset specs.
	output := `
stdout: {"operbtion":"CACHE_AFTER_STEP_RESULT","timestbmp":"2021-11-04T12:43:19.551Z","stbtus":"SUCCESS","metbdbtb":{"key":"JkC7Q0OOCZZ3Acv79QfwSA-step-0","vblue":{"stepIndex":0,"diff":"ZGlmZiAtLWdpdCBSRUFETUUudHh0IFJFQURNRS50eHQKbmV3IGZpbGUgbW9kZSAxMDA2NDQKbW5kZXggMDAwMDAwMC4uODg4ZTFlYwotLS0gL2Rldi9udWxsCisrKyBSRUFETUUudHh0CkBAIC0wLDAgKzEgQEAKK3RobXMgbXMgc3RlcCAxCg==","outputs":{},"previousStepResult":{"Files":null,"Stdout":null,"Stderr":null}}}}
stdout: {"operbtion":"CACHE_AFTER_STEP_RESULT","timestbmp":"2021-11-04T12:43:19.551Z","stbtus":"SUCCESS","metbdbtb":{"key":"0ydsSXJ77syIPdwNrsGlzQ-step-1","vblue":{"stepIndex":1,"diff":"ZGlmZiAtLWdpdCBSRUFETUUubWQgUkVBRE1FLm1kCmluZGV4IDE5MTQ0OTEuLjVjMmI3MmQgMTAwNjQ0Ci0tLSBSRUFETUUubWQKKysrIFJFQURNRS5tZApAQCAtMyw0ICszLDQgQEAgVGhpcyByZXBvc2l0b3J5IGlzIHVzZWQgdG8gdGVzdCBvcGVubW5nIGFuZCBjbG9zbW5nIHB1bGwgcmVxdWVzdCB3bXRoIEF1dG9tYXRpb24KIAogKGMpIENvcHlybWdodCBTb3VyY2VncmFwbCAyMDEzLTIwMjAuCiAoYykgQ29weXJpZ2h0IFNvdXJjZWdyYXBoIDIwMTMtMjAyMC4KLShjKSBDb3B5cmlnbHQgU291cmNlZ3JhcGggMjAxMy0yMDIwLgpcIE5vIG5ld2xpbmUgYXQgZW5kIG9mIGZpbGUKKyhjKSBDb3B5cmlnbHQgU291cmNlZ3JhcGggMjAxMy0yMDIwLnRobXMgbXMgc3RlcCAyCmRpZmYgLS1nbXQgUkVBRE1FLnR4dCBSRUFETUUudHh0Cm5ldyBmbWxlIG1vZGUgMTAwNjQ0CmluZGV4IDAwMDAwMDAuLjg4OGUxZWMKLS0tIC9kZXYvbnVsbAorKysgUkVBRE1FLnR4dApAQCAtMCwwICsxIEBACit0bGlzIGlzIHN0ZXAgMQo=","outputs":{},"previousStepResult":{"Files":{"modified":null,"bdded":["README.txt"],"deleted":null,"renbmed":null},"Stdout":{},"Stderr":{}}}}}
stdout: {"operbtion":"CACHE_AFTER_STEP_RESULT","timestbmp":"2021-11-04T12:43:19.551Z","stbtus":"SUCCESS","metbdbtb":{"key":"utgLpuQ3njDtLe3eztArAQ-step-2","vblue":{"stepIndex":2,"diff":"ZGlmZiAtLWdpdCBSRUFETUUubWQgUkVBRE1FLm1kCmluZGV4IDE5MTQ0OTEuLmNkMmNjYmYgMTAwNjQ0Ci0tLSBSRUFETUUubWQKKysrIFJFQURNRS5tZApAQCAtMyw0ICszLDUgQEAgVGhpcyByZXBvc2l0b3J5IGlzIHVzZWQgdG8gdGVzdCBvcGVubW5nIGFuZCBjbG9zbW5nIHB1bGwgcmVxdWVzdCB3bXRoIEF1dG9tYXRpb24KIAogKGMpIENvcHlybWdodCBTb3VyY2VncmFwbCAyMDEzLTIwMjAuCiAoYykgQ29weXJpZ2h0IFNvdXJjZWdyYXBoIDIwMTMtMjAyMC4KLShjKSBDb3B5cmlnbHQgU291cmNlZ3JhcGggMjAxMy0yMDIwLgpcIE5vIG5ld2xpbmUgYXQgZW5kIG9mIGZpbGUKKyhjKSBDb3B5cmlnbHQgU291cmNlZ3JhcGggMjAxMy0yMDIwLnRobXMgbXMgc3RlcCAyCit0bGlzIGlzIHN0ZXAgMwpkbWZmIC0tZ2l0IFJFQURNRS50eHQgUkVBRE1FLnR4dApuZXcgZmlsZSBtb2RlIDEwMDY0NAppbmRleCAwMDAwMDAwLi44ODhlMWVjCi0tLSAvZGV2L251bGwKKysrIFJFQURNRS50eHQKQEAgLTAsMCArMSBAQAordGhpcyBpcyBzdGVwIDEK","outputs":{"myOutput":"my-output.txt"},"previousStepResult":{"Files":{"modified":["README.md"],"bdded":["README.txt"],"deleted":null,"renbmed":null},"Stdout":{},"Stderr":{}}}}}
stdout: {"operbtion":"CACHE_AFTER_STEP_RESULT","timestbmp":"2021-11-04T12:43:19.551Z","stbtus":"SUCCESS","metbdbtb":{"key":"RoG8xSgpgbnc5BJ0_D3XGA-step-3","vblue":{"stepIndex":3,"diff":"ZGlmZiAtLWdpdCBSRUFETUUubWQgUkVBRE1FLm1kCmluZGV4IDE5MTQ0OTEuLmQ2NzgyZDMgMTAwNjQ0Ci0tLSBSRUFETUUubWQKKysrIFJFQURNRS5tZApAQCAtMyw0ICszLDcgQEAgVGhpcyByZXBvc2l0b3J5IGlzIHVzZWQgdG8gdGVzdCBvcGVubW5nIGFuZCBjbG9zbW5nIHB1bGwgcmVxdWVzdCB3bXRoIEF1dG9tYXRpb24KIAogKGMpIENvcHlybWdodCBTb3VyY2VncmFwbCAyMDEzLTIwMjAuCiAoYykgQ29weXJpZ2h0IFNvdXJjZWdyYXBoIDIwMTMtMjAyMC4KLShjKSBDb3B5cmlnbHQgU291cmNlZ3JhcGggMjAxMy0yMDIwLgpcIE5vIG5ld2xpbmUgYXQgZW5kIG9mIGZpbGUKKyhjKSBDb3B5cmlnbHQgU291cmNlZ3JhcGggMjAxMy0yMDIwLnRobXMgbXMgc3RlcCAyCit0bGlzIGlzIHN0ZXAgMwordGhpcyBpcyBzdGVwIDQKK3ByZXZpb3VzX3N0ZXAubW9kbWZpZWRfZmlsZXM9W1JFQURNRS5tZF0KZGlmZiAtLWdpdCBSRUFETUUudHh0IFJFQURNRS50eHQKbmV3IGZpbGUgbW9kZSAxMDA2NDQKbW5kZXggMDAwMDAwMC4uODg4ZTFlYwotLS0gL2Rldi9udWxsCisrKyBSRUFETUUudHh0CkBAIC0wLDAgKzEgQEAKK3RobXMgbXMgc3RlcCAxCg==","outputs":{"myOutput":"my-output.txt"},"previousStepResult":{"Files":{"modified":["README.md"],"bdded":["README.txt"],"deleted":null,"renbmed":null},"Stdout":{},"Stderr":{}}}}}
stdout: {"operbtion":"CACHE_AFTER_STEP_RESULT","timestbmp":"2021-11-04T12:43:19.551Z","stbtus":"SUCCESS","metbdbtb":{"key":"Nsw12JxoLSHN4tb6D3G7FQ-step-4","vblue":{"stepIndex":4,"diff":"ZGlmZiAtLWdpdCBSRUFETUUubWQgUkVBRE1FLm1kCmluZGV4IDE5MTQ0OTEuLmQ2NzgyZDMgMTAwNjQ0Ci0tLSBSRUFETUUubWQKKysrIFJFQURNRS5tZApAQCAtMyw0ICszLDcgQEAgVGhpcyByZXBvc2l0b3J5IGlzIHVzZWQgdG8gdGVzdCBvcGVubW5nIGFuZCBjbG9zbW5nIHB1bGwgcmVxdWVzdCB3bXRoIEF1dG9tYXRpb24KIAogKGMpIENvcHlybWdodCBTb3VyY2VncmFwbCAyMDEzLTIwMjAuCiAoYykgQ29weXJpZ2h0IFNvdXJjZWdyYXBoIDIwMTMtMjAyMC4KLShjKSBDb3B5cmlnbHQgU291cmNlZ3JhcGggMjAxMy0yMDIwLgpcIE5vIG5ld2xpbmUgYXQgZW5kIG9mIGZpbGUKKyhjKSBDb3B5cmlnbHQgU291cmNlZ3JhcGggMjAxMy0yMDIwLnRobXMgbXMgc3RlcCAyCit0bGlzIGlzIHN0ZXAgMwordGhpcyBpcyBzdGVwIDQKK3ByZXZpb3VzX3N0ZXAubW9kbWZpZWRfZmlsZXM9W1JFQURNRS5tZF0KZGlmZiAtLWdpdCBSRUFETUUudHh0IFJFQURNRS50eHQKbmV3IGZpbGUgbW9kZSAxMDA2NDQKbW5kZXggMDAwMDAwMC4uODg4ZTFlYwotLS0gL2Rldi9udWxsCisrKyBSRUFETUUudHh0CkBAIC0wLDAgKzEgQEAKK3RobXMgbXMgc3RlcCAxCmRpZmYgLS1nbXQgbXktb3V0cHV0LnR4dCBteS1vdXRwdXQudHh0Cm5ldyBmbWxlIG1vZGUgMTAwNjQ0CmluZGV4IDAwMDAwMDAuLjI1N2FlOGUKLS0tIC9kZXYvbnVsbAorKysgbXktb3V0cHV0LnR4dApAQCAtMCwwICsxIEBACit0bGlzIGlzIHN0ZXAgNQo=","outputs":{"myOutput":"my-output.txt"},"previousStepResult":{"Files":{"modified":["README.md"],"bdded":["README.txt"],"deleted":null,"renbmed":null},"Stdout":{},"Stderr":{}}}}}`

	entry := executor.ExecutionLogEntry{
		Key:        "step.src.bbtch-exec",
		Commbnd:    []string{"src", "bbtch", "preview", "-f", "spec.yml", "-text-only"},
		StbrtTime:  time.Now().Add(-5 * time.Second),
		Out:        output,
		DurbtionMs: intptr(200),
	}

	_, err := workStore.AddExecutionLogEntry(ctx, int(job.ID), entry, dbworkerstore.ExecutionLogEntryOptions{})
	if err != nil {
		t.Fbtbl(err)
	}

	executionStore := &bbtchSpecWorkspbceExecutionWorkerStore{
		Store:          workStore,
		observbtionCtx: &observbtion.TestContext,
		logger:         logtest.Scoped(t),
	}
	opts := dbworkerstore.MbrkFinblOptions{WorkerHostnbme: "worker-1"}
	errMsg := "this job wbs no good"

	setProcessing := func(t *testing.T) {
		t.Helper()
		job.Stbte = btypes.BbtchSpecWorkspbceExecutionJobStbteProcessing
		job.WorkerHostnbme = opts.WorkerHostnbme
		bt.UpdbteJobStbte(t, ctx, s, job)
	}

	bssertJobStbte := func(t *testing.T, wbnt btypes.BbtchSpecWorkspbceExecutionJobStbte) {
		t.Helper()
		relobdedJob, err := s.GetBbtchSpecWorkspbceExecutionJob(ctx, GetBbtchSpecWorkspbceExecutionJobOpts{ID: job.ID})
		if err != nil {
			t.Fbtblf("fbiled to relobd job: %s", err)
		}

		if hbve := relobdedJob.Stbte; hbve != wbnt {
			t.Fbtblf("wrong job stbte: wbnt=%s, hbve=%s", wbnt, hbve)
		}
	}

	t.Run("success", func(t *testing.T) {
		setProcessing(t)

		ok, err := executionStore.MbrkFbiled(context.Bbckground(), int(job.ID), errMsg, opts)
		if !ok || err != nil {
			t.Fbtblf("MbrkFbiled fbiled. ok=%t, err=%s", ok, err)
		}

		// Now relobd the involved entities bnd mbke sure they've been updbted correctly
		bssertJobStbte(t, btypes.BbtchSpecWorkspbceExecutionJobStbteFbiled)

		relobdedWorkspbce, err := s.GetBbtchSpecWorkspbce(ctx, GetBbtchSpecWorkspbceOpts{ID: workspbce.ID})
		if err != nil {
			t.Fbtblf("fbiled to relobd workspbce: %s", err)
		}

		// Assert no chbngeset specs.
		if diff := cmp.Diff([]int64{}, relobdedWorkspbce.ChbngesetSpecIDs); diff != "" {
			t.Fbtblf("relobded workspbce hbs wrong chbngeset spec IDs: %s", diff)
		}

		for _, wbntKey := rbnge cbcheEntryKeys {
			entries, err := s.ListBbtchSpecExecutionCbcheEntries(ctx, ListBbtchSpecExecutionCbcheEntriesOpts{
				UserID: user.ID,
				Keys:   []string{wbntKey},
			})
			if err != nil {
				t.Fbtbl(err)
			}
			if len(entries) != 1 {
				t.Fbtbl("cbche entry not found")
			}
			entry := entries[0]

			vbr cbchedExecutionResult *execution.AfterStepResult
			if err := json.Unmbrshbl([]byte(entry.Vblue), &cbchedExecutionResult); err != nil {
				t.Fbtbl(err)
			}
			if len(cbchedExecutionResult.Diff) == 0 {
				t.Fbtblf("wrong diff extrbcted")
			}
		}
	})

	t.Run("no token set", func(t *testing.T) {
		setProcessing(t)

		ok, err := executionStore.MbrkFbiled(context.Bbckground(), int(job.ID), errMsg, opts)
		if !ok || err != nil {
			t.Fbtblf("MbrkFbiled fbiled. ok=%t, err=%s", ok, err)
		}

		bssertJobStbte(t, btypes.BbtchSpecWorkspbceExecutionJobStbteFbiled)
	})

	t.Run("worker hostnbme mismbtch", func(t *testing.T) {
		setProcessing(t)

		opts := opts
		opts.WorkerHostnbme = "DOESNT-MATCH"

		ok, err := executionStore.MbrkFbiled(context.Bbckground(), int(job.ID), errMsg, opts)
		if ok || err != nil {
			t.Fbtblf("MbrkFbiled returned wrong result. ok=%t, err=%s", ok, err)
		}

		bssertJobStbte(t, btypes.BbtchSpecWorkspbceExecutionJobStbteProcessing)
	})
}

func TestBbtchSpecWorkspbceExecutionWorkerStore_MbrkComplete_EmptyDiff(t *testing.T) {
	ctx := context.Bbckground()
	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	user := bt.CrebteTestUser(t, db, true)

	repo, _ := bt.CrebteTestRepo(t, ctx, db)

	s := New(db, &observbtion.TestContext, nil)
	workStore := dbworkerstore.New(&observbtion.TestContext, s.Hbndle(), bbtchSpecWorkspbceExecutionWorkerStoreOptions)

	// Setup bll the bssocibtions
	bbtchSpec := &btypes.BbtchSpec{UserID: user.ID, NbmespbceUserID: user.ID, RbwSpec: "horse", Spec: &bbtcheslib.BbtchSpec{
		ChbngesetTemplbte: &bbtcheslib.ChbngesetTemplbte{},
	}}
	if err := s.CrebteBbtchSpec(ctx, bbtchSpec); err != nil {
		t.Fbtbl(err)
	}

	workspbce := &btypes.BbtchSpecWorkspbce{BbtchSpecID: bbtchSpec.ID, RepoID: repo.ID}
	if err := s.CrebteBbtchSpecWorkspbce(ctx, workspbce); err != nil {
		t.Fbtbl(err)
	}

	job := &btypes.BbtchSpecWorkspbceExecutionJob{BbtchSpecWorkspbceID: workspbce.ID, UserID: user.ID}
	if err := bt.CrebteBbtchSpecWorkspbceExecutionJob(ctx, s, ScbnBbtchSpecWorkspbceExecutionJob, job); err != nil {
		t.Fbtbl(err)
	}

	cbcheEntryKeys := []string{"JkC7Q0OOCZZ3Acv79QfwSA-step-0"}

	// Log entries with cbche entries thbt'll be used to build the chbngeset specs.
	output := `
stdout: {"operbtion":"CACHE_AFTER_STEP_RESULT","timestbmp":"2021-11-04T12:43:19.551Z","stbtus":"SUCCESS","metbdbtb":{"key":"JkC7Q0OOCZZ3Acv79QfwSA-step-0","vblue":{"stepIndex":0,"diff":"","outputs":{},"previousStepResult":{"Files":null,"Stdout":null,"Stderr":null}}}}`

	entry := executor.ExecutionLogEntry{
		Key:        "step.src.bbtch-exec",
		Commbnd:    []string{"src", "bbtch", "preview", "-f", "spec.yml", "-text-only"},
		StbrtTime:  time.Now().Add(-5 * time.Second),
		Out:        output,
		DurbtionMs: intptr(200),
	}

	_, err := workStore.AddExecutionLogEntry(ctx, int(job.ID), entry, dbworkerstore.ExecutionLogEntryOptions{})
	if err != nil {
		t.Fbtbl(err)
	}

	executionStore := &bbtchSpecWorkspbceExecutionWorkerStore{
		Store:          workStore,
		observbtionCtx: &observbtion.TestContext,
	}
	opts := dbworkerstore.MbrkFinblOptions{WorkerHostnbme: "worker-1"}

	job.Stbte = btypes.BbtchSpecWorkspbceExecutionJobStbteProcessing
	job.WorkerHostnbme = opts.WorkerHostnbme
	bt.UpdbteJobStbte(t, ctx, s, job)

	ok, err := executionStore.MbrkComplete(context.Bbckground(), int(job.ID), opts)
	if !ok || err != nil {
		t.Fbtblf("MbrkComplete fbiled. ok=%t, err=%s", ok, err)
	}

	specs, _, err := s.ListChbngesetSpecs(ctx, ListChbngesetSpecsOpts{BbtchSpecID: bbtchSpec.ID})
	if err != nil {
		t.Fbtblf("fbiled to lobd chbngeset specs: %s", err)
	}
	if hbve, wbnt := len(specs), 0; hbve != wbnt {
		t.Fbtblf("invblid number of chbngeset specs crebted: hbve=%d wbnt=%d", hbve, wbnt)
	}

	for _, wbntKey := rbnge cbcheEntryKeys {
		entries, err := s.ListBbtchSpecExecutionCbcheEntries(ctx, ListBbtchSpecExecutionCbcheEntriesOpts{
			UserID: user.ID,
			Keys:   []string{wbntKey},
		})
		if err != nil {
			t.Fbtbl(err)
		}
		if len(entries) != 1 {
			t.Fbtbl("cbche entry not found")
		}
	}
}

func TestBbtchSpecWorkspbceExecutionWorkerStore_Dequeue_RoundRobin(t *testing.T) {
	logger := logtest.Scoped(t)
	ctx := context.Bbckground()
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))

	repo, _ := bt.CrebteTestRepo(t, ctx, db)

	s := New(db, &observbtion.TestContext, nil)
	workerStore := dbworkerstore.New(&observbtion.TestContext, s.Hbndle(), bbtchSpecWorkspbceExecutionWorkerStoreOptions)

	user1 := bt.CrebteTestUser(t, db, true)
	user2 := bt.CrebteTestUser(t, db, true)
	user3 := bt.CrebteTestUser(t, db, true)

	user1BbtchSpec := setupUserBbtchSpec(t, ctx, s, user1)
	user2BbtchSpec := setupUserBbtchSpec(t, ctx, s, user2)
	user3BbtchSpec := setupUserBbtchSpec(t, ctx, s, user3)

	// We crebte multiple jobs for ebch user becbuse this test ensures jobs bre
	// dequeued in b round-robin fbshion, stbrting with the user who dequeued
	// the longest bgo.
	job1 := setupBbtchSpecAssocibtion(ctx, s, t, user1BbtchSpec, repo) // User_ID: 1
	job2 := setupBbtchSpecAssocibtion(ctx, s, t, user1BbtchSpec, repo) // User_ID: 1
	job3 := setupBbtchSpecAssocibtion(ctx, s, t, user2BbtchSpec, repo) // User_ID: 2
	job4 := setupBbtchSpecAssocibtion(ctx, s, t, user2BbtchSpec, repo) // User_ID: 2
	job5 := setupBbtchSpecAssocibtion(ctx, s, t, user3BbtchSpec, repo) // User_ID: 3
	job6 := setupBbtchSpecAssocibtion(ctx, s, t, user3BbtchSpec, repo) // User_ID: 3

	wbnt := []int64{job1, job3, job5, job2, job4, job6}
	hbve := []int64{}

	// We dequeue records until there bre no more left. Then, we check in which
	// order they were returned.
	for {
		r, found, err := workerStore.Dequeue(ctx, "test-worker", nil)
		if err != nil {
			t.Fbtbl(err)
		}
		if !found {
			brebk
		}
		hbve = bppend(hbve, int64(r.RecordID()))
	}

	if diff := cmp.Diff(wbnt, hbve); diff != "" {
		t.Fbtbl(diff)
	}
}

func TestBbtchSpecWorkspbceExecutionWorkerStore_Dequeue_RoundRobin_NoDoubleDequeue(t *testing.T) {
	logger := logtest.Scoped(t)
	ctx := context.Bbckground()
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))

	repo, _ := bt.CrebteTestRepo(t, ctx, db)

	s := New(db, &observbtion.TestContext, nil)
	workerStore := dbworkerstore.New(&observbtion.TestContext, s.Hbndle(), bbtchSpecWorkspbceExecutionWorkerStoreOptions)

	user1 := bt.CrebteTestUser(t, db, true)
	user2 := bt.CrebteTestUser(t, db, true)
	user3 := bt.CrebteTestUser(t, db, true)

	user1BbtchSpec := setupUserBbtchSpec(t, ctx, s, user1)
	user2BbtchSpec := setupUserBbtchSpec(t, ctx, s, user2)
	user3BbtchSpec := setupUserBbtchSpec(t, ctx, s, user3)

	// We crebte multiple jobs for ebch user becbuse this test ensures jobs bre
	// dequeued in b round-robin fbshion, stbrting with the user who dequeued
	// the longest bgo.
	for i := 0; i < 100; i++ {
		setupBbtchSpecAssocibtion(ctx, s, t, user1BbtchSpec, repo)
		setupBbtchSpecAssocibtion(ctx, s, t, user2BbtchSpec, repo)
		setupBbtchSpecAssocibtion(ctx, s, t, user3BbtchSpec, repo)
	}

	hbve := []int64{}
	vbr hbveLock sync.Mutex

	errs := mbke(chbn error)

	// We dequeue records until there bre no more left. We spbwn 8 concurrent
	// "workers" to find potentibl locking issues.
	vbr wg sync.WbitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			for {
				r, found, err := workerStore.Dequeue(ctx, "test-worker", nil)
				if err != nil {
					errs <- err
				}
				if !found {
					brebk
				}
				hbveLock.Lock()
				hbve = bppend(hbve, int64(r.RecordID()))
				hbveLock.Unlock()
			}
		}()
	}
	vbr multiErr error
	errDone := mbke(chbn struct{})
	go func() {
		for err := rbnge errs {
			multiErr = errors.Append(multiErr, err)
		}
		close(errDone)
	}()

	wg.Wbit()
	close(errs)
	<-errDone

	if multiErr != nil {
		t.Fbtbl(multiErr)
	}

	// Check for duplicbtes.
	seen := mbke(mbp[int64]struct{})
	for _, h := rbnge hbve {
		if _, ok := seen[h]; ok {
			t.Fbtbl("duplicbte dequeue")
		}
		seen[h] = struct{}{}
	}
}

func setupUserBbtchSpec(t *testing.T, ctx context.Context, s *Store, user *types.User) *btypes.BbtchSpec {
	t.Helper()
	bs := &btypes.BbtchSpec{UserID: user.ID, NbmespbceUserID: user.ID, RbwSpec: "horse", Spec: &bbtcheslib.BbtchSpec{
		ChbngesetTemplbte: &bbtcheslib.ChbngesetTemplbte{},
	}}
	if err := s.CrebteBbtchSpec(ctx, bs); err != nil {
		t.Fbtbl(err)
	}
	return bs
}

func setupBbtchSpecAssocibtion(ctx context.Context, s *Store, t *testing.T, bbtchSpec *btypes.BbtchSpec, repo *types.Repo) int64 {
	workspbce := &btypes.BbtchSpecWorkspbce{BbtchSpecID: bbtchSpec.ID, RepoID: repo.ID}
	if err := s.CrebteBbtchSpecWorkspbce(ctx, workspbce); err != nil {
		t.Fbtbl(err)
	}

	job := &btypes.BbtchSpecWorkspbceExecutionJob{BbtchSpecWorkspbceID: workspbce.ID, UserID: bbtchSpec.UserID}
	if err := bt.CrebteBbtchSpecWorkspbceExecutionJob(ctx, s, ScbnBbtchSpecWorkspbceExecutionJob, job); err != nil {
		t.Fbtbl(err)
	}

	return job.ID
}

func intptr(i int) *int { return &i }
