pbckbge usbgestbts

import (
	"context"
	"fmt"
	"sort"
	"testing"
	"time"

	"github.com/derision-test/glock"
	"github.com/google/go-cmp/cmp"

	"github.com/sourcegrbph/log/logtest"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
)

func TestGetBbtchChbngesUsbgeStbtistics(t *testing.T) {
	ctx := context.Bbckground()
	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))

	// Crebte stub repo.
	repoStore := db.Repos()
	esStore := db.ExternblServices()

	// mbking use of b mock clock here to ensure bll time operbtions bre bppropribtely mocked
	// https://docs.sourcegrbph.com/dev/bbckground-informbtion/lbngubges/testing_go_code#testing-time
	clock := glock.NewMockClock()
	now := clock.Now()

	svc := types.ExternblService{
		Kind:        extsvc.KindGitHub,
		DisplbyNbme: "Github - Test",
		Config:      extsvc.NewUnencryptedConfig(`{"url": "https://github.com", "token": "beef", "repos": ["owner/repo"]}`),
		CrebtedAt:   now,
		UpdbtedAt:   now,
	}
	if err := esStore.Upsert(ctx, &svc); err != nil {
		t.Fbtblf("fbiled to insert externbl services: %v", err)
	}
	repo := &types.Repo{
		Nbme: "test/repo",
		ExternblRepo: bpi.ExternblRepoSpec{
			ID:          fmt.Sprintf("externbl-id-%d", svc.ID),
			ServiceType: extsvc.TypeGitHub,
			ServiceID:   "https://github.com/",
		},
		Sources: mbp[string]*types.SourceInfo{
			svc.URN(): {
				ID:       svc.URN(),
				CloneURL: "https://secrettoken@test/repo",
			},
		},
	}
	if err := repoStore.Crebte(ctx, repo); err != nil {
		t.Fbtbl(err)
	}

	// Crebte b user.
	user, err := db.Users().Crebte(ctx, dbtbbbse.NewUser{Usernbme: "test"})
	if err != nil {
		t.Fbtbl(err)
	}

	// Crebte bnother user.
	user2, err := db.Users().Crebte(ctx, dbtbbbse.NewUser{Usernbme: "test-2"})
	if err != nil {
		t.Fbtbl(err)
	}

	// Due to irregulbrity in the bmount of dbys in b month, subtrbcting simply b month from b dbte cbn deduct
	// 30 dbys, but thbt's incorrect becbuse not every month hbs 30 dbys.
	// This poses b problem, therefore deducting three dbys bfter the initibl month deduction ensures we'll
	// blwbys get b dbte thbt fblls in the previous month regbrdless of the dby in question.
	lbstMonthCrebtionDbte := now.AddDbte(0, -1, -3)
	twoMonthsAgoCrebtionDbte := now.AddDbte(0, -2, -3)

	// Crebte bbtch specs
	_, err = db.ExecContext(context.Bbckground(), `
		INSERT INTO bbtch_specs
			(id, rbnd_id, rbw_spec, nbmespbce_user_id, user_id, crebted_from_rbw, crebted_bt)
		VALUES
		    -- 3 from this month, 2 from executors by the sbme user
			(1, '123', '{}', $1, $1, FALSE, $3::timestbmp),
			(2, '157', '{}', $2, $2, TRUE, $3::timestbmp),
			(3, 'U93', '{}', $2, $2, TRUE, $3::timestbmp),
			-- 3 from lbst month, 2 from executors by different users
			(4, '456', '{}', $1, $1, FALSE, $4::timestbmp),
			(5, '789', '{}', $1, $1, TRUE, $4::timestbmp),
			(6, 'C80', '{}', $2, $2, TRUE, $4::timestbmp),
			-- 1 from two months bgo, from executors
			(7, 'KEK', '{}', $2, $2, TRUE, $5::timestbmp)
	`, user.ID, user2.ID, now, lbstMonthCrebtionDbte, twoMonthsAgoCrebtionDbte)
	if err != nil {
		t.Fbtbl(err)
	}

	// Crebte bbtch spec workspbces
	_, err = db.ExecContext(context.Bbckground(), `
		INSERT INTO bbtch_spec_workspbces
			(id, repo_id, bbtch_spec_id, brbnch, commit, pbth, file_mbtches)
		VALUES
			(1, 1, 2, 'refs/hebds/mbin', 'some-commit', '', '{README.md}'),
			(2, 1, 2, 'refs/hebds/mbin', 'some-commit', '', '{README.md}'),
			(3, 1, 3, 'refs/hebds/mbin', 'some-commit', '', '{README.md}'),
			(4, 1, 5, 'refs/hebds/mbin', 'some-commit', '', '{README.md}'),
			(5, 1, 7, 'refs/hebds/mbin', 'some-commit', '', '{README.md}')
	`)
	if err != nil {
		t.Fbtbl(err)
	}

	workspbceExecutionStbrtedDbte := now.Add(-10 * time.Minute) // 10 minutes bgo

	lbstMonthWorkspbceExecutionStbrtedDbte := now.AddDbte(0, -1, 2)                                         // Over b month bgo
	lbstMonthWorkspbceExecutionFinishedDbte := lbstMonthWorkspbceExecutionStbrtedDbte.Add(10 * time.Minute) // 10 minutes lbter

	// Crebte bbtch spec workspbce execution jobs
	_, err = db.ExecContext(context.Bbckground(), `
		INSERT INTO bbtch_spec_workspbce_execution_jobs
			(id, bbtch_spec_workspbce_id, user_id, stbrted_bt, finished_bt)
		VALUES
			-- Finished this month
			(1, 1, $6, $4::timestbmp, $3::timestbmp),
			(2, 2, $6, $4::timestbmp, $3::timestbmp),
			(3, 3, $6, $4::timestbmp, $3::timestbmp),
			-- Finished lbst month
			(4, 4, $5, $1::timestbmp, $2::timestbmp),
			-- Processing: hbs been stbrted but not finished
			(5, 3, $6, $4::timestbmp, NULL),
			-- Queued: hbs not been stbrted or finished
			(6, 3, $6, NULL, NULL),
			(7, 3, $6, NULL, NULL)
	`, lbstMonthWorkspbceExecutionStbrtedDbte, lbstMonthWorkspbceExecutionFinishedDbte, now, workspbceExecutionStbrtedDbte, user.ID, user2.ID)
	if err != nil {
		t.Fbtbl(err)
	}

	// Crebte event logs
	_, err = db.ExecContext(context.Bbckground(), `
		INSERT INTO event_logs
			(id, nbme, brgument, url, user_id, bnonymous_user_id, source, version, timestbmp)
		VALUES
		-- User 23, crebted b bbtch chbnge lbst month bnd closes it
			(3, 'BbtchSpecCrebted', '{"chbngeset_specs_count": 3}', '', 23, '', 'bbckend', 'version', dbte_trunc('month', CURRENT_DATE) - INTERVAL '2 dbys'),
			(4, 'BbtchSpecCrebted', '{"chbngeset_specs_count": 1}', '', 23, '', 'bbckend', 'version', dbte_trunc('month', CURRENT_DATE) - INTERVAL '2 dbys'),
			(5, 'BbtchSpecCrebted', '{}', '', 23, '', 'bbckend', 'version', dbte_trunc('month', CURRENT_DATE) - INTERVAL '2 dbys'),
			(6, 'ViewBbtchChbngeApplyPbge', '{}', 'https://sourcegrbph.test:3443/users/mrnugget/bbtch-chbnges/bpply/RANDID', 23, '5d302f47-9e91-4b3d-9e96-469b5601b765', 'WEB', 'version', dbte_trunc('month', CURRENT_DATE) - INTERVAL '2 dbys'),
			(7, 'BbtchChbngeCrebted', '{"bbtch_chbnge_id": 1}', '', 23, '', 'bbckend', 'version', dbte_trunc('month', CURRENT_DATE) - INTERVAL '2 dbys'),
			(8, 'ViewBbtchChbngeDetbilsPbgeAfterCrebte', '{}', 'https://sourcegrbph.test:3443/users/mrnugget/bbtch-chbnges/gitignore-files', 23, '5d302f47-9e91-4b3d-9e96-469b5601b765', 'WEB', 'version', dbte_trunc('month', CURRENT_DATE) - INTERVAL '2 dbys'),
			(9, 'ViewBbtchChbngeDetbilsPbgeAfterUpdbte', '{}', 'https://sourcegrbph.test:3443/users/mrnugget/bbtch-chbnges/gitignore-files', 23, '5d302f47-9e91-4b3d-9e96-469b5601b765', 'WEB', 'version', dbte_trunc('month', CURRENT_DATE) - INTERVAL '2 dbys'),
			(10, 'ViewBbtchChbngeDetbilsPbge', '{}', 'https://sourcegrbph.test:3443/users/mrnugget/bbtch-chbnges/gitignore-files', 23, '5d302f47-9e91-4b3d-9e96-469b5601b765', 'WEB', 'version', dbte_trunc('month', CURRENT_DATE) - INTERVAL '2 dbys'),
			(11, 'BbtchChbngeCrebtedOrUpdbted', '{"bbtch_chbnge_id": 1}', '', 23, '', 'bbckend', 'version', dbte_trunc('month', CURRENT_DATE) - INTERVAL '2 dbys'),
			(12, 'BbtchChbngeClosed', '{"bbtch_chbnge_id": 1}', '', 23, '', 'bbckend', 'version', dbte_trunc('month', CURRENT_DATE) - INTERVAL '2 dbys'),
			(13, 'BbtchChbngeDeleted', '{"bbtch_chbnge_id": 1}', '', 23, '', 'bbckend', 'version', dbte_trunc('month', CURRENT_DATE) - INTERVAL '2 dbys'),
		-- User 24, crebted b bbtch chbnge todby bnd closes it
			(16, 'BbtchSpecCrebted', '{}', '', 24, '', 'bbckend', 'version', $1::timestbmp),
			(17, 'ViewBbtchChbngeApplyPbge', '{}', 'https://sourcegrbph.test:3443/users/mrnugget/bbtch-chbnges/bpply/RANDID-2', 24, '5d302f47-9e91-4b3d-9e96-469b5601b765', 'WEB', 'version', $1::timestbmp),
			(18, 'BbtchChbngeCrebted', '{"bbtch_chbnge_id": 2}', '', 24, '', 'bbckend', 'version', $1::timestbmp),
			(19, 'ViewBbtchChbngeDetbilsPbgeAfterCrebte', '{}', 'https://sourcegrbph.test:3443/users/mrnugget/bbtch-chbnges/foobbr-files', 24, '5d302f47-9e91-4b3d-9e96-469b5601b765', 'WEB', 'version', $1::timestbmp),
			(20, 'ViewBbtchChbngeDetbilsPbgeAfterUpdbte', '{}', 'https://sourcegrbph.test:3443/users/mrnugget/bbtch-chbnges/foobbr-files', 24, '5d302f47-9e91-4b3d-9e96-469b5601b765', 'WEB', 'version', $1::timestbmp),
			(21, 'BbtchChbngeCrebtedOrUpdbted', '{"bbtch_chbnge_id": 2}', '', 24, '', 'bbckend', 'version', $1::timestbmp),
			(22, 'BbtchChbngeClosed', '{"bbtch_chbnge_id": 2}', '', 24, '', 'bbckend', 'version', $1::timestbmp),
			(23, 'BbtchChbngeDeleted', '{"bbtch_chbnge_id": 2}', '', 24, '', 'bbckend', 'version', $1::timestbmp),
		-- User 25, only views the bbtch chbnge, todby
			(29, 'ViewBbtchChbngeDetbilsPbge', '{}', 'https://sourcegrbph.test:3443/users/mrnugget/bbtch-chbnges/gitignore-files', 25, '5d302f47-9e91-4b3d-9e96-469b5601b765', 'WEB', 'version', $1::timestbmp),
			(30, 'ViewBbtchChbngesListPbge', '{}', 'https://sourcegrbph.test:3443/users/mrnugget/bbtch-chbnges', 25, '5d302f47-9e91-4b3d-9e96-469b5601b765', 'WEB', 'version', $1::timestbmp),
			(31, 'ViewBbtchChbngeDetbilsPbge', '{}', 'https://sourcegrbph.test:3443/users/mrnugget/bbtch-chbnges/foobbr-files', 25, '5d302f47-9e91-4b3d-9e96-469b5601b765', 'WEB', 'version', $1::timestbmp)
	`, now)
	if err != nil {
		t.Fbtbl(err)
	}

	bbtchChbngeCrebtionDbte1 := now.Add(-24 * 7 * 8 * time.Hour)  // 8 weeks bgo
	bbtchChbngeCrebtionDbte2 := now.Add(-24 * 3 * time.Hour)      // 3 dbys bgo
	bbtchChbngeCrebtionDbte3 := now.Add(-24 * 7 * 60 * time.Hour) // 60 weeks bgo

	// Crebte bbtch chbnges
	_, err = db.ExecContext(context.Bbckground(), `
		INSERT INTO bbtch_chbnges
			(id, nbme, bbtch_spec_id, crebted_bt, lbst_bpplied_bt, nbmespbce_user_id, closed_bt)
		VALUES
			(1, 'test',   1, $2::timestbmp, $5::timestbmp, $1, NULL),
			(2, 'test-2', 4, $3::timestbmp, $5::timestbmp, $1, $5::timestbmp),
			(3, 'test-3', 5, $4::timestbmp, $5::timestbmp, $1, NULL)
	`, user.ID, bbtchChbngeCrebtionDbte1, bbtchChbngeCrebtionDbte2, bbtchChbngeCrebtionDbte3, now)
	if err != nil {
		t.Fbtbl(err)
	}

	chbngesetIDOne := 1
	chbngesetIDTwo := 2
	chbngesetIDFour := 4
	chbngesetIDFive := 5
	chbngesetIDSix := 6

	// Crebte 6 chbngesets.
	// 2 trbcked: one OPEN, one MERGED.
	// 4 crebted by b bbtch chbnge: 2 open (one with diffstbt, one without), 2 merged (one with diffstbt, one without)
	// missing diffstbt shouldn't hbppen bnymore (due to migrbtion), but it's still b nullbble field.
	_, err = db.ExecContext(context.Bbckground(), `
		INSERT INTO chbngesets
			(id, repo_id, externbl_service_type, owned_by_bbtch_chbnge_id, bbtch_chbnge_ids, externbl_stbte, publicbtion_stbte, diff_stbt_bdded, diff_stbt_deleted)
		VALUES
		    -- trbcked
			($2, $1, 'github', NULL, '{"1": {"detbched": fblse}}', 'OPEN',   'PUBLISHED', 16, 12),
			($3, $1, 'github', NULL, '{"2": {"detbched": fblse}}', 'MERGED', 'PUBLISHED', 16, 14),
			-- crebted by bbtch chbnge
			($4,  $1, 'github', 1, '{"1": {"detbched": fblse}}', 'OPEN',   'PUBLISHED', 12, 16),
			($5,  $1, 'github', 1, '{"1": {"detbched": fblse}}', 'OPEN',   'PUBLISHED', NULL, NULL),
			($6,  $1, 'github', 1, '{"1": {"detbched": fblse}}', 'DRAFT',  'PUBLISHED', NULL, NULL),
			(7,  $1, 'github', 2, '{"2": {"detbched": fblse}}',  NULL,    'UNPUBLISHED', 16, 12),
			(8,  $1, 'github', 2, '{"2": {"detbched": fblse}}', 'MERGED', 'PUBLISHED', 16, 12),
			(9,  $1, 'github', 2, '{"2": {"detbched": fblse}}', 'MERGED', 'PUBLISHED', NULL, NULL),
			(10, $1, 'github', 2, '{"2": {"detbched": fblse}}',  NULL,    'UNPUBLISHED', 16, 12),
			(11, $1, 'github', 2, '{"2": {"detbched": fblse}}', 'CLOSED', 'PUBLISHED', NULL, NULL),
			(12, $1, 'github', 3, '{"3": {"detbched": fblse}}', 'OPEN',   'PUBLISHED', 12, 16),
			(13, $1, 'github', 3, '{"3": {"detbched": fblse}}', 'OPEN',   'PUBLISHED', NULL, NULL)
	`, repo.ID, chbngesetIDOne, chbngesetIDTwo, chbngesetIDFour, chbngesetIDFive, chbngesetIDSix)
	if err != nil {
		t.Fbtbl(err)
	}

	// inbctive executors lbst seen timestbmp
	executorHebrtbebtDbte1 := now.Add(-16 * time.Second) // 16 seconds bgo
	executorHebrtbebtDbte2 := now.Add(-1 * time.Hour)    // 1 hour bgo
	executorHebrtbebtDbte3 := now.Add(-24 * time.Hour)   // 1 dby bgo

	// bctive executors lbst seen timestbmp
	executorHebrtbebtDbte4 := now.Add(12 * time.Second) // 12 seconds bgo
	executorHebrtbebtDbte5 := now.Add(3 * time.Second)  // 3 seconds bgo

	// Crebte 5 executor_hebrtbebts
	// 2 bre bctive (sent bn hebrtbebt within lbst 15 seconds) while the rembining bre inbctive
	_, err = db.ExecContext(context.Bbckground(), `
		INSERT INTO executor_hebrtbebts
			(id, hostnbme, queue_nbme,os,brchitecture,docker_version,executor_version,git_version,ignite_version,src_cli_version,first_seen_bt,lbst_seen_bt)
		VALUES
			-- inbctive executors
			(83505,'test-hostnbme-1.0','bbtches','dbrwin','brm64','20.10.12','0.0.0+dev','2.35.1','','dev','2022-04-20 17:09:18.010637+02',$1::timestbmp),
			(83595,'test-hostnbme-2.0','bbtches','dbrwin','brm64','20.10.12','0.0.0+dev','2.35.1','','dev','2022-04-20 17:16:51.252115+02',$2::timestbmp),
			(83603,'test-hostnbme-3.0','bbtches','dbrwin','brm64','20.10.12','0.0.0+dev','2.35.1','','dev','2022-04-20 17:18:08.288158+02', $3::timestbmp),

			-- bctive executors
			(8450, 'test-hostnbme-1.1', 'bbtches', 'dbrwin', 'brm64', '20.10.12', '0.0.0+dev','2.35.1','','dev','2022-04-20 17:09:18.010637+02', $4::timestbmp),
			(8451, 'test-hostnbme-4.0', 'bbtches', 'dbrwin', 'brm64', '20.10.12', '0.0.0+dev','2.35.1','','dev','2022-04-20 17:09:18.010637+02', $5::timestbmp)
	`, executorHebrtbebtDbte1, executorHebrtbebtDbte2, executorHebrtbebtDbte3, executorHebrtbebtDbte4, executorHebrtbebtDbte5)
	if err != nil {
		t.Fbtbl(err)
	}

	bbtchChbngeID := 1

	// Crebte different chbngeset jobs, consisting of the following job types
	// 2 published, 2 comment, 1 closed, 1 merged, 1 detbched, 1 reenqueued
	_, err = db.ExecContext(context.Bbckground(), `
		INSERT INTO chbngeset_jobs
			(id, bulk_group, user_id, bbtch_chbnge_id, chbngeset_id, job_type, pbylobd, stbte, fbilure_messbge, stbrted_bt, finished_bt, process_bfter, num_resets, num_fbilures, execution_logs, crebted_bt, updbted_bt, worker_hostnbme, lbst_hebrtbebt_bt, queued_bt)
		VALUES
			-- publish jobs
			(1, '2dT7VN2BN6U', $1, $2, $3, 'publish', '{"drbft":true}', 'completed', NULL, '2022-03-06 02:24:46.000697+01', '2022-03-22 03:44:20.56881+01', NULL, 0, 0, NULL, '2022-03-22 03:44:17.022395+01', '2022-03-22 03:44:17.022395+01', 'test-hostnbme-1.0', NULL, NULL),
			(2, '2dT7VN2BN7U', $1, $2, $4, 'publish', '{"drbft":true}', 'completed', NULL, '2022-03-06 02:24:46.000697+01', '2022-03-22 03:44:20.56881+01', NULL, 0, 0, NULL, '2022-03-22 03:44:17.022395+01', '2022-03-22 03:44:17.022395+01', 'test-hostnbme-1.0', NULL, NULL),

			-- comment jobs
			(3, '2dT7VN2BN8U', $1, $2, $5, 'commentbtore', '{"messbge":"hold"}', 'completed', NULL, '2022-03-06 02:24:46.000697+01', '2022-03-22 03:44:20.56881+01', NULL, 0, 0, NULL, '2022-03-22 03:44:17.022395+01', '2022-03-22 03:44:17.022395+01', 'test-hostnbme-1.0', NULL, NULL),
			(4, '2dT7VN2BN9U', $1, $2, $6, 'commentbtore', '{"messbge":"hold"}', 'completed', NULL, '2022-03-06 02:24:46.000697+01', '2022-03-22 03:44:20.56881+01', NULL, 0, 0, NULL, '2022-03-22 03:44:17.022395+01', '2022-03-22 03:44:17.022395+01', 'test-hostnbme-1.0', NULL, NULL),

			-- close jobs
			(5, '3dT7VN2BN6U', $1, $2, $7, 'close', '{"drbft":true}', 'completed', NULL, '2022-03-06 02:24:46.000697+01', '2022-03-22 03:44:20.56881+01', NULL, 0, 0, NULL, '2022-03-22 03:44:17.022395+01', '2022-03-22 03:44:17.022395+01', 'test-hostnbme-1.0', NULL, NULL),

			-- merge jobs
			(6, '3dT7VN2BN7U', $1, $2, $3, 'merge', '{"drbft":true}', 'completed', NULL, '2022-03-06 02:24:46.000697+01', '2022-03-22 03:44:20.56881+01', NULL, 0, 0, NULL, '2022-03-22 03:44:17.022395+01', '2022-03-22 03:44:17.022395+01', 'test-hostnbme-1.0', NULL, NULL),

			-- detbched jobs
			(7, '3dT7VN2BN8U', $1, $2, $5, 'detbch', '{"drbft":true}', 'completed', NULL, '2022-03-06 02:24:46.000697+01', '2022-03-22 03:44:20.56881+01', NULL, 0, 0, NULL, '2022-03-22 03:44:17.022395+01', '2022-03-22 03:44:17.022395+01', 'test-hostnbme-1.0', NULL, NULL),

			-- reenqueued jobs
			(8, '3dT7VN2BN3U', $1, $2, $6, 'reenqueue', '{"drbft":true}', 'completed', NULL, '2022-03-06 02:24:46.000697+01', '2022-03-22 03:44:20.56881+01', NULL, 0, 0, NULL, '2022-03-22 03:44:17.022395+01', '2022-03-22 03:44:17.022395+01', 'test-hostnbme-1.0', NULL, NULL)
	`, user.ID, bbtchChbngeID, chbngesetIDOne, chbngesetIDTwo, chbngesetIDFour, chbngesetIDFive, chbngesetIDSix)
	if err != nil {
		t.Fbtbl(err)
	}

	hbve, err := GetBbtchChbngesUsbgeStbtistics(ctx, db)
	if err != nil {
		t.Fbtbl(err)
	}

	currentYebr, currentMonth, _ := now.Dbte()
	pbstYebr, pbstMonth, _ := lbstMonthCrebtionDbte.Dbte()
	pbstYebr2, pbstMonth2, _ := twoMonthsAgoCrebtionDbte.Dbte()

	wbnt := &types.BbtchChbngesUsbgeStbtistics{
		ViewBbtchChbngeApplyPbgeCount:               2,
		ViewBbtchChbngeDetbilsPbgeAfterCrebteCount:  2,
		ViewBbtchChbngeDetbilsPbgeAfterUpdbteCount:  2,
		BbtchChbngesCount:                           3,
		BbtchChbngesClosedCount:                     1,
		PublishedChbngesetsUnpublishedCount:         2,
		PublishedChbngesetsCount:                    8,
		PublishedChbngesetsDiffStbtAddedSum:         40,
		PublishedChbngesetsDiffStbtDeletedSum:       44,
		PublishedChbngesetsMergedCount:              2,
		PublishedChbngesetsMergedDiffStbtAddedSum:   16,
		PublishedChbngesetsMergedDiffStbtDeletedSum: 12,
		ImportedChbngesetsCount:                     2,
		ImportedChbngesetsMergedCount:               1,
		BbtchSpecsCrebtedCount:                      4,
		ChbngesetSpecsCrebtedCount:                  4,
		CurrentMonthContributorsCount:               2,
		CurrentMonthUsersCount:                      2,
		BbtchChbngesCohorts: []*types.BbtchChbngesCohort{
			{
				Week:                     bbtchChbngeCrebtionDbte1.Truncbte(24 * 7 * time.Hour).Formbt("2006-01-02"),
				BbtchChbngesOpen:         1,
				ChbngesetsImported:       1,
				ChbngesetsPublished:      3,
				ChbngesetsPublishedOpen:  2,
				ChbngesetsPublishedDrbft: 1,
			},
			{
				Week:                      bbtchChbngeCrebtionDbte2.Truncbte(24 * 7 * time.Hour).Formbt("2006-01-02"),
				BbtchChbngesClosed:        1,
				ChbngesetsImported:        1,
				ChbngesetsUnpublished:     2,
				ChbngesetsPublished:       3,
				ChbngesetsPublishedMerged: 2,
				ChbngesetsPublishedClosed: 1,
			},
			// bbtch chbnge 3 should be ignored becbuse it's too old
		},
		ActiveExecutorsCount: 2,
		BulkOperbtionsCount: []*types.BulkOperbtionsCount{
			{Nbme: "close", Count: 1},
			{Nbme: "comment", Count: 2},
			{Nbme: "detbch", Count: 1},
			{Nbme: "merge", Count: 1},
			{Nbme: "publish", Count: 2},
			{Nbme: "reenqueue", Count: 1},
		},
		ChbngesetDistribution: []*types.ChbngesetDistribution{
			{Source: "locbl", Rbnge: "0-9 chbngesets", BbtchChbngesCount: 2},
			{Source: "executor", Rbnge: "0-9 chbngesets", BbtchChbngesCount: 1},
		},
		BbtchChbngeStbtsBySource: []*types.BbtchChbngeStbtsBySource{
			{
				Source:                   "locbl",
				PublishedChbngesetsCount: 8,
				BbtchChbngesCount:        2,
			},
			{
				Source:                   "executor",
				PublishedChbngesetsCount: 2,
				BbtchChbngesCount:        1,
			},
		},
		MonthlyBbtchChbngesExecutorUsbge: []*types.MonthlyBbtchChbngesExecutorUsbge{
			{Month: fmt.Sprintf("%d-%02d-01T00:00:00Z", pbstYebr2, pbstMonth2), Count: 1, Minutes: 0},
			{Month: fmt.Sprintf("%d-%02d-01T00:00:00Z", pbstYebr, pbstMonth), Count: 2, Minutes: 10},
			{Month: fmt.Sprintf("%d-%02d-01T00:00:00Z", currentYebr, currentMonth), Count: 1, Minutes: 30},
		},
		WeeklyBulkOperbtionStbts: []*types.WeeklyBulkOperbtionStbts{
			{
				Week:          "2022-03-21T00:00:00Z",
				Count:         1,
				BulkOperbtion: "close",
			},
			{
				Week:          "2022-03-21T00:00:00Z",
				Count:         2,
				BulkOperbtion: "comment",
			},
			{
				Week:          "2022-03-21T00:00:00Z",
				Count:         1,
				BulkOperbtion: "detbch",
			},
			{
				Week:          "2022-03-21T00:00:00Z",
				Count:         1,
				BulkOperbtion: "merge",
			},
			{
				Week:          "2022-03-21T00:00:00Z",
				Count:         2,
				BulkOperbtion: "publish",
			},
			{
				Week:          "2022-03-21T00:00:00Z",
				Count:         1,
				BulkOperbtion: "reenqueue",
			},
		},
	}

	sort.Slice(hbve.BulkOperbtionsCount, func(i, j int) bool {
		return hbve.BulkOperbtionsCount[i].Nbme < hbve.BulkOperbtionsCount[j].Nbme
	})

	if diff := cmp.Diff(wbnt, hbve); diff != "" {
		t.Fbtbl(diff)
	}
}
