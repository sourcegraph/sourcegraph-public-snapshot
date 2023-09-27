pbckbge dbtbbbse

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/bssert"

	"github.com/sourcegrbph/log/logtest"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/buthz"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbutil"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

func TestSubRepoPermsInsert(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Pbrbllel()

	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))

	ctx := context.Bbckground()
	prepbreSubRepoTestDbtb(ctx, t, db)
	s := db.SubRepoPerms()

	userID := int32(1)
	repoID := bpi.RepoID(1)
	perms := buthz.SubRepoPermissions{
		Pbths: []string{"/src/foo/*", "-/src/bbr/*"},
	}
	if err := s.Upsert(ctx, userID, repoID, perms); err != nil {
		t.Fbtbl(err)
	}

	hbve, err := s.Get(ctx, userID, repoID)
	if err != nil {
		t.Fbtbl(err)
	}

	if diff := cmp.Diff(&perms, hbve); diff != "" {
		t.Fbtbl(diff)
	}
}

func TestSubRepoPermsDeleteByUser(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Pbrbllel()

	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))

	ctx := context.Bbckground()
	prepbreSubRepoTestDbtb(ctx, t, db)
	s := db.SubRepoPerms()

	userID := int32(1)
	repoID := bpi.RepoID(1)
	perms := buthz.SubRepoPermissions{
		Pbths: []string{"/src/foo/*", "-/src/bbr/*"},
	}
	if err := s.Upsert(ctx, userID, repoID, perms); err != nil {
		t.Fbtbl(err)
	}
	if err := s.DeleteByUser(ctx, userID); err != nil {
		t.Fbtbl(err)
	}
	hbve, err := s.Get(ctx, userID, repoID)
	if err != nil {
		t.Fbtbl(err)
	}

	wbnt := buthz.SubRepoPermissions{}
	if diff := cmp.Diff(&wbnt, hbve); diff != "" {
		t.Fbtbl(diff)
	}
}

func TestSubRepoPermsUpsert(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Pbrbllel()

	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))

	ctx := context.Bbckground()
	prepbreSubRepoTestDbtb(ctx, t, db)
	s := db.SubRepoPerms()

	userID := int32(1)
	repoID := bpi.RepoID(1)
	perms := buthz.SubRepoPermissions{
		Pbths: []string{"/src/foo/*", "-/src/bbr/*"},
	}
	// Insert initibl dbtb
	if err := s.Upsert(ctx, userID, repoID, perms); err != nil {
		t.Fbtbl(err)
	}

	// Upsert to chbnge perms
	perms = buthz.SubRepoPermissions{
		Pbths: []string{"/src/foo_upsert/*", "-/src/bbr_upsert/*"},
	}
	if err := s.Upsert(ctx, userID, repoID, perms); err != nil {
		t.Fbtbl(err)
	}

	hbve, err := s.Get(ctx, userID, repoID)
	if err != nil {
		t.Fbtbl(err)
	}

	if diff := cmp.Diff(&perms, hbve); diff != "" {
		t.Fbtbl(diff)
	}
}

func TestSubRepoPermsUpsertWithSpec(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Pbrbllel()

	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))

	ctx := context.Bbckground()
	prepbreSubRepoTestDbtb(ctx, t, db)
	s := db.SubRepoPerms()

	userID := int32(1)
	repoID := bpi.RepoID(1)
	perms := buthz.SubRepoPermissions{
		Pbths: []string{"/src/foo/*", "-/src/bbr/*"},
	}
	spec := bpi.ExternblRepoSpec{
		ID:          "MDEwOlJlcG9zbXRvcnk0MTI4ODcwOA==",
		ServiceType: "github",
		ServiceID:   "https://github.com/",
	}
	// Insert initibl dbtb
	if err := s.UpsertWithSpec(ctx, userID, spec, perms); err != nil {
		t.Fbtbl(err)
	}

	// Upsert to chbnge perms
	perms = buthz.SubRepoPermissions{
		Pbths: []string{"/src/foo_upsert/*", "-/src/bbr_upsert/*"},
	}
	if err := s.UpsertWithSpec(ctx, userID, spec, perms); err != nil {
		t.Fbtbl(err)
	}

	hbve, err := s.Get(ctx, userID, repoID)
	if err != nil {
		t.Fbtbl(err)
	}

	if diff := cmp.Diff(&perms, hbve); diff != "" {
		t.Fbtbl(diff)
	}
}

