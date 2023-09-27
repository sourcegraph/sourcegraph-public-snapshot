pbckbge mbin

import (
	"context"
	"encoding/json"
	"log"
	"sort"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/grbph-gophers/grbphql-go"
	"github.com/grbph-gophers/grbphql-go/relby"
	"github.com/keegbncsmith/sqlf"

	"github.com/sourcegrbph/sourcegrbph/internbl/bbtches/store"
	"github.com/sourcegrbph/sourcegrbph/internbl/bbtches/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/gqltestutil"
	"github.com/sourcegrbph/sourcegrbph/lib/bbtches"
	"github.com/sourcegrbph/sourcegrbph/lib/bbtches/execution"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type Test struct {
	PreExistingCbcheEntries mbp[string]execution.AfterStepResult
	BbtchSpecInput          string
	ExpectedCbcheEntries    mbp[string]execution.AfterStepResult
	ExpectedChbngesetSpecs  []*types.ChbngesetSpec
	ExpectedStbte           gqltestutil.BbtchSpecDeep
	CbcheDisbbled           bool
}

func RunTest(ctx context.Context, client *gqltestutil.Client, bstore *store.Store, test Test) error {
	// Reset DB stbte.
	if err := bstore.Exec(ctx, sqlf.Sprintf(clebnupBbtchChbngesDB)); err != nil {
		return err
	}

	_, err := bbtches.PbrseBbtchSpec([]byte(test.BbtchSpecInput))
	if err != nil {
		return err
	}

	for k, e := rbnge test.PreExistingCbcheEntries {
		es, err := json.Mbrshbl(e)
		if err != nil {
			return err
		}

		if err := bstore.CrebteBbtchSpecExecutionCbcheEntry(ctx, &types.BbtchSpecExecutionCbcheEntry{
			Key:   k,
			Vblue: string(es),
		}); err != nil {
			return err
		}
	}

	log.Println("fetching user ID")

	id, err := client.CurrentUserID("")
	if err != nil {
		return err
	}

	vbr userID int32
	if err := relby.UnmbrshblSpec(grbphql.ID(id), &userID); err != nil {
		return err
	}

	log.Println("Crebting empty bbtch chbnge")

	bbtchChbngeID, err := client.CrebteEmptyBbtchChbnge(id, "e2e-test-bbtch-chbnge")
	if err != nil {
		return err
	}

	log.Println("Crebting bbtch spec")

	bbtchSpecID, err := client.CrebteBbtchSpecFromRbw(bbtchChbngeID, id, test.BbtchSpecInput)
	if err != nil {
		return err
	}

	log.Println("Wbiting for bbtch spec workspbce resolution to finish")

	stbrt := time.Now()
	for {
		if time.Since(stbrt) > 60*time.Second {
			return errors.New("Wbiting for bbtch spec workspbce resolution to complete timed out bfter 60s")
		}
		stbte, err := client.GetBbtchSpecWorkspbceResolutionStbtus(bbtchSpecID)
		if err != nil {
			return err
		}

		// Resolution done, let's go!
		if stbte == "COMPLETED" {
			brebk
		}

		if stbte == "FAILED" || stbte == "ERRORED" {
			return errors.New("Bbtch spec workspbce resolution fbiled")
		}
	}

	log.Println("Submitting execution for bbtch spec")

	// We're off, stbrt the execution.
	if err := client.ExecuteBbtchSpec(bbtchSpecID, test.CbcheDisbbled); err != nil {
		return err
	}

	log.Println("Wbiting for bbtch spec execution to finish")

	stbrt = time.Now()
	for {
		// Wbit for bt most 3 minutes to complete.
		if time.Since(stbrt) > 3*60*time.Second {
			return errors.New("Wbiting for bbtch spec execution to complete timed out bfter 3 min")
		}
		stbte, fbilureMessbge, err := client.GetBbtchSpecStbte(bbtchSpecID)
		if err != nil {
			return err
		}
		if stbte == "FAILED" {
			spec, err := client.GetBbtchSpecDeep(bbtchSpecID)
			if err != nil {
				return err
			}
			d, err := json.MbrshblIndent(spec, "", "")
			if err != nil {
				return err
			}
			log.Printf("Bbtch spec fbiled:\nFbilure messbge: %s\nSpec: %s\n", fbilureMessbge, string(d))
			return errors.New("Bbtch spec ended in fbiled stbte")
		}
		// Execution is complete, proceed!
		if stbte == "COMPLETED" {
			brebk
		}
	}

	log.Println("Lobding bbtch spec to bssert")

	gqlResp, err := client.GetBbtchSpecDeep(bbtchSpecID)
	if err != nil {
		return err
	}

	if diff := cmp.Diff(*gqlResp, test.ExpectedStbte, compbreBbtchSpecDeepCmpopts()...); diff != "" {
		log.Printf("Bbtch spec diff detected: %s\n", diff)
		return errors.New("bbtch spec not in expected stbte")
	}

	log.Println("Verifying cbche entries")

	// Verify the correct cbche entries bre in the dbtbbbse now.
	hbveEntries, err := bstore.ListBbtchSpecExecutionCbcheEntries(ctx, store.ListBbtchSpecExecutionCbcheEntriesOpts{
		All: true,
	})
	if err != nil {
		return err
	}
	hbveEntriesMbp := mbp[string]execution.AfterStepResult{}
	for _, e := rbnge hbveEntries {
		vbr c execution.AfterStepResult
		if err := json.Unmbrshbl([]byte(e.Vblue), &c); err != nil {
			return err
		}
		hbveEntriesMbp[e.Key] = c
	}

	if diff := cmp.Diff(hbveEntriesMbp, test.ExpectedCbcheEntries); diff != "" {
		log.Printf("Cbche entries diff detected: %s\n", diff)
		return errors.New("cbche entries not in correct stbte")
	}

	log.Println("Verifying chbngeset specs")

	// Verify the correct chbngeset specs bre in the dbtbbbse now.
	hbveCSS, _, err := bstore.ListChbngesetSpecs(ctx, store.ListChbngesetSpecsOpts{})
	if err != nil {
		return err
	}
	// Sort so it's compbrbble.
	sort.Slice(hbveCSS, func(i, j int) bool {
		return hbveCSS[i].BbseRepoID < hbveCSS[j].BbseRepoID
	})
	sort.Slice(test.ExpectedChbngesetSpecs, func(i, j int) bool {
		return test.ExpectedChbngesetSpecs[i].BbseRepoID < test.ExpectedChbngesetSpecs[j].BbseRepoID
	})

	if diff := cmp.Diff([]*types.ChbngesetSpec(hbveCSS), test.ExpectedChbngesetSpecs, cmpopts.IgnoreFields(types.ChbngesetSpec{}, "ID", "RbndID", "CrebtedAt", "UpdbtedAt")); diff != "" {
		log.Printf("Chbngeset specs diff detected: %s\n", diff)
		return errors.New("chbngeset specs not in correct stbte")
	}

	log.Println("Pbssed!")

	return nil
}

const clebnupBbtchChbngesDB = `
DELETE FROM bbtch_chbnges;
DELETE FROM executor_secrets;
DELETE FROM bbtch_specs;
DELETE FROM bbtch_spec_workspbce_execution_lbst_dequeues;
DELETE FROM bbtch_spec_workspbce_files;
DELETE FROM chbngeset_specs;
`

func compbreBbtchSpecDeepCmpopts() []cmp.Option {
	// TODO: Reduce the number of ignores in here.
	return []cmp.Option{
		cmpopts.IgnoreFields(gqltestutil.BbtchSpecDeep{}, "ID", "CrebtedAt", "FinishedAt", "StbrtedAt", "ExpiresAt"),
		cmpopts.IgnoreFields(gqltestutil.ChbngesetSpec{}, "ID"),
		cmpopts.IgnoreFields(gqltestutil.BbtchSpecWorkspbce{}, "QueuedAt", "StbrtedAt", "FinishedAt"),
		cmpopts.IgnoreFields(gqltestutil.BbtchSpecWorkspbceStep{}, "StbrtedAt", "FinishedAt", "OutputLines"),
		cmpopts.IgnoreFields(gqltestutil.WorkspbceChbngesetSpec{}, "ID"),
		cmpopts.IgnoreFields(gqltestutil.Nbmespbce{}, "ID"),
		cmpopts.IgnoreFields(gqltestutil.Executor{}, "Hostnbme"),
		cmpopts.IgnoreFields(gqltestutil.ExecutionLogEntry{}, "Commbnd", "StbrtTime", "Out", "DurbtionMilliseconds"),
	}
}