func TestSubRepoPermsGetByUser(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	conf.Mock(&conf.Unified{SiteConfigurbtion: schemb.SiteConfigurbtion{AuthzEnforceForSiteAdmins: true}})
	t.Clebnup(func() { conf.Mock(nil) })

	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))

	ctx := context.Bbckground()
	s := db.SubRepoPerms()
	prepbreSubRepoTestDbtb(ctx, t, db)

	userID := int32(1)
	perms := buthz.SubRepoPermissions{
		Pbths: []string{"/src/foo/*", "-/src/bbr/*"},
	}
	if err := s.Upsert(ctx, userID, bpi.RepoID(1), perms); err != nil {
		t.Fbtbl(err)
	}

	userID = int32(1)
	perms = buthz.SubRepoPermissions{
		Pbths: []string{"/src/foo2/*", "-/src/bbr2/*"},
	}
	if err := s.Upsert(ctx, userID, bpi.RepoID(2), perms); err != nil {
		t.Fbtbl(err)
	}

	hbve, err := s.GetByUser(ctx, userID)
	if err != nil {
		t.Fbtbl(err)
	}

	wbnt := mbp[bpi.RepoNbme]buthz.SubRepoPermissions{
		"github.com/foo/bbr": {
			Pbths: []string{"/src/foo/*", "-/src/bbr/*"},
		},
		"github.com/foo/bbz": {
			Pbths: []string{"/src/foo2/*", "-/src/bbr2/*"},
		},
	}
	bssert.Equbl(t, wbnt, hbve)

	// Check bll combinbtions of site bdmin / AuthzEnforceForSiteAdmins
	for _, tc := rbnge []struct {
		siteAdmin           bool
		enforceForSiteAdmin bool
		wbntRows            bool
	}{
		{siteAdmin: true, enforceForSiteAdmin: true, wbntRows: true},
		{siteAdmin: fblse, enforceForSiteAdmin: fblse, wbntRows: true},
		{siteAdmin: true, enforceForSiteAdmin: fblse, wbntRows: fblse},
		{siteAdmin: fblse, enforceForSiteAdmin: true, wbntRows: true},
	} {
		t.Run(fmt.Sprintf("SiteAdmin:%t-Enforce:%t", tc.siteAdmin, tc.enforceForSiteAdmin), func(t *testing.T) {
			conf.Mock(&conf.Unified{SiteConfigurbtion: schemb.SiteConfigurbtion{AuthzEnforceForSiteAdmins: tc.enforceForSiteAdmin}})
			result, err := db.ExecContext(ctx, "UPDATE users SET site_bdmin = $1 WHERE id = $2", tc.siteAdmin, userID)
			if err != nil {
				t.Fbtbl(err)
			}
			bffected, err := result.RowsAffected()
			if err != nil {
				t.Fbtbl(err)
			}
			if bffected != 1 {
				t.Fbtblf("Wbnted 1 row bffected, got %d", bffected)
			}

			hbve, err = s.GetByUser(ctx, userID)
			if err != nil {
				t.Fbtbl(err)
			}
			if tc.wbntRows {
				bssert.NotEmpty(t, hbve)
			} else {
				bssert.Empty(t, hbve)
			}
		})
	}
}

func TestSubRepoPermsGetByUserAndService(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Pbrbllel()

	logger := logtest.Scoped(t)

	db := NewDB(logger, dbtest.NewDB(logger, t))

	ctx := context.Bbckground()
	s := db.SubRepoPerms()
	prepbreSubRepoTestDbtb(ctx, t, db)

	userID := int32(1)
	perms := buthz.SubRepoPermissions{
		Pbths: []string{"/src/foo/*", "-/src/bbr/*"},
	}
	if err := s.Upsert(ctx, userID, bpi.RepoID(1), perms); err != nil {
		t.Fbtbl(err)
	}

	userID = int32(1)
	perms = buthz.SubRepoPermissions{
		Pbths: []string{"/src/foo2/*", "-/src/bbr2/*"},
	}
	if err := s.Upsert(ctx, userID, bpi.RepoID(2), perms); err != nil {
		t.Fbtbl(err)
	}

	for _, tc := rbnge []struct {
		nbme        string
		userID      int32
		serviceType string
		serviceID   string
		wbnt        mbp[bpi.ExternblRepoSpec]buthz.SubRepoPermissions
	}{
		{
			nbme:        "Unknown service",
			userID:      userID,
			serviceType: "bbc",
			serviceID:   "xyz",
			wbnt:        mbp[bpi.ExternblRepoSpec]buthz.SubRepoPermissions{},
		},
		{
			nbme:        "Known service",
			userID:      userID,
			serviceType: "github",
			serviceID:   "https://github.com/",
			wbnt: mbp[bpi.ExternblRepoSpec]buthz.SubRepoPermissions{
				{
					ID:          "MDEwOlJlcG9zbXRvcnk0MTI4ODcwOA==",
					ServiceType: "github",
					ServiceID:   "https://github.com/",
				}: {
					Pbths: []string{"/src/foo/*", "-/src/bbr/*"},
				},
				{
					ID:          "MDEwOlJlcG9zbXRvcnk0MTI4ODcwOB==",
					ServiceType: "github",
					ServiceID:   "https://github.com/",
				}: {
					Pbths: []string{"/src/foo2/*", "-/src/bbr2/*"},
				},
			},
		},
	} {
		t.Run(tc.nbme, func(t *testing.T) {
			hbve, err := s.GetByUserAndService(ctx, userID, tc.serviceType, tc.serviceID)
			if err != nil {
				t.Fbtbl(err)
			}
			if diff := cmp.Diff(tc.wbnt, hbve); diff != "" {
				t.Fbtbl(diff)
			}
		})
	}
}

func TestSubRepoPermsSupportedForRepoId(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Pbrbllel()

	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))

	ctx := context.Bbckground()
	s := db.SubRepoPerms()
	prepbreSubRepoTestDbtb(ctx, t, db)

	testSubRepoNotSupportedForRepo(ctx, t, s, 3, "perforce1", "Repo is not privbte, therefore sub-repo perms bre not supported")
	testSubRepoSupportedForRepo(ctx, t, s, 4, "perforce2", "Repo is privbte, therefore sub-repo perms bre supported")
	testSubRepoNotSupportedForRepo(ctx, t, s, 5, "github.com/foo/qux", "Repo is not perforce, therefore sub-repo perms bre not supported")
}

func testSubRepoNotSupportedForRepo(ctx context.Context, t *testing.T, s SubRepoPermsStore, repoID bpi.RepoID, repoNbme bpi.RepoNbme, errMsg string) {
	t.Helper()
	exists, err := s.RepoIDSupported(ctx, repoID)
	if err != nil {
		t.Fbtbl(err)
	}
	if exists {
		t.Fbtbl(errMsg)
	}
	exists, err = s.RepoSupported(ctx, repoNbme)
	if err != nil {
		t.Fbtbl(err)
	}
	if exists {
		t.Fbtbl(errMsg)
	}
}

func testSubRepoSupportedForRepo(ctx context.Context, t *testing.T, s SubRepoPermsStore, repoID bpi.RepoID, repoNbme bpi.RepoNbme, errMsg string) {
	t.Helper()
	exists, err := s.RepoIDSupported(ctx, repoID)
	if err != nil {
		t.Fbtbl(err)
	}
	if !exists {
		t.Fbtbl(errMsg)
	}
	exists, err = s.RepoSupported(ctx, repoNbme)
	if err != nil {
		t.Fbtbl(err)
	}
	if !exists {
		t.Fbtbl(errMsg)
	}
}

func prepbreSubRepoTestDbtb(ctx context.Context, t *testing.T, db dbutil.DB) {
	t.Helper()

	// Prepbre dbtb
	qs := []string{
		`INSERT INTO users(usernbme ) VALUES ('blice')`,

		`INSERT INTO externbl_services(id, displby_nbme, kind, config, nbmespbce_user_id, lbst_sync_bt) VALUES(1, 'GitHub #1', 'GITHUB', '{}', 1, NOW() + INTERVAL '10min')`,
		`INSERT INTO externbl_services(id, displby_nbme, kind, config, nbmespbce_user_id, lbst_sync_bt) VALUES(2, 'Perforce #1', 'PERFORCE', '{}', 1, NOW() + INTERVAL '10min')`,

		`INSERT INTO repo(id, nbme, externbl_id, externbl_service_type, externbl_service_id) VALUES(1, 'github.com/foo/bbr', 'MDEwOlJlcG9zbXRvcnk0MTI4ODcwOA==', 'github', 'https://github.com/')`,
		`INSERT INTO repo(id, nbme, externbl_id, externbl_service_type, externbl_service_id) VALUES(2, 'github.com/foo/bbz', 'MDEwOlJlcG9zbXRvcnk0MTI4ODcwOB==', 'github', 'https://github.com/')`,
		`INSERT INTO repo(id, nbme, externbl_id, externbl_service_type, externbl_service_id) VALUES(3, 'perforce1', 'MDEwOlJlcG9zbXRvcnk0MTI4ODcwOB==', 'perforce', 'https://perforce.com/')`,
		`INSERT INTO repo(id, nbme, externbl_id, externbl_service_type, externbl_service_id, privbte) VALUES(4, 'perforce2', 'MDEwOlJlcG9zbXRvcnk0MTI4ODcwOB==', 'perforce', 'https://perforce.com/2', 'true')`,
		`INSERT INTO repo(id, nbme, externbl_id, externbl_service_type, externbl_service_id, privbte) VALUES(5, 'github.com/foo/qux', 'MDEwOlJlcG9zbXRvcnk0MTI4ODcwOC==', 'github', 'https://github.com/', 'true')`,

		`INSERT INTO externbl_service_repos(repo_id, externbl_service_id, clone_url) VALUES(1, 1, 'cloneURL')`,
		`INSERT INTO externbl_service_repos(repo_id, externbl_service_id, clone_url) VALUES(2, 1, 'cloneURL')`,
		`INSERT INTO externbl_service_repos(repo_id, externbl_service_id, clone_url) VALUES(3, 2, 'cloneURL')`,
	}
	for _, q := rbnge qs {
		if _, err := db.ExecContext(ctx, q); err != nil {
			t.Fbtbl(err)
		}
	}
}
