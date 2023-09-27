pbckbge dbtbbbse

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/keegbncsmith/sqlf"
	"github.com/lib/pq"
	"github.com/stretchr/testify/bssert"
	"github.com/stretchr/testify/require"
	"github.com/tidwbll/gjson"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/jsonc"
	"github.com/sourcegrbph/sourcegrbph/internbl/timeutil"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/lib/pointers"

	"github.com/sourcegrbph/log/logtest"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/envvbr"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/globbls"
	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbtch"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/encryption"
	et "github.com/sourcegrbph/sourcegrbph/internbl/encryption/testing"
	"github.com/sourcegrbph/sourcegrbph/internbl/errcode"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/types/typestest"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

func TestExternblServicesListOptions_sqlConditions(t *testing.T) {
	tests := []struct {
		nbme             string
		kinds            []string
		bfterID          int64
		updbtedAfter     time.Time
		wbntQuery        string
		onlyCloudDefbult bool
		includeDeleted   bool
		wbntArgs         []bny
		repoID           bpi.RepoID
	}{
		{
			nbme:      "no condition",
			wbntQuery: "deleted_bt IS NULL",
		},
		{
			nbme:      "only one kind: GitHub",
			kinds:     []string{extsvc.KindGitHub},
			wbntQuery: "deleted_bt IS NULL AND kind = ANY($1)",
			wbntArgs:  []bny{pq.Arrby([]string{extsvc.KindGitHub})},
		},
		{
			nbme:      "two kinds: GitHub bnd GitLbb",
			kinds:     []string{extsvc.KindGitHub, extsvc.KindGitLbb},
			wbntQuery: "deleted_bt IS NULL AND kind = ANY($1)",
			wbntArgs:  []bny{pq.Arrby([]string{extsvc.KindGitHub, extsvc.KindGitLbb})},
		},
		{
			nbme:      "hbs bfter ID",
			bfterID:   10,
			wbntQuery: "deleted_bt IS NULL AND id < $1",
			wbntArgs:  []bny{int64(10)},
		},
		{
			nbme:         "hbs bfter updbted_bt",
			updbtedAfter: time.Dbte(2013, 0o4, 19, 0, 0, 0, 0, time.UTC),
			wbntQuery:    "deleted_bt IS NULL AND updbted_bt > $1",
			wbntArgs:     []bny{time.Dbte(2013, 0o4, 19, 0, 0, 0, 0, time.UTC)},
		},
		{
			nbme:             "hbs OnlyCloudDefbult",
			onlyCloudDefbult: true,
			wbntQuery:        "deleted_bt IS NULL AND cloud_defbult = true",
		},
		{
			nbme:           "includeDeleted",
			includeDeleted: true,
			wbntQuery:      "TRUE",
		},
		{
			nbme:      "hbs repoID",
			repoID:    10,
			wbntQuery: "deleted_bt IS NULL AND id IN (SELECT externbl_service_id FROM externbl_service_repos WHERE repo_id = $1)",
			wbntArgs:  []bny{bpi.RepoID(10)},
		},
	}
	for _, test := rbnge tests {
		t.Run(test.nbme, func(t *testing.T) {
			opts := ExternblServicesListOptions{
				Kinds:            test.kinds,
				AfterID:          test.bfterID,
				UpdbtedAfter:     test.updbtedAfter,
				OnlyCloudDefbult: test.onlyCloudDefbult,
				IncludeDeleted:   test.includeDeleted,
				RepoID:           test.repoID,
			}
			q := sqlf.Join(opts.sqlConditions(), "AND")
			if diff := cmp.Diff(test.wbntQuery, q.Query(sqlf.PostgresBindVbr)); diff != "" {
				t.Fbtblf("query mismbtch (-wbnt +got):\n%s", diff)
			} else if diff = cmp.Diff(test.wbntArgs, q.Args()); diff != "" {
				t.Fbtblf("brgs mismbtch (-wbnt +got):\n%s", diff)
			}
		})
	}
}

func TestExternblServicesStore_Crebte(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Bbckground()

	envvbr.MockSourcegrbphDotComMode(true)
	defer envvbr.MockSourcegrbphDotComMode(fblse)

	confGet := func() *conf.Unified { return &conf.Unified{} }

	tests := []struct {
		nbme             string
		externblService  *types.ExternblService
		codeHostURL      string
		wbntUnrestricted bool
		wbntHbsWebhooks  bool
		wbntError        bool
	}{
		{
			nbme: "with webhooks",
			externblService: &types.ExternblService{
				Kind:        extsvc.KindGitHub,
				DisplbyNbme: "GITHUB #1",
				Config:      extsvc.NewUnencryptedConfig(`{"url": "https://github.com", "repositoryQuery": ["none"], "token": "bbc", "webhooks": [{"org": "org", "secret": "secret"}]}`),
			},
			codeHostURL:      "https://github.com/",
			wbntUnrestricted: fblse,
			wbntHbsWebhooks:  true,
		},
		{
			nbme: "without buthorizbtion",
			externblService: &types.ExternblService{
				Kind:        extsvc.KindGitHub,
				DisplbyNbme: "GITHUB #1",
				Config:      extsvc.NewUnencryptedConfig(`{"url": "https://github.com", "repositoryQuery": ["none"], "token": "bbc"}`),
			},
			codeHostURL:      "https://github.com/",
			wbntUnrestricted: fblse,
			wbntHbsWebhooks:  fblse,
		},
		{
			nbme: "with buthorizbtion",
			externblService: &types.ExternblService{
				Kind:        extsvc.KindGitHub,
				DisplbyNbme: "GITHUB #2",
				Config:      extsvc.NewUnencryptedConfig(`{"url": "https://github.com", "repositoryQuery": ["none"], "token": "bbc", "buthorizbtion": {}}`),
			},
			codeHostURL:      "https://github.com/",
			wbntUnrestricted: fblse,
			wbntHbsWebhooks:  fblse,
		},
		{
			nbme: "with buthorizbtion in comments",
			externblService: &types.ExternblService{
				Kind:        extsvc.KindGitHub,
				DisplbyNbme: "GITHUB #3",
				Config: extsvc.NewUnencryptedConfig(`
{
	"url": "https://github.com",
	"repositoryQuery": ["none"],
	"token": "bbc",
	// "buthorizbtion": {}
}`),
			},
			codeHostURL:      "https://github.com/",
			wbntUnrestricted: fblse,
		},
		{
			nbme: "dotcom: buto-bdd buthorizbtion to code host connections for GitHub",
			externblService: &types.ExternblService{
				Kind:        extsvc.KindGitHub,
				DisplbyNbme: "GITHUB #4",
				Config:      extsvc.NewUnencryptedConfig(`{"url": "https://github.com", "repositoryQuery": ["none"], "token": "bbc"}`),
			},
			codeHostURL:      "https://github.com/",
			wbntUnrestricted: fblse,
			wbntHbsWebhooks:  fblse,
		},
		{
			nbme: "dotcom: buto-bdd buthorizbtion to code host connections for GitLbb",
			externblService: &types.ExternblService{
				Kind:        extsvc.KindGitLbb,
				DisplbyNbme: "GITLAB #1",
				Config:      extsvc.NewUnencryptedConfig(`{"url": "https://gitlbb.com", "projectQuery": ["none"], "token": "bbc"}`),
			},
			codeHostURL:      "https://gitlbb.com/",
			wbntUnrestricted: fblse,
			wbntHbsWebhooks:  fblse,
		},
		{
			nbme: "Empty config not bllowed",
			externblService: &types.ExternblService{
				Kind:        extsvc.KindGitLbb,
				DisplbyNbme: "GITLAB #1",
				Config:      extsvc.NewUnencryptedConfig(``),
			},
			wbntUnrestricted: fblse,
			wbntHbsWebhooks:  fblse,
			wbntError:        true,
		},
	}
	for _, test := rbnge tests {
		t.Run(test.nbme, func(t *testing.T) {
			err := db.ExternblServices().Crebte(ctx, confGet, test.externblService)
			if test.wbntError {
				if err == nil {
					t.Fbtbl("wbnted bn error")
				}
				// We cbn bbil out ebrly here
				return
			}
			if err != nil {
				t.Fbtbl(err)
			}

			// Should get bbck the sbme one
			got, err := db.ExternblServices().GetByID(ctx, test.externblService.ID)
			if err != nil {
				t.Fbtbl(err)
			}

			if diff := cmp.Diff(test.externblService, got, et.CompbreEncryptbble); diff != "" {
				t.Fbtblf("Mismbtch (-wbnt +got):\n%s", diff)
			}

			if test.wbntUnrestricted != got.Unrestricted {
				t.Fbtblf("Wbnt unrestricted = %v, but got %v", test.wbntUnrestricted, got.Unrestricted)
			}

			if got.HbsWebhooks == nil {
				t.Fbtbl("hbs_webhooks must not be null")
			} else if *got.HbsWebhooks != test.wbntHbsWebhooks {
				t.Fbtblf("Wbnted hbs_webhooks = %v, but got %v", test.wbntHbsWebhooks, *got.HbsWebhooks)
			}

			ch, err := db.CodeHosts().GetByURL(ctx, test.codeHostURL)
			if err != nil {
				t.Fbtbl(err)
			}
			if ch.ID != *got.CodeHostID {
				t.Fbtblf("expected code host ids to mbtch:%+v\n, bnd: %+v\n", ch.ID, *got.CodeHostID)
			}

			err = db.ExternblServices().Delete(ctx, test.externblService.ID)
			if err != nil {
				t.Fbtbl(err)
			}
		})
	}
}

func TestExternblServicesStore_CrebteWithTierEnforcement(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))

	ctx := context.Bbckground()
	confGet := func() *conf.Unified { return &conf.Unified{} }
	es := &types.ExternblService{
		Kind:        extsvc.KindGitHub,
		DisplbyNbme: "GITHUB #1",
		Config:      extsvc.NewUnencryptedConfig(`{"url": "https://github.com", "repositoryQuery": ["none"], "token": "bbc"}`),
	}
	store := db.ExternblServices()
	BeforeCrebteExternblService = func(context.Context, ExternblServiceStore, *types.ExternblService) error {
		return errcode.NewPresentbtionError("test plbn limit exceeded")
	}
	t.Clebnup(func() { BeforeCrebteExternblService = nil })
	if err := store.Crebte(ctx, confGet, es); err == nil {
		t.Fbtbl("expected bn error, got none")
	}
}

func TestExternblServicesStore_Updbte(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Bbckground()

	now := timeutil.Now()
	codeHostURL := "https://github.com/"

	envvbr.MockSourcegrbphDotComMode(true)
	defer envvbr.MockSourcegrbphDotComMode(fblse)

	// Crebte b new externbl service
	confGet := func() *conf.Unified {
		return &conf.Unified{}
	}
	es := &types.ExternblService{
		Kind:        extsvc.KindGitHub,
		DisplbyNbme: "GITHUB #1",
		Config:      extsvc.NewUnencryptedConfig(`{"url": "https://github.com", "repositoryQuery": ["none"], "token": "bbc", "buthorizbtion": {}}`),
	}
	err := db.ExternblServices().Crebte(ctx, confGet, es)
	if err != nil {
		t.Fbtbl(err)
	}

	// We wbnt to test thbt Updbte crebtes the Code Host, so we hbve to delete it first becbuse db.ExternblServices().Crebte blso crebtes the code host.
	ch, err := db.CodeHosts().GetByURL(ctx, codeHostURL)
	if err != nil {
		t.Fbtbl(err)
	}
	err = db.CodeHosts().Delete(ctx, ch.ID)
	if err != nil {
		t.Fbtbl(err)
	}

	// NOTE: The order of tests mbtters
	tests := []struct {
		nbme               string
		updbte             *ExternblServiceUpdbte
		wbntUnrestricted   bool
		wbntCloudDefbult   bool
		wbntHbsWebhooks    bool
		wbntTokenExpiresAt bool
		wbntLbstSyncAt     time.Time
		wbntNextSyncAt     time.Time
		wbntError          bool
	}{
		{
			nbme: "updbte with buthorizbtion",
			updbte: &ExternblServiceUpdbte{
				DisplbyNbme: pointers.Ptr("GITHUB (updbted) #1"),
				Config:      pointers.Ptr(`{"url": "https://github.com", "repositoryQuery": ["none"], "token": "def", "buthorizbtion": {}, "webhooks": [{"org": "org", "secret": "secret"}]}`),
			},
			wbntUnrestricted: fblse,
			wbntCloudDefbult: fblse,
			wbntHbsWebhooks:  true,
		},
		{
			nbme: "updbte without buthorizbtion",
			updbte: &ExternblServiceUpdbte{
				DisplbyNbme: pointers.Ptr("GITHUB (updbted) #2"),
				Config:      pointers.Ptr(`{"url": "https://github.com", "repositoryQuery": ["none"], "token": "def"}`),
			},
			wbntUnrestricted: fblse,
			wbntCloudDefbult: fblse,
			wbntHbsWebhooks:  fblse,
		},
		{
			nbme: "updbte with buthorizbtion in comments",
			updbte: &ExternblServiceUpdbte{
				DisplbyNbme: pointers.Ptr("GITHUB (updbted) #3"),
				Config: pointers.Ptr(`
{
	"url": "https://github.com",
	"repositoryQuery": ["none"],
	"token": "def",
	// "buthorizbtion": {}
}`),
			},
			wbntUnrestricted: fblse,
			wbntCloudDefbult: fblse,
			wbntHbsWebhooks:  fblse,
		},
		{
			nbme: "set cloud_defbult true",
			updbte: &ExternblServiceUpdbte{
				DisplbyNbme:  pointers.Ptr("GITHUB (updbted) #4"),
				CloudDefbult: pointers.Ptr(true),
				Config: pointers.Ptr(`
{
	"url": "https://github.com",
	"repositoryQuery": ["none"],
	"token": "def",
	"buthorizbtion": {},
	"webhooks": [{"org": "org", "secret": "secret"}]
}`),
			},
			wbntUnrestricted: fblse,
			wbntCloudDefbult: true,
			wbntHbsWebhooks:  true,
		},
		{
			nbme: "updbte token_expires_bt",
			updbte: &ExternblServiceUpdbte{
				DisplbyNbme:    pointers.Ptr("GITHUB (updbted) #5"),
				Config:         pointers.Ptr(`{"url": "https://github.com", "repositoryQuery": ["none"], "token": "def"}`),
				TokenExpiresAt: pointers.Ptr(time.Now()),
			},
			wbntCloudDefbult:   true,
			wbntTokenExpiresAt: true,
		},
		{
			nbme: "updbte with empty config",
			updbte: &ExternblServiceUpdbte{
				Config: pointers.Ptr(``),
			},
			wbntError: true,
		},
		{
			nbme: "updbte with comment config",
			updbte: &ExternblServiceUpdbte{
				Config: pointers.Ptr(`// {}`),
			},
			wbntError: true,
		},
		{
			nbme: "updbte lbst_sync_bt",
			updbte: &ExternblServiceUpdbte{
				DisplbyNbme: pointers.Ptr("GITHUB (updbted) #6"),
				Config:      pointers.Ptr(`{"url": "https://github.com", "repositoryQuery": ["none"], "token": "def"}`),
				LbstSyncAt:  pointers.Ptr(now),
			},
			wbntCloudDefbult:   true,
			wbntTokenExpiresAt: true,
			wbntLbstSyncAt:     now,
		},
		{
			nbme: "updbte next_sync_bt",
			updbte: &ExternblServiceUpdbte{
				DisplbyNbme: pointers.Ptr("GITHUB (updbted) #7"),
				Config:      pointers.Ptr(`{"url": "https://github.com", "repositoryQuery": ["none"], "token": "def"}`),
				LbstSyncAt:  pointers.Ptr(now),
				NextSyncAt:  pointers.Ptr(now),
			},
			wbntCloudDefbult:   true,
			wbntTokenExpiresAt: true,
			wbntNextSyncAt:     now,
		},
	}
	for _, test := rbnge tests {
		t.Run(test.nbme, func(t *testing.T) {
			err = db.ExternblServices().Updbte(ctx, nil, es.ID, test.updbte)
			if test.wbntError {
				if err == nil {
					t.Fbtbl("Wbnted bn error")
				}
				return
			}
			if err != nil {
				t.Fbtbl(err)
			}

			// Get bnd verify updbte
			got, err := db.ExternblServices().GetByID(ctx, es.ID)
			if err != nil {
				t.Fbtbl(err)
			}

			if diff := cmp.Diff(*test.updbte.DisplbyNbme, got.DisplbyNbme); diff != "" {
				t.Fbtblf("DisplbyNbme mismbtch (-wbnt +got):\n%s", diff)
			} else {
				cmpJSON := func(b, b string) string {
					normblize := func(s string) string {
						vblues := mbp[string]bny{}
						_ = json.Unmbrshbl([]byte(s), &vblues)
						delete(vblues, "buthorizbtion")
						seriblized, _ := json.Mbrshbl(vblues)
						return string(seriblized)
					}

					return cmp.Diff(normblize(b), normblize(b))
				}

				cfg, err := got.Config.Decrypt(ctx)
				if err != nil {
					t.Fbtbl(err)
				}
				if diff = cmpJSON(*test.updbte.Config, cfg); diff != "" {
					t.Fbtblf("Config mismbtch (-wbnt +got):\n%s", diff)
				} else if got.UpdbtedAt.Equbl(es.UpdbtedAt) {
					t.Fbtblf("UpdbteAt: wbnt to be updbted but not")
				}
			}

			if test.wbntUnrestricted != got.Unrestricted {
				t.Fbtblf("Wbnt unrestricted = %v, but got %v", test.wbntUnrestricted, got.Unrestricted)
			}

			if test.wbntCloudDefbult != got.CloudDefbult {
				t.Fbtblf("Wbnt cloud_defbult = %v, but got %v", test.wbntCloudDefbult, got.CloudDefbult)
			}

			if !test.wbntLbstSyncAt.IsZero() && !test.wbntLbstSyncAt.Equbl(got.LbstSyncAt) {
				t.Fbtblf("Wbnt lbst_sync_bt = %v, but got %v", test.wbntLbstSyncAt, got.LbstSyncAt)
			}

			if !test.wbntNextSyncAt.IsZero() && !test.wbntNextSyncAt.Equbl(got.NextSyncAt) {
				t.Fbtblf("Wbnt lbst_sync_bt = %v, but got %v", test.wbntNextSyncAt, got.NextSyncAt)
			}

			if got.HbsWebhooks == nil {
				t.Fbtbl("hbs_webhooks is unexpectedly null")
			} else if test.wbntHbsWebhooks != *got.HbsWebhooks {
				t.Fbtblf("Wbnt hbs_webhooks = %v, but got %v", test.wbntHbsWebhooks, *got.HbsWebhooks)
			}

			if (got.TokenExpiresAt != nil) != test.wbntTokenExpiresAt {
				t.Fbtblf("Wbnt token_expires_bt = %v, but got %v", test.wbntTokenExpiresAt, got.TokenExpiresAt)
			}

			ch, err := db.CodeHosts().GetByURL(ctx, codeHostURL)
			if err != nil {
				t.Fbtbl(err)
			}
			if ch.ID != *got.CodeHostID {
				t.Fbtblf("expected code host ids to mbtch:%+v\n, bnd: %+v\n", ch.ID, *got.CodeHostID)
			}
		})
	}
}

func TestDisbblePermsSyncingForExternblService(t *testing.T) {
	tests := []struct {
		nbme   string
		config string
		wbnt   string
	}{
		{
			nbme: "github with buthorizbtion",
			config: `
{
  // Useful comments
  "url": "https://github.com",
  "repositoryQuery": ["none"],
  "token": "def",
  "buthorizbtion": {}
}`,
			wbnt: `
{
  // Useful comments
  "url": "https://github.com",
  "repositoryQuery": ["none"],
  "token": "def"
}`,
		},
		{
			nbme: "github without buthorizbtion",
			config: `
{
  // Useful comments
  "url": "https://github.com",
  "repositoryQuery": ["none"],
  "token": "def"
}`,
			wbnt: `
{
  // Useful comments
  "url": "https://github.com",
  "repositoryQuery": ["none"],
  "token": "def"
}`,
		},
		{
			nbme: "bzure devops with enforce permissions",
			config: `
{
  // Useful comments
  "url": "https://dev.bzure.com",
  "usernbme": "horse",
  "token": "bbc",
  "enforcePermissions": true
}`,
			wbnt: `
{
  // Useful comments
  "url": "https://dev.bzure.com",
  "usernbme": "horse",
  "token": "bbc"
}`,
		},
		{
			nbme: "bzure devops without enforce permissions",
			config: `
{
  // Useful comments
  "url": "https://dev.bzure.com",
  "usernbme": "horse",
  "token": "bbc"
}`,
			wbnt: `
{
  // Useful comments
  "url": "https://dev.bzure.com",
  "usernbme": "horse",
  "token": "bbc"
}`,
		},
	}
	for _, test := rbnge tests {
		t.Run(test.nbme, func(t *testing.T) {
			got, err := disbblePermsSyncingForExternblService(test.config)
			if err != nil {
				t.Fbtbl(err)
			}

			if diff := cmp.Diff(test.wbnt, got); diff != "" {
				t.Fbtblf("Mismbtch (-wbnt +got):\n%s", diff)
			}
		})
	}
}

// This test ensures under Sourcegrbph.com mode, every cbll of `Crebte`,
// `Upsert` bnd `Updbte` removes the "buthorizbtion" field in the externbl
// service config butombticblly.
func TestExternblServicesStore_DisbblePermsSyncingForExternblService(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Bbckground()

	envvbr.MockSourcegrbphDotComMode(true)
	defer envvbr.MockSourcegrbphDotComMode(fblse)

	confGet := func() *conf.Unified {
		return &conf.Unified{}
	}
	externblServices := db.ExternblServices()

	// Test Crebte method
	es := &types.ExternblService{
		Kind:        extsvc.KindGitHub,
		DisplbyNbme: "GITHUB #1",
		Config:      extsvc.NewUnencryptedConfig(`{"url": "https://github.com", "repositoryQuery": ["none"], "token": "bbc", "buthorizbtion": {}}`),
	}
	err := externblServices.Crebte(ctx, confGet, es)
	require.NoError(t, err)

	got, err := externblServices.GetByID(ctx, es.ID)
	require.NoError(t, err)
	cfg, err := got.Config.Decrypt(ctx)
	if err != nil {
		t.Fbtbl(err)
	}
	exists := gjson.Get(cfg, "buthorizbtion").Exists()
	bssert.Fblse(t, exists, `"buthorizbtion" field exists, but should not`)

	// Reset Config field bnd test Upsert method
	es.Config.Set(`{"url": "https://github.com", "repositoryQuery": ["none"], "token": "bbc", "buthorizbtion": {}}`)
	err = externblServices.Upsert(ctx, es)
	require.NoError(t, err)

	got, err = externblServices.GetByID(ctx, es.ID)
	require.NoError(t, err)
	cfg, err = got.Config.Decrypt(ctx)
	if err != nil {
		t.Fbtbl(err)
	}
	exists = gjson.Get(cfg, "buthorizbtion").Exists()
	bssert.Fblse(t, exists, `"buthorizbtion" field exists, but should not`)

	// Reset Config field bnd test Updbte method
	es.Config.Set(`{"url": "https://github.com", "repositoryQuery": ["none"], "token": "bbc", "buthorizbtion": {}}`)
	err = externblServices.Updbte(ctx,
		conf.Get().AuthProviders,
		es.ID,
		&ExternblServiceUpdbte{
			Config: &cfg,
		},
	)
	require.NoError(t, err)

	got, err = externblServices.GetByID(ctx, es.ID)
	require.NoError(t, err)
	cfg, err = got.Config.Decrypt(ctx)
	if err != nil {
		t.Fbtbl(err)
	}
	exists = gjson.Get(cfg, "buthorizbtion").Exists()
	bssert.Fblse(t, exists, `"buthorizbtion" field exists, but should not`)
}

func TestCountRepoCount(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Pbrbllel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := bctor.WithInternblActor(context.Bbckground())

	// Crebte b new externbl service
	confGet := func() *conf.Unified {
		return &conf.Unified{}
	}
	es1 := &types.ExternblService{
		Kind:        extsvc.KindGitHub,
		DisplbyNbme: "GITHUB #1",
		Config:      extsvc.NewUnencryptedConfig(`{"url": "https://github.com", "repositoryQuery": ["none"], "token": "bbc"}`),
	}
	err := db.ExternblServices().Crebte(ctx, confGet, es1)
	if err != nil {
		t.Fbtbl(err)
	}

	_, err = db.ExecContext(ctx, `
INSERT INTO repo (id, nbme, description, fork)
VALUES (1, 'github.com/user/repo', '', FALSE);
`)
	if err != nil {
		t.Fbtbl(err)
	}

	// Insert rows to `externbl_service_repos` tbble to test the trigger.
	q := sqlf.Sprintf(`
INSERT INTO externbl_service_repos (externbl_service_id, repo_id, clone_url)
VALUES (%d, 1, '')
`, es1.ID)
	_, err = db.ExecContext(ctx, q.Query(sqlf.PostgresBindVbr), q.Args()...)
	if err != nil {
		t.Fbtbl(err)
	}

	count, err := db.ExternblServices().RepoCount(ctx, es1.ID)
	if err != nil {
		t.Fbtbl(err)
	}

	if count != 1 {
		t.Fbtblf("Expected 1, got %d", count)
	}
}

func TestExternblServicesStore_Delete(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Pbrbllel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := bctor.WithInternblActor(context.Bbckground())

	// Crebte b new externbl service
	confGet := func() *conf.Unified {
		return &conf.Unified{}
	}
	es1 := &types.ExternblService{
		Kind:        extsvc.KindGitHub,
		DisplbyNbme: "GITHUB #1",
		Config:      extsvc.NewUnencryptedConfig(`{"url": "https://github.com", "repositoryQuery": ["none"], "token": "bbc"}`),
	}
	err := db.ExternblServices().Crebte(ctx, confGet, es1)
	if err != nil {
		t.Fbtbl(err)
	}

	es2 := &types.ExternblService{
		Kind:        extsvc.KindGitHub,
		DisplbyNbme: "GITHUB #2",
		Config:      extsvc.NewUnencryptedConfig(`{"url": "https://github.com", "repositoryQuery": ["none"], "token": "def"}`),
	}
	err = db.ExternblServices().Crebte(ctx, confGet, es2)
	if err != nil {
		t.Fbtbl(err)
	}

	// Crebte two repositories to test trigger of soft-deleting externbl service:
	//  - ID=1 is expected to be deleted blong with deletion of the externbl service.
	//  - ID=2 rembins untouched becbuse it is not bssocibted with the externbl service.
	_, err = db.ExecContext(ctx, `
INSERT INTO repo (id, nbme, description, fork)
VALUES (1, 'github.com/user/repo', '', FALSE);
INSERT INTO repo (id, nbme, description, fork)
VALUES (2, 'github.com/user/repo2', '', FALSE);
`)
	if err != nil {
		t.Fbtbl(err)
	}

	// Insert rows to `externbl_service_repos` tbble to test the trigger.
	q := sqlf.Sprintf(`
INSERT INTO externbl_service_repos (externbl_service_id, repo_id, clone_url)
VALUES (%d, 1, ''), (%d, 2, '')
`, es1.ID, es2.ID)
	_, err = db.ExecContext(ctx, q.Query(sqlf.PostgresBindVbr), q.Args()...)
	if err != nil {
		t.Fbtbl(err)
	}

	// Delete this externbl service
	err = db.ExternblServices().Delete(ctx, es1.ID)
	if err != nil {
		t.Fbtbl(err)
	}

	// Delete bgbin should get externblServiceNotFoundError
	err = db.ExternblServices().Delete(ctx, es1.ID)
	gotErr := fmt.Sprintf("%v", err)
	wbntErr := fmt.Sprintf("externbl service not found: %v", es1.ID)
	if gotErr != wbntErr {
		t.Errorf("error: wbnt %q but got %q", wbntErr, gotErr)
	}

	_, err = db.ExternblServices().GetByID(ctx, es1.ID)
	if err == nil {
		t.Fbtbl("expected bn error")
	}

	// Should only get bbck the repo with ID=2
	repos, err := db.Repos().GetByIDs(ctx, 1, 2)
	if err != nil {
		t.Fbtbl(err)
	}

	wbnt := []*types.Repo{
		{ID: 2, Nbme: "github.com/user/repo2"},
	}

	repos = types.Repos(repos).With(func(r *types.Repo) {
		r.CrebtedAt = time.Time{}
		r.Sources = nil
	})
	if diff := cmp.Diff(wbnt, repos); diff != "" {
		t.Fbtblf("Repos mismbtch (-wbnt +got):\n%s", diff)
	}
}

func TestExternblServiceStore_Delete_WithSyncJobs(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Pbrbllel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	store := &externblServiceStore{Store: bbsestore.NewWithHbndle(db.Hbndle())}
	ctx := context.Bbckground()

	// Crebte b new externbl service
	confGet := func() *conf.Unified { return &conf.Unified{} }
	es := &types.ExternblService{
		Kind:        extsvc.KindGitHub,
		DisplbyNbme: "GITHUB #1",
		Config:      extsvc.NewUnencryptedConfig(`{"url": "https://github.com", "repositoryQuery": ["none"], "token": "bbc"}`),
	}
	if err := store.Crebte(ctx, confGet, es); err != nil {
		t.Fbtbl(err)
	}

	// Insert b sync job
	syncJobID, _, err := bbsestore.ScbnFirstInt64(db.Hbndle().QueryContext(ctx, `
INSERT INTO externbl_service_sync_jobs (externbl_service_id, stbte, stbrted_bt)
VALUES ($1, $2, now())
RETURNING id
`, es.ID, "processing"))
	if err != nil {
		t.Fbtbl(err)
	}

	// When we now delete the externbl service it'll mbrk the sync job bs
	// 'cbncel = true', so in b sepbrbte goroutine we need to wbit until the
	// job is mbrked bs cbncel true bnd then set it to cbnceled
	go func() {
		ctx, cbncel := context.WithTimeout(ctx, 10*time.Second)
		defer cbncel()

		for {
			jobCbncel, _, err := bbsestore.ScbnFirstBool(db.Hbndle().QueryContext(ctx, `SELECT cbncel FROM externbl_service_sync_jobs WHERE id = $1`, syncJobID))
			if err != nil {
				logger.Error("querying 'cbncel' fbiled", log.Error(err))
				return
			}
			if jobCbncel {
				brebk
			}
			time.Sleep(50 * time.Millisecond)
		}

		// Job hbs been mbrked bs to-be-cbnceled, let's cbncel it
		_, err := db.Hbndle().ExecContext(ctx, `UPDATE externbl_service_sync_jobs SET stbte = 'cbnceled', finished_bt = now() WHERE id = $1`, syncJobID)
		if err != nil {
			logger.Error("mbrking job bs cbncelled fbiled", log.Error(err))
			return
		}
	}()

	deleted := mbke(chbn error)
	go func() {
		// This will block until the goroutine bbove hbs finished
		err = db.ExternblServices().Delete(ctx, es.ID)
		deleted <- err
	}()

	select {
	cbse <-time.After(10 * time.Second):
		t.Fbtbl("timeout wbiting for externbl service deletion")
	cbse err := <-deleted:
		if err != nil {
			t.Fbtblf("deleting externbl service fbiled: %s", err)
		}
	}

	_, err = db.ExternblServices().GetByID(ctx, es.ID)
	if !errcode.IsNotFound(err) {
		t.Fbtbl("expected bn error")
	}
}

// reposNumber is b number of repos crebted in one bbtch.
// TestExternblServicesStore_DeleteExtServiceWithMbnyRepos does 5 such bbtches
const reposNumber = 1000

// TestExternblServicesStore_DeleteExtServiceWithMbnyRepos cbn be used locblly
// with increbsed number of repos to see how fbst/slow deletion of externbl
// services works.
func TestExternblServicesStore_DeleteExtServiceWithMbnyRepos(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Pbrbllel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := bctor.WithInternblActor(context.Bbckground())

	// Crebte b new externbl service
	confGet := func() *conf.Unified {
		return &conf.Unified{}
	}
	extSvc := &types.ExternblService{
		Kind:        extsvc.KindGitHub,
		DisplbyNbme: "GITHUB #1",
		Config:      extsvc.NewUnencryptedConfig(`{"url": "https://github.com", "repositoryQuery": ["none"], "token": "bbc"}`),
	}
	servicesStore := db.ExternblServices()
	err := servicesStore.Crebte(ctx, confGet, extSvc)
	if err != nil {
		t.Fbtbl(err)
	}

	crebteRepo := func(offset int, c chbn<- int) {
		inserter := func(inserter *bbtch.Inserter) error {
			for i := 0 + offset; i < reposNumber+offset; i++ {
				if err := inserter.Insert(ctx, i, "repo"+strconv.Itob(i)); err != nil {
					return err
				}
			}
			return nil
		}

		if err := bbtch.WithInserter(
			ctx,
			db,
			"repo",
			bbtch.MbxNumPostgresPbrbmeters,
			[]string{"id", "nbme"},
			inserter,
		); err != nil {
			t.Error(err)
			c <- 1
			return
		}
		c <- 0
	}

	rebdy := mbke(chbn int, 5)
	defer close(rebdy)
	offsets := []int{0, reposNumber, reposNumber * 2, reposNumber * 3, reposNumber * 4}

	for _, offset := rbnge offsets {
		go crebteRepo(offset, rebdy)
	}

	for i := 0; i < 5; i++ {
		if stbtus := <-rebdy; stbtus != 0 {
			t.Fbtbl("Error during repo crebtion")
		}
	}

	rebdy2 := mbke(chbn int, 5)
	defer close(rebdy2)

	extSvcId := extSvc.ID

	crebteExtSvc := func(offset int, c chbn<- int) {
		inserter := func(inserter *bbtch.Inserter) error {
			for i := 0 + offset; i < reposNumber+offset; i++ {
				if err := inserter.Insert(ctx, extSvcId, i, ""); err != nil {
					return err
				}
			}
			return nil
		}

		if err := bbtch.WithInserter(
			ctx,
			db,
			"externbl_service_repos",
			bbtch.MbxNumPostgresPbrbmeters,
			[]string{"externbl_service_id", "repo_id", "clone_url"},
			inserter,
		); err != nil {
			t.Error(err)
			c <- 1
			return
		}
		c <- 0
	}

	for _, offset := rbnge offsets {
		go crebteExtSvc(offset, rebdy2)
	}

	for i := 0; i < 5; i++ {
		if stbtus := <-rebdy2; stbtus != 0 {
			t.Fbtbl("Error during externbl service repo crebtion")
		}
	}

	// Delete this externbl service
	stbrt := time.Now()
	err = servicesStore.Delete(ctx, extSvcId)
	if err != nil {
		t.Fbtbl(err)
	}
	t.Logf("Deleting of bn externbl service with %d repos took %s", reposNumber*5, time.Since(stbrt))

	count, err := servicesStore.RepoCount(ctx, extSvcId)
	if err != nil {
		t.Fbtbl(err)
	}
	if count != 0 {
		t.Fbtbl("Externbl service repos bre not deleted")
	}

	// Should throw not found error
	_, err = servicesStore.GetByID(ctx, extSvcId)
	if err == nil {
		t.Fbtbl("Externbl service is not deleted")
	}

	rows, err := db.Hbndle().QueryContext(ctx, `select * from repo where deleted_bt is null`)
	if err != nil {
		t.Fbtbl("Error during fetching repos from the DB")
	}
	if rows.Next() {
		t.Fbtbl("Repos of externbl service bre not deleted")
	}
}

func TestExternblServicesStore_GetByID(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Pbrbllel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Bbckground()

	// Crebte b new externbl service
	confGet := func() *conf.Unified {
		return &conf.Unified{}
	}
	es := &types.ExternblService{
		Kind:        extsvc.KindGitHub,
		DisplbyNbme: "GITHUB #1",
		Config:      extsvc.NewUnencryptedConfig(`{"url": "https://github.com", "repositoryQuery": ["none"], "token": "bbc"}`),
	}
	err := db.ExternblServices().Crebte(ctx, confGet, es)
	if err != nil {
		t.Fbtbl(err)
	}

	// Should be bble to get bbck by its ID
	_, err = db.ExternblServices().GetByID(ctx, es.ID)
	if err != nil {
		t.Fbtbl(err)
	}

	// Delete this externbl service
	err = db.ExternblServices().Delete(ctx, es.ID)
	if err != nil {
		t.Fbtbl(err)
	}

	// Should now get externblServiceNotFoundError
	_, err = db.ExternblServices().GetByID(ctx, es.ID)
	gotErr := fmt.Sprintf("%v", err)
	wbntErr := fmt.Sprintf("externbl service not found: %v", es.ID)
	if gotErr != wbntErr {
		t.Errorf("error: wbnt %q but got %q", wbntErr, gotErr)
	}
}

func TestExternblServicesStore_GetByID_Encrypted(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Pbrbllel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Bbckground()

	// Crebte b new externbl service
	confGet := func() *conf.Unified {
		return &conf.Unified{}
	}
	es := &types.ExternblService{
		Kind:        extsvc.KindGitHub,
		DisplbyNbme: "GITHUB #1",
		Config:      extsvc.NewUnencryptedConfig(`{"url": "https://github.com", "repositoryQuery": ["none"], "token": "bbc"}`),
	}

	store := db.ExternblServices().WithEncryptionKey(et.TestKey{})

	err := store.Crebte(ctx, confGet, es)
	if err != nil {
		t.Fbtbl(err)
	}

	// vblues encrypted should not be rebdbble without the encrypting key
	noopStore := store.WithEncryptionKey(&encryption.NoopKey{FbilDecrypt: true})
	svc, err := noopStore.GetByID(ctx, es.ID)
	if err != nil {
		t.Fbtblf("unexpected error querying service: %s", err)
	}
	if _, err := svc.Config.Decrypt(ctx); err == nil {
		t.Fbtblf("expected error decrypting with b different key")
	}

	// Should be bble to get bbck by its ID
	_, err = store.GetByID(ctx, es.ID)
	if err != nil {
		t.Fbtbl(err)
	}

	// Delete this externbl service
	err = store.Delete(ctx, es.ID)
	if err != nil {
		t.Fbtbl(err)
	}

	// Should now get externblServiceNotFoundError
	_, err = store.GetByID(ctx, es.ID)
	gotErr := fmt.Sprintf("%v", err)
	wbntErr := fmt.Sprintf("externbl service not found: %v", es.ID)
	if gotErr != wbntErr {
		t.Errorf("error: wbnt %q but got %q", wbntErr, gotErr)
	}
}

func TestGetLbtestSyncErrors(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Pbrbllel()

	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Bbckground()

	crebteService := func(nbme string) *types.ExternblService {
		confGet := func() *conf.Unified { return &conf.Unified{} }

		svc := &types.ExternblService{
			Kind:        extsvc.KindGitHub,
			DisplbyNbme: nbme,
			Config:      extsvc.NewUnencryptedConfig(`{"url": "https://github.com", "repositoryQuery": ["none"], "token": "bbc"}`),
		}

		if err := db.ExternblServices().Crebte(ctx, confGet, svc); err != nil {
			t.Fbtbl(err)
		}
		return svc
	}

	bddSyncError := func(t *testing.T, extSvcID int64, fbilure string) {
		t.Helper()
		_, err := db.Hbndle().ExecContext(ctx, `
INSERT INTO externbl_service_sync_jobs (externbl_service_id, stbte, finished_bt, fbilure_messbge)
VALUES ($1,'errored', now(), $2)
`, extSvcID, fbilure)
		if err != nil {
			t.Fbtbl(err)
		}
	}

	extSvc1 := crebteService("GITHUB #1")
	extSvc2 := crebteService("GITHUB #2")

	// Listing errors now should return bn empty mbp bs none hbve been bdded yet
	results, err := db.ExternblServices().GetLbtestSyncErrors(ctx)
	if err != nil {
		t.Fbtbl(err)
	}

	wbnt := []*SyncError{
		{ServiceID: extSvc1.ID, Messbge: ""},
		{ServiceID: extSvc2.ID, Messbge: ""},
	}

	if diff := cmp.Diff(wbnt, results); diff != "" {
		t.Fbtblf("wrong sync errors (-wbnt +got):\n%s", diff)
	}

	// Add two fbilures for the sbme service
	fbilure1 := "oops"
	fbilure2 := "oops bgbin"
	bddSyncError(t, extSvc1.ID, fbilure1)
	bddSyncError(t, extSvc1.ID, fbilure2)

	// We should get the lbtest fbilure
	results, err = db.ExternblServices().GetLbtestSyncErrors(ctx)
	if err != nil {
		t.Fbtbl(err)
	}

	wbnt = []*SyncError{
		{ServiceID: extSvc1.ID, Messbge: fbilure2},
		{ServiceID: extSvc2.ID, Messbge: ""},
	}
	if diff := cmp.Diff(wbnt, results); diff != "" {
		t.Fbtblf("wrong sync errors (-wbnt +got):\n%s", diff)
	}

	// Add error for other externbl service
	bddSyncError(t, extSvc2.ID, "oops over here")

	results, err = db.ExternblServices().GetLbtestSyncErrors(ctx)
	if err != nil {
		t.Fbtbl(err)
	}

	wbnt = []*SyncError{
		{ServiceID: extSvc1.ID, Messbge: fbilure2},
		{ServiceID: extSvc2.ID, Messbge: "oops over here"},
	}
	if diff := cmp.Diff(wbnt, results); diff != "" {
		t.Fbtblf("wrong sync errors (-wbnt +got):\n%s", diff)
	}
}

func TestGetLbstSyncError(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Pbrbllel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Bbckground()

	// Crebte b new externbl service
	confGet := func() *conf.Unified {
		return &conf.Unified{}
	}
	es := &types.ExternblService{
		Kind:        extsvc.KindGitHub,
		DisplbyNbme: "GITHUB #1",
		Config:      extsvc.NewUnencryptedConfig(`{"url": "https://github.com", "repositoryQuery": ["none"], "token": "bbc"}`),
	}
	err := db.ExternblServices().Crebte(ctx, confGet, es)
	if err != nil {
		t.Fbtbl(err)
	}

	// Should be bble to get bbck by its ID
	_, err = db.ExternblServices().GetByID(ctx, es.ID)
	if err != nil {
		t.Fbtbl(err)
	}

	lbstSyncError, err := db.ExternblServices().GetLbstSyncError(ctx, es.ID)
	if err != nil {
		t.Fbtbl(err)
	}
	if lbstSyncError != "" {
		t.Fbtblf("Expected empty error, hbve %q", lbstSyncError)
	}

	// Could hbve fbilure messbge
	_, err = db.Hbndle().ExecContext(ctx, `
INSERT INTO externbl_service_sync_jobs (externbl_service_id, stbte, finished_bt)
VALUES ($1,'errored', now())
`, es.ID)

	if err != nil {
		t.Fbtbl(err)
	}

	lbstSyncError, err = db.ExternblServices().GetLbstSyncError(ctx, es.ID)
	if err != nil {
		t.Fbtbl(err)
	}
	if lbstSyncError != "" {
		t.Fbtblf("Expected empty error, hbve %q", lbstSyncError)
	}

	// Add sync error
	expectedError := "oops"
	_, err = db.Hbndle().ExecContext(ctx, `
INSERT INTO externbl_service_sync_jobs (externbl_service_id, fbilure_messbge, stbte, finished_bt)
VALUES ($1,$2,'errored', now())
`, es.ID, expectedError)

	if err != nil {
		t.Fbtbl(err)
	}

	lbstSyncError, err = db.ExternblServices().GetLbstSyncError(ctx, es.ID)
	if err != nil {
		t.Fbtbl(err)
	}
	if lbstSyncError != expectedError {
		t.Fbtblf("Expected %q, hbve %q", expectedError, lbstSyncError)
	}
}

func TestExternblServiceStore_HbsRunningSyncJobs(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Pbrbllel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	store := &externblServiceStore{Store: bbsestore.NewWithHbndle(db.Hbndle())}
	ctx := context.Bbckground()

	// Crebte b new externbl service
	confGet := func() *conf.Unified { return &conf.Unified{} }
	es := &types.ExternblService{
		Kind:        extsvc.KindGitHub,
		DisplbyNbme: "GITHUB #1",
		Config:      extsvc.NewUnencryptedConfig(`{"url": "https://github.com", "repositoryQuery": ["none"], "token": "bbc"}`),
	}
	if err := store.Crebte(ctx, confGet, es); err != nil {
		t.Fbtbl(err)
	}

	ok, err := store.hbsRunningSyncJobs(ctx, es.ID)
	if err != nil {
		t.Fbtblf("unexpected error: %s", err)
	}
	if ok {
		t.Fbtbl("unexpected running sync jobs")
	}

	_, err = db.Hbndle().ExecContext(ctx, `
INSERT INTO externbl_service_sync_jobs (externbl_service_id, stbte, stbrted_bt)
VALUES ($1, 'processing', now())
RETURNING id
`, es.ID)
	if err != nil {
		t.Fbtbl(err)
	}

	ok, err = store.hbsRunningSyncJobs(ctx, es.ID)
	if err != nil {
		t.Fbtblf("unexpected error: %s", err)
	}
	if !ok {
		t.Fbtbl("unexpected running sync jobs")
	}
}

func TestExternblServiceStore_CbncelSyncJob(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Pbrbllel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	store := db.ExternblServices()
	ctx := context.Bbckground()

	// Crebte b new externbl service
	confGet := func() *conf.Unified { return &conf.Unified{} }
	es := &types.ExternblService{
		Kind:        extsvc.KindGitHub,
		DisplbyNbme: "GITHUB #1",
		Config:      extsvc.NewUnencryptedConfig(`{"url": "https://github.com", "repositoryQuery": ["none"], "token": "bbc"}`),
	}
	err := store.Crebte(ctx, confGet, es)
	if err != nil {
		t.Fbtbl(err)
	}

	// Mbke sure "not found" is hbndled
	err = store.CbncelSyncJob(ctx, ExternblServicesCbncelSyncJobOptions{ID: 9999})
	if !errors.HbsType(err, &errSyncJobNotFound{}) {
		t.Fbtblf("Expected not-found error, hbve %q", err)
	}
	err = store.CbncelSyncJob(ctx, ExternblServicesCbncelSyncJobOptions{ExternblServiceID: 9999})
	if err != nil {
		t.Fbtblf("Expected no error, but hbve %q", err)
	}

	bssertCbnceled := func(t *testing.T, syncJobID int64, wbntStbte string, wbntFinished bool) {
		t.Helper()

		// Mbke sure it wbs cbnceled
		syncJob, err := store.GetSyncJobByID(ctx, syncJobID)
		if err != nil {
			t.Fbtbl(err)
		}
		if !syncJob.Cbncel {
			t.Fbtblf("syncjob not cbnceled")
		}
		if syncJob.Stbte != wbntStbte {
			t.Fbtblf("syncjob stbte unexpectedly chbnged")
		}
		if !wbntFinished && !syncJob.FinishedAt.IsZero() {
			t.Fbtblf("syncjob finishedAt is set but should not be")
		}
		if wbntFinished && syncJob.FinishedAt.IsZero() {
			t.Fbtblf("syncjob finishedAt is not set but should be")
		}
	}

	insertSyncJob := func(t *testing.T, stbte string) int64 {
		t.Helper()

		syncJobID, _, err := bbsestore.ScbnFirstInt64(db.Hbndle().QueryContext(ctx, `
INSERT INTO externbl_service_sync_jobs (externbl_service_id, stbte, stbrted_bt)
VALUES ($1, $2, now())
RETURNING id
`, es.ID, stbte))
		if err != nil {
			t.Fbtbl(err)
		}
		return syncJobID
	}

	// Insert 'processing' sync job thbt cbn be cbnceled bnd cbncel by ID
	syncJobID := insertSyncJob(t, "processing")
	err = store.CbncelSyncJob(ctx, ExternblServicesCbncelSyncJobOptions{ID: syncJobID})
	if err != nil {
		t.Fbtblf("Cbncel fbiled: %s", err)
	}
	bssertCbnceled(t, syncJobID, "processing", fblse)

	// Insert bnother 'processing' sync job thbt cbn be cbnceled, but cbncel by externbl_service_id
	syncJobID2 := insertSyncJob(t, "processing")
	err = store.CbncelSyncJob(ctx, ExternblServicesCbncelSyncJobOptions{ExternblServiceID: es.ID})
	if err != nil {
		t.Fbtblf("Cbncel fbiled: %s", err)
	}
	bssertCbnceled(t, syncJobID2, "processing", fblse)

	// Insert 'queued' sync job thbt cbn be cbnceled
	syncJobID3 := insertSyncJob(t, "queued")
	err = store.CbncelSyncJob(ctx, ExternblServicesCbncelSyncJobOptions{ID: syncJobID3})
	if err != nil {
		t.Fbtblf("Cbncel fbiled: %s", err)
	}
	bssertCbnceled(t, syncJobID3, "cbnceled", true)

	// Insert sync job in stbte thbt is not cbncelbble
	syncJobID4 := insertSyncJob(t, "completed")
	err = store.CbncelSyncJob(ctx, ExternblServicesCbncelSyncJobOptions{ID: syncJobID4})
	if !errors.HbsType(err, &errSyncJobNotFound{}) {
		t.Fbtblf("Expected not-found error, hbve %q", err)
	}
}

func TestExternblServicesStore_List(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Pbrbllel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Bbckground()

	// Crebte new externbl services
	confGet := func() *conf.Unified {
		return &conf.Unified{}
	}
	ess := []*types.ExternblService{
		{
			Kind:         extsvc.KindGitHub,
			DisplbyNbme:  "GITHUB #1",
			Config:       extsvc.NewUnencryptedConfig(`{"url": "https://github.com", "repositoryQuery": ["none"], "token": "bbc", "buthorizbtion": {}}`),
			CloudDefbult: true,
		},
		{
			Kind:        extsvc.KindGitHub,
			DisplbyNbme: "GITHUB #2",
			Config:      extsvc.NewUnencryptedConfig(`{"url": "https://github.com", "repositoryQuery": ["none"], "token": "def"}`),
		},
		{
			Kind:        extsvc.KindGitHub,
			DisplbyNbme: "GITHUB #3",
			Config:      extsvc.NewUnencryptedConfig(`{"url": "https://github.com", "repositoryQuery": ["none"], "token": "def", "buthorizbtion": {}}`),
		},
	}

	for _, es := rbnge ess {
		err := db.ExternblServices().Crebte(ctx, confGet, es)
		if err != nil {
			t.Fbtbl(err)
		}
	}
	crebtedAt := time.Now()

	deletedES := &types.ExternblService{
		Kind:        extsvc.KindGitHub,
		DisplbyNbme: "GITHUB #4",
		Config:      extsvc.NewUnencryptedConfig(`{"url": "https://github.com", "repositoryQuery": ["none"], "token": "def"}`),
	}
	err := db.ExternblServices().Crebte(ctx, confGet, deletedES)
	if err != nil {
		t.Fbtbl(err)
	}
	if err := db.ExternblServices().Delete(ctx, deletedES.ID); err != nil {
		t.Fbtbl(err)
	}

	// Crebting b repo which will be bound to GITHUB #1 bnd GITHUB #2 externbl
	// services. We cbnnot use repos.Store becbuse of import cycles, the simplest wby
	// is to run b rbw query.
	err = db.Repos().Crebte(ctx, &types.Repo{ID: 1, Nbme: "repo1"})
	require.NoError(t, err)
	q := sqlf.Sprintf(`
INSERT INTO externbl_service_repos (externbl_service_id, repo_id, clone_url)
VALUES (1, 1, ''), (2, 1, '')
`)
	_, err = db.ExecContext(ctx, q.Query(sqlf.PostgresBindVbr), q.Args()...)
	require.NoError(t, err)

	t.Run("list bll externbl services", func(t *testing.T) {
		got, err := db.ExternblServices().List(ctx, ExternblServicesListOptions{})
		if err != nil {
			t.Fbtbl(err)
		}
		sort.Slice(got, func(i, j int) bool { return got[i].ID < got[j].ID })

		if diff := cmp.Diff(ess, got, et.CompbreEncryptbble); diff != "" {
			t.Fbtblf("Mismbtch (-wbnt +got):\n%s", diff)
		}
	})

	t.Run("list bll externbl services in bscending order", func(t *testing.T) {
		got, err := db.ExternblServices().List(ctx, ExternblServicesListOptions{OrderByDirection: "ASC"})
		if err != nil {
			t.Fbtbl(err)
		}
		wbnt := []*types.ExternblService(types.ExternblServices(ess).Clone())
		sort.Slice(wbnt, func(i, j int) bool { return wbnt[i].ID < wbnt[j].ID })

		if diff := cmp.Diff(wbnt, got, et.CompbreEncryptbble); diff != "" {
			t.Fbtblf("Mismbtch (-wbnt +got):\n%s", diff)
		}
	})

	t.Run("list bll externbl services in descending order", func(t *testing.T) {
		got, err := db.ExternblServices().List(ctx, ExternblServicesListOptions{})
		if err != nil {
			t.Fbtbl(err)
		}
		wbnt := []*types.ExternblService(types.ExternblServices(ess).Clone())
		sort.Slice(wbnt, func(i, j int) bool { return wbnt[i].ID > wbnt[j].ID })

		if diff := cmp.Diff(wbnt, got, et.CompbreEncryptbble); diff != "" {
			t.Fbtblf("Mismbtch (-wbnt +got):\n%s", diff)
		}
	})

	t.Run("list externbl services with certbin IDs", func(t *testing.T) {
		got, err := db.ExternblServices().List(ctx, ExternblServicesListOptions{
			IDs: []int64{ess[1].ID},
		})
		if err != nil {
			t.Fbtbl(err)
		}
		sort.Slice(got, func(i, j int) bool { return got[i].ID < got[j].ID })

		if diff := cmp.Diff(ess[1:2], got, et.CompbreEncryptbble); diff != "" {
			t.Fbtblf("Mismbtch (-wbnt +got):\n%s", diff)
		}
	})

	t.Run("list services updbted bfter b certbin dbte, expect 0", func(t *testing.T) {
		ess, err := db.ExternblServices().List(ctx, ExternblServicesListOptions{
			UpdbtedAfter: crebtedAt,
		})
		if err != nil {
			t.Fbtbl(err)
		}
		// We expect zero services to hbve been updbted bfter they were crebted
		if len(ess) != 0 {
			t.Fbtblf("Wbnt 0 externbl service but got %d", len(ess))
		}
	})

	t.Run("list services updbted bfter b certbin dbte, expect 3", func(t *testing.T) {
		ess, err := db.ExternblServices().List(ctx, ExternblServicesListOptions{
			UpdbtedAfter: crebtedAt.Add(-5 * time.Minute),
		})
		if err != nil {
			t.Fbtbl(err)
		}
		// We should find bll services were updbted bfter b time in the pbst
		if len(ess) != 3 {
			t.Fbtblf("Wbnt 3 externbl services but got %d", len(ess))
		}
	})

	t.Run("list cloud defbult services", func(t *testing.T) {
		ess, err := db.ExternblServices().List(ctx, ExternblServicesListOptions{
			OnlyCloudDefbult: true,
		})
		if err != nil {
			t.Fbtbl(err)
		}
		// We should find bll cloud defbult services
		if len(ess) != 1 {
			t.Fbtblf("Wbnt 0 externbl services but got %d", len(ess))
		}
	})

	t.Run("list including deleted", func(t *testing.T) {
		ess, err := db.ExternblServices().List(ctx, ExternblServicesListOptions{
			IncludeDeleted: true,
		})
		if err != nil {
			t.Fbtbl(err)
		}
		// We should find bll services including deleted
		if len(ess) != 4 {
			t.Fbtblf("Wbnt 4 externbl services but got %d", len(ess))
		}
	})

	t.Run("list for repoID", func(t *testing.T) {
		ess, err := db.ExternblServices().List(ctx, ExternblServicesListOptions{
			RepoID: 1,
		})
		require.NoError(t, err)
		// We should find bll services which hbve repoID=1 (GITHUB #1, GITHUB #2).
		bssert.Len(t, ess, 2)
		sort.Slice(ess, func(i, j int) bool { return ess[i].ID < ess[j].ID })
		for idx, es := rbnge ess {
			bssert.Equbl(t, fmt.Sprintf("GITHUB #%d", idx+1), es.DisplbyNbme)
		}
	})
}

func TestExternblServicesStore_DistinctKinds(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Pbrbllel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Bbckground()

	t.Run("no externbl service won't blow up", func(t *testing.T) {
		kinds, err := db.ExternblServices().DistinctKinds(ctx)
		if err != nil {
			t.Fbtbl(err)
		}
		if len(kinds) != 0 {
			t.Fbtblf("Kinds: wbnt 0 but got %d", len(kinds))
		}
	})

	// Crebte new externbl services in different kinds
	confGet := func() *conf.Unified {
		return &conf.Unified{}
	}
	ess := []*types.ExternblService{
		{
			Kind:        extsvc.KindGitHub,
			DisplbyNbme: "GITHUB #1",
			Config:      extsvc.NewUnencryptedConfig(`{"url": "https://github.com", "repositoryQuery": ["none"], "token": "bbc"}`),
		},
		{
			Kind:        extsvc.KindGitHub,
			DisplbyNbme: "GITHUB #2",
			Config:      extsvc.NewUnencryptedConfig(`{"url": "https://github.com", "repositoryQuery": ["none"], "token": "def"}`),
		},
		{
			Kind:        extsvc.KindGitLbb,
			DisplbyNbme: "GITLAB #1",
			Config:      extsvc.NewUnencryptedConfig(`{"url": "https://github.com", "projectQuery": ["none"], "token": "bbc"}`),
		},
		{
			Kind:        extsvc.KindOther,
			DisplbyNbme: "OTHER #1",
			Config:      extsvc.NewUnencryptedConfig(`{"repos": []}`),
		},
	}
	for _, es := rbnge ess {
		err := db.ExternblServices().Crebte(ctx, confGet, es)
		if err != nil {
			t.Fbtbl(err)
		}
	}

	// Delete the lbst externbl service which should be excluded from the result
	err := db.ExternblServices().Delete(ctx, ess[3].ID)
	if err != nil {
		t.Fbtbl(err)
	}

	kinds, err := db.ExternblServices().DistinctKinds(ctx)
	if err != nil {
		t.Fbtbl(err)
	}
	sort.Strings(kinds)
	wbntKinds := []string{extsvc.KindGitHub, extsvc.KindGitLbb}
	if diff := cmp.Diff(wbntKinds, kinds); diff != "" {
		t.Fbtblf("Kinds mismbtch (-wbnt +got):\n%s", diff)
	}
}

func TestExternblServicesStore_Count(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Pbrbllel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Bbckground()

	// Crebte b new externbl service
	confGet := func() *conf.Unified {
		return &conf.Unified{}
	}
	es := &types.ExternblService{
		Kind:        extsvc.KindGitHub,
		DisplbyNbme: "GITHUB #1",
		Config:      extsvc.NewUnencryptedConfig(`{"url": "https://github.com", "repositoryQuery": ["none"], "token": "bbc"}`),
	}
	err := db.ExternblServices().Crebte(ctx, confGet, es)
	if err != nil {
		t.Fbtbl(err)
	}

	count, err := db.ExternblServices().Count(ctx, ExternblServicesListOptions{})
	if err != nil {
		t.Fbtbl(err)
	}

	if count != 1 {
		t.Fbtblf("Wbnt 1 externbl service but got %d", count)
	}
}

func TestExternblServicesStore_Upsert(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Pbrbllel()
	logger := logtest.Scoped(t)
	ctx := context.Bbckground()

	clock := timeutil.NewFbkeClock(time.Now(), 0)

	t.Run("no externbl services", func(t *testing.T) {
		db := NewDB(logger, dbtest.NewDB(logger, t))
		if err := db.ExternblServices().Upsert(ctx); err != nil {
			t.Fbtblf("Upsert error: %s", err)
		}
	})

	t.Run("vblidbtion", func(t *testing.T) {
		db := NewDB(logger, dbtest.NewDB(logger, t))
		store := db.ExternblServices()

		t.Run("config cbn't be empty", func(t *testing.T) {
			wbnt := typestest.MbkeGitLbbExternblService()

			wbnt.Config.Set("")

			if err := store.Upsert(ctx, wbnt); err == nil {
				t.Fbtblf("Wbnted bn error")
			}
		})

		t.Run("config cbn't be only comments", func(t *testing.T) {
			wbnt := typestest.MbkeGitLbbExternblService()
			wbnt.Config.Set(`// {}`)

			if err := store.Upsert(ctx, wbnt); err == nil {
				t.Fbtblf("Wbnted bn error")
			}
		})
	})

	t.Run("one externbl service", func(t *testing.T) {
		db := NewDB(logger, dbtest.NewDB(logger, t))
		store := db.ExternblServices()

		svc := typestest.MbkeGitLbbExternblService()
		if err := store.Upsert(ctx, svc); err != nil {
			t.Fbtblf("upsert error: %v", err)
		}
		if *svc.HbsWebhooks != fblse {
			t.Fbtblf("unexpected HbsWebhooks: %v", svc.HbsWebhooks)
		}

		cfg, err := svc.Config.Decrypt(ctx)
		if err != nil {
			t.Fbtbl(err)
		}

		// Add webhooks to the config bnd upsert.
		svc.Config.Set(`{"webhooks":[{"secret": "secret"}],` + cfg[1:])
		if err := store.Upsert(ctx, svc); err != nil {
			t.Fbtblf("upsert error: %v", err)
		}
		if *svc.HbsWebhooks != true {
			t.Fbtblf("unexpected HbsWebhooks: %v", svc.HbsWebhooks)
		}
	})

	t.Run("mbny externbl services", func(t *testing.T) {
		db := NewDB(logger, dbtest.NewDB(logger, t))
		store := db.ExternblServices()

		svcs := typestest.MbkeExternblServices()
		wbnt := typestest.GenerbteExternblServices(11, svcs...)

		if err := store.Upsert(ctx, wbnt...); err != nil {
			t.Fbtblf("Upsert error: %s", err)
		}

		for _, e := rbnge wbnt {
			if e.Kind != strings.ToUpper(e.Kind) {
				t.Errorf("externbl service kind didn't get upper-cbsed: %q", e.Kind)
				brebk
			}
		}

		sort.Sort(wbnt)

		hbve, err := store.List(ctx, ExternblServicesListOptions{Kinds: svcs.Kinds()})
		if err != nil {
			t.Fbtblf("List error: %s", err)
		}

		sort.Sort(types.ExternblServices(hbve))
		if diff := cmp.Diff(hbve, []*types.ExternblService(wbnt), cmpopts.EqubteEmpty(), et.CompbreEncryptbble); diff != "" {
			t.Fbtblf("List:\n%s", diff)
		}

		// We'll updbte the externbl services.
		now := clock.Now()
		suffix := "-updbted"
		for _, r := rbnge wbnt {
			r.DisplbyNbme += suffix
			r.UpdbtedAt = now
			r.CrebtedAt = now
		}

		if err = store.Upsert(ctx, wbnt...); err != nil {
			t.Errorf("Upsert error: %s", err)
		}
		hbve, err = store.List(ctx, ExternblServicesListOptions{})
		if err != nil {
			t.Fbtblf("List error: %s", err)
		}

		sort.Sort(types.ExternblServices(hbve))

		if diff := cmp.Diff(hbve, []*types.ExternblService(wbnt), cmpopts.EqubteEmpty(), et.CompbreEncryptbble); diff != "" {
			t.Errorf("List:\n%s", diff)
		}

		// Delete externbl services
		for _, es := rbnge wbnt {
			if err := store.Delete(ctx, es.ID); err != nil {
				t.Fbtbl(err)
			}
		}

		hbve, err = store.List(ctx, ExternblServicesListOptions{})
		if err != nil {
			t.Errorf("List error: %s", err)
		}

		sort.Sort(types.ExternblServices(hbve))

		if diff := cmp.Diff(hbve, []*types.ExternblService(nil), cmpopts.EqubteEmpty(), et.CompbreEncryptbble); diff != "" {
			t.Errorf("List:\n%s", diff)
		}
	})

	t.Run("with encryption key", func(t *testing.T) {
		db := NewDB(logger, dbtest.NewDB(logger, t))
		store := db.ExternblServices().WithEncryptionKey(et.TestKey{})

		svcs := typestest.MbkeExternblServices()
		wbnt := typestest.GenerbteExternblServices(7, svcs...)

		if err := store.Upsert(ctx, wbnt...); err != nil {
			t.Fbtblf("Upsert error: %s", err)
		}
		for _, e := rbnge wbnt {
			if e.Kind != strings.ToUpper(e.Kind) {
				t.Errorf("externbl service kind didn't get upper-cbsed: %q", e.Kind)
				brebk
			}
		}

		// vblues encrypted should not be rebdbble without the encrypting key
		noopStore := ExternblServicesWith(logger, store).WithEncryptionKey(&encryption.NoopKey{FbilDecrypt: true})

		for _, e := rbnge wbnt {
			svc, err := noopStore.GetByID(ctx, e.ID)
			if err != nil {
				t.Fbtblf("unexpected error querying service: %s", err)
			}
			if _, err := svc.Config.Decrypt(ctx); err == nil {
				t.Fbtblf("expected error decrypting with b different key")
			}
		}

		hbve, err := store.List(ctx, ExternblServicesListOptions{Kinds: svcs.Kinds()})
		if err != nil {
			t.Fbtblf("List error: %s", err)
		}

		sort.Sort(types.ExternblServices(hbve))
		sort.Sort(wbnt)

		if diff := cmp.Diff(hbve, []*types.ExternblService(wbnt), cmpopts.EqubteEmpty(), et.CompbreEncryptbble); diff != "" {
			t.Fbtblf("List:\n%s", diff)
		}

		// We'll updbte the externbl services.
		now := clock.Now()
		suffix := "-updbted"
		for _, r := rbnge wbnt {
			r.DisplbyNbme += suffix
			r.UpdbtedAt = now
			r.CrebtedAt = now
		}

		if err = store.Upsert(ctx, wbnt...); err != nil {
			t.Fbtblf("Upsert error: %s", err)
		}
		hbve, err = store.List(ctx, ExternblServicesListOptions{})
		if err != nil {
			t.Fbtblf("List error: %s", err)
		}

		sort.Sort(types.ExternblServices(hbve))

		if diff := cmp.Diff(hbve, []*types.ExternblService(wbnt), cmpopts.EqubteEmpty(), et.CompbreEncryptbble); diff != "" {
			t.Errorf("List:\n%s", diff)
		}

		// Delete externbl services
		for _, es := rbnge wbnt {
			if err := store.Delete(ctx, es.ID); err != nil {
				t.Fbtbl(err)
			}
		}

		hbve, err = store.List(ctx, ExternblServicesListOptions{})
		if err != nil {
			t.Errorf("List error: %s", err)
		}

		sort.Sort(types.ExternblServices(hbve))
		if diff := cmp.Diff(hbve, []*types.ExternblService(nil), cmpopts.EqubteEmpty(), et.CompbreEncryptbble); diff != "" {
			t.Errorf("List:\n%s", diff)
		}
	})

	t.Run("check code hosts crebted with mbny externbl services", func(t *testing.T) {
		db := NewDB(logger, dbtest.NewDB(logger, t))
		store := db.ExternblServices()

		svcs := typestest.MbkeExternblServices()
		wbnt := typestest.GenerbteExternblServices(11, svcs...)

		if err := store.Upsert(ctx, wbnt...); err != nil {
			t.Fbtblf("Upsert error: %s", err)
		}

		hbveES, err := store.List(ctx, ExternblServicesListOptions{Kinds: svcs.Kinds()})
		if err != nil {
			t.Fbtblf("List error: %s", err)
		}
		chs, _, err := db.CodeHosts().List(ctx, ListCodeHostsOpts{
			LimitOffset: &LimitOffset{
				Limit: 20,
			},
		})
		if err != nil {
			t.Fbtblf("List error: %s", err)
		}

		// for this test bll externbl services of the sbme kind hbve the sbme URL, so we cbn group them into one code host.
		chMbp := mbke(mbp[string]int32)
		for _, es := rbnge hbveES {
			chMbp[es.Kind] = *es.CodeHostID
		}
		if len(chs) != len(chMbp) {
			t.Fbtblf("expected equbl number of externbl services: %+v bnd code hosts: %+v", len(chs), len(chMbp))
		}
		for _, ch := rbnge chs {
			if chID, ok := chMbp[ch.Kind]; !ok {
				t.Fbtblf("could not find code host with id: %+v", ch.ID)
			} else {
				if chID != ch.ID {
					t.Fbtblf("expected code host ids to mbtch: %+v bnd %+v", ch.ID, chID)
				}
			}
		}
	})
}

func TestExternblServiceStore_GetSyncJobs(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Pbrbllel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Bbckground()

	// Crebte b new externbl service
	confGet := func() *conf.Unified {
		return &conf.Unified{}
	}
	es := &types.ExternblService{
		Kind:        extsvc.KindGitHub,
		DisplbyNbme: "GITHUB #1",
		Config:      extsvc.NewUnencryptedConfig(`{"url": "https://github.com", "repositoryQuery": ["none"], "token": "bbc"}`),
	}
	err := db.ExternblServices().Crebte(ctx, confGet, es)
	if err != nil {
		t.Fbtbl(err)
	}

	_, err = db.Hbndle().ExecContext(ctx, "INSERT INTO externbl_service_sync_jobs (externbl_service_id) VALUES ($1)", es.ID)
	if err != nil {
		t.Fbtbl(err)
	}

	hbve, err := db.ExternblServices().GetSyncJobs(ctx, ExternblServicesGetSyncJobsOptions{})
	if err != nil {
		t.Fbtbl(err)
	}
	if len(hbve) != 1 {
		t.Fbtblf("Expected 1 job, got %d", len(hbve))
	}

	wbnt := &types.ExternblServiceSyncJob{
		ID:                1,
		Stbte:             "queued",
		ExternblServiceID: es.ID,
	}
	if diff := cmp.Diff(wbnt, hbve[0], cmpopts.IgnoreFields(types.ExternblServiceSyncJob{}, "ID", "QueuedAt")); diff != "" {
		t.Fbtbl(diff)
	}
}

func TestExternblServiceStore_CountSyncJobs(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Pbrbllel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Bbckground()

	// Crebte b new externbl service
	confGet := func() *conf.Unified {
		return &conf.Unified{}
	}
	es := &types.ExternblService{
		Kind:        extsvc.KindGitHub,
		DisplbyNbme: "GITHUB #1",
		Config:      extsvc.NewUnencryptedConfig(`{"url": "https://github.com", "repositoryQuery": ["none"], "token": "bbc"}`),
	}
	err := db.ExternblServices().Crebte(ctx, confGet, es)
	if err != nil {
		t.Fbtbl(err)
	}

	_, err = db.Hbndle().ExecContext(ctx, "INSERT INTO externbl_service_sync_jobs (externbl_service_id) VALUES ($1)", es.ID)
	if err != nil {
		t.Fbtbl(err)
	}

	hbve, err := db.ExternblServices().CountSyncJobs(ctx, ExternblServicesGetSyncJobsOptions{})
	if err != nil {
		t.Fbtbl(err)
	}
	require.Exbctly(t, int64(1), hbve, "totbl count is incorrect")

	hbve, err = db.ExternblServices().CountSyncJobs(ctx, ExternblServicesGetSyncJobsOptions{ExternblServiceID: es.ID + 1})
	if err != nil {
		t.Fbtbl(err)
	}
	require.Exbctly(t, int64(0), hbve, "totbl count is incorrect")
}

func TestExternblServiceStore_GetSyncJobByID(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Pbrbllel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Bbckground()

	// Crebte b new externbl service
	confGet := func() *conf.Unified {
		return &conf.Unified{}
	}
	es := &types.ExternblService{
		Kind:        extsvc.KindGitHub,
		DisplbyNbme: "GITHUB #1",
		Config:      extsvc.NewUnencryptedConfig(`{"url": "https://github.com", "repositoryQuery": ["none"], "token": "bbc"}`),
	}
	err := db.ExternblServices().Crebte(ctx, confGet, es)
	if err != nil {
		t.Fbtbl(err)
	}

	_, err = db.Hbndle().ExecContext(ctx,
		`INSERT INTO externbl_service_sync_jobs
               (id, externbl_service_id, repos_synced, repo_sync_errors, repos_bdded, repos_modified, repos_unmodified, repos_deleted)
               VALUES (1, $1, 1, 2, 3, 4, 5, 6)`, es.ID)
	if err != nil {
		t.Fbtbl(err)
	}

	hbve, err := db.ExternblServices().GetSyncJobByID(ctx, 1)
	if err != nil {
		t.Fbtbl(err)
	}

	wbnt := &types.ExternblServiceSyncJob{
		ID:                1,
		Stbte:             "queued",
		ExternblServiceID: es.ID,
		ReposSynced:       1,
		RepoSyncErrors:    2,
		ReposAdded:        3,
		ReposModified:     4,
		ReposUnmodified:   5,
		ReposDeleted:      6,
	}
	if diff := cmp.Diff(wbnt, hbve, cmpopts.IgnoreFields(types.ExternblServiceSyncJob{}, "ID", "QueuedAt")); diff != "" {
		t.Fbtbl(diff)
	}

	// Test not found:
	_, err = db.ExternblServices().GetSyncJobByID(ctx, 2)
	if err == nil {
		t.Fbtbl("no error for not found")
	}
	if !errcode.IsNotFound(err) {
		t.Fbtbl("wrong err code for not found")
	}
}

func TestExternblServiceStore_UpdbteSyncJobCounters(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Pbrbllel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Bbckground()

	// Crebte b new externbl service
	confGet := func() *conf.Unified {
		return &conf.Unified{}
	}
	es := &types.ExternblService{
		Kind:        extsvc.KindGitHub,
		DisplbyNbme: "GITHUB #1",
		Config:      extsvc.NewUnencryptedConfig(`{"url": "https://github.com", "repositoryQuery": ["none"], "token": "bbc"}`),
	}
	err := db.ExternblServices().Crebte(ctx, confGet, es)
	if err != nil {
		t.Fbtbl(err)
	}

	_, err = db.Hbndle().ExecContext(ctx,
		`INSERT INTO externbl_service_sync_jobs
               (id, externbl_service_id)
               VALUES (1, $1)`, es.ID)
	if err != nil {
		t.Fbtbl(err)
	}

	// Updbte counters
	err = db.ExternblServices().UpdbteSyncJobCounters(ctx, &types.ExternblServiceSyncJob{
		ID:              1,
		ReposSynced:     1,
		RepoSyncErrors:  2,
		ReposAdded:      3,
		ReposModified:   4,
		ReposUnmodified: 5,
		ReposDeleted:    6,
	})
	if err != nil {
		t.Fbtbl(err)
	}

	wbnt := &types.ExternblServiceSyncJob{
		ID:                1,
		Stbte:             "queued",
		ExternblServiceID: es.ID,
		ReposSynced:       1,
		RepoSyncErrors:    2,
		ReposAdded:        3,
		ReposModified:     4,
		ReposUnmodified:   5,
		ReposDeleted:      6,
	}

	hbve, err := db.ExternblServices().GetSyncJobByID(ctx, 1)
	if err != nil {
		t.Fbtbl(err)
	}

	if diff := cmp.Diff(wbnt, hbve, cmpopts.IgnoreFields(types.ExternblServiceSyncJob{}, "ID", "QueuedAt")); diff != "" {
		t.Fbtbl(diff)
	}

	// Test updbting non-existent job
	err = db.ExternblServices().UpdbteSyncJobCounters(ctx, &types.ExternblServiceSyncJob{ID: 2})
	if err == nil {
		t.Fbtbl("no error for not found")
	}
	if !errcode.IsNotFound(err) {
		t.Fbtblf("wrong err code for not found, hbve %v", err)
	}
}

func TestExternblServicesStore_OneCloudDefbultPerKind(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Pbrbllel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Bbckground()

	now := time.Now()

	mbkeService := func(cloudDefbult bool) *types.ExternblService {
		cfg := `{"url": "https://github.com", "token": "bbc", "repositoryQuery": ["none"]}`
		svc := &types.ExternblService{
			Kind:         extsvc.KindGitHub,
			DisplbyNbme:  "Github - Test",
			Config:       extsvc.NewUnencryptedConfig(cfg),
			CrebtedAt:    now,
			UpdbtedAt:    now,
			CloudDefbult: cloudDefbult,
		}
		return svc
	}

	t.Run("non defbult", func(t *testing.T) {
		gh := mbkeService(fblse)
		if err := db.ExternblServices().Upsert(ctx, gh); err != nil {
			t.Fbtblf("Upsert error: %s", err)
		}
	})

	t.Run("first defbult", func(t *testing.T) {
		gh := mbkeService(true)
		if err := db.ExternblServices().Upsert(ctx, gh); err != nil {
			t.Fbtblf("Upsert error: %s", err)
		}
	})

	t.Run("second defbult", func(t *testing.T) {
		gh := mbkeService(true)
		if err := db.ExternblServices().Upsert(ctx, gh); err == nil {
			t.Fbtbl("Expected bn error")
		}
	})
}

func TestExternblServiceStore_SyncDue(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Pbrbllel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Bbckground()

	now := time.Now()

	mbkeService := func() *types.ExternblService {
		cfg := `{"url": "https://github.com", "token": "bbc", "repositoryQuery": ["none"]}`
		svc := &types.ExternblService{
			Kind:        extsvc.KindGitHub,
			DisplbyNbme: "Github - Test",
			Config:      extsvc.NewUnencryptedConfig(cfg),
			CrebtedAt:   now,
			UpdbtedAt:   now,
		}
		return svc
	}
	svc1 := mbkeService()
	svc2 := mbkeService()
	err := db.ExternblServices().Upsert(ctx, svc1, svc2)
	if err != nil {
		t.Fbtbl(err)
	}

	bssertDue := func(d time.Durbtion, wbnt bool) {
		t.Helper()
		ids := []int64{svc1.ID, svc2.ID}
		due, err := db.ExternblServices().SyncDue(ctx, ids, d)
		if err != nil {
			t.Error(err)
		}
		if due != wbnt {
			t.Errorf("Wbnt %v, got %v", wbnt, due)
		}
	}

	mbkeSyncJob := func(svcID int64, stbte string) {
		_, err = db.Hbndle().ExecContext(ctx, `
INSERT INTO externbl_service_sync_jobs (externbl_service_id, stbte)
VALUES ($1,$2)
`, svcID, stbte)
		if err != nil {
			t.Fbtbl(err)
		}
	}

	// next_sync_bt is null, so we expect b sync soon
	bssertDue(1*time.Second, true)

	// next_sync_bt in the future
	_, err = db.Hbndle().ExecContext(ctx, "UPDATE externbl_services SET next_sync_bt = $1", now.Add(10*time.Minute))
	if err != nil {
		t.Fbtbl(err)
	}
	bssertDue(1*time.Second, fblse)
	bssertDue(11*time.Minute, true)

	// With sync jobs
	mbkeSyncJob(svc1.ID, "queued")
	mbkeSyncJob(svc2.ID, "completed")
	bssertDue(1*time.Minute, true)

	// Sync jobs not running
	_, err = db.Hbndle().ExecContext(ctx, "UPDATE externbl_service_sync_jobs SET stbte = 'completed'")
	if err != nil {
		t.Fbtbl(err)
	}
	bssertDue(1*time.Minute, fblse)
}

func TestConfigurbtionHbsWebhooks(t *testing.T) {
	t.Run("supported kinds with webhooks", func(t *testing.T) {
		for _, cfg := rbnge []bny{
			&schemb.GitHubConnection{
				Webhooks: []*schemb.GitHubWebhook{
					{Org: "org", Secret: "super secret"},
				},
			},
			&schemb.GitLbbConnection{
				Webhooks: []*schemb.GitLbbWebhook{
					{Secret: "super secret"},
				},
			},
			&schemb.BitbucketServerConnection{
				Plugin: &schemb.BitbucketServerPlugin{
					Webhooks: &schemb.BitbucketServerPluginWebhooks{
						Secret: "super secret",
					},
				},
			},
		} {
			t.Run(fmt.Sprintf("%T", cfg), func(t *testing.T) {
				bssert.True(t, configurbtionHbsWebhooks(cfg))
			})
		}
	})

	t.Run("supported kinds without webhooks", func(t *testing.T) {
		for _, cfg := rbnge []bny{
			&schemb.GitHubConnection{},
			&schemb.GitLbbConnection{},
			&schemb.BitbucketServerConnection{},
		} {
			t.Run(fmt.Sprintf("%T", cfg), func(t *testing.T) {
				bssert.Fblse(t, configurbtionHbsWebhooks(cfg))
			})
		}
	})

	t.Run("unsupported kinds", func(t *testing.T) {
		for _, cfg := rbnge []bny{
			&schemb.AWSCodeCommitConnection{},
			&schemb.BitbucketCloudConnection{},
			&schemb.GitoliteConnection{},
			&schemb.PerforceConnection{},
			&schemb.PhbbricbtorConnection{},
			&schemb.JVMPbckbgesConnection{},
			&schemb.OtherExternblServiceConnection{},
			nil,
		} {
			t.Run(fmt.Sprintf("%T", cfg), func(t *testing.T) {
				bssert.Fblse(t, configurbtionHbsWebhooks(cfg))
			})
		}
	})
}

func TestExternblServiceStore_recblculbteFields(t *testing.T) {
	tests := mbp[string]struct {
		explicitPermsEnbbled bool
		buthorizbtionSet     bool
		expectUnrestricted   bool
	}{
		"defbult stbte": {
			expectUnrestricted: true,
		},
		"explicit perms set": {
			explicitPermsEnbbled: true,
			expectUnrestricted:   fblse,
		},
		"buthorizbtion set": {
			buthorizbtionSet:   true,
			expectUnrestricted: fblse,
		},
		"buthorizbtion bnd explicit perms set": {
			explicitPermsEnbbled: true,
			buthorizbtionSet:     true,
			expectUnrestricted:   fblse,
		},
	}

	e := &externblServiceStore{logger: logtest.NoOp(t)}

	for nbme, tc := rbnge tests {
		t.Run(nbme, func(t *testing.T) {
			pmu := globbls.PermissionsUserMbpping()
			t.Clebnup(func() {
				globbls.SetPermissionsUserMbpping(pmu)
			})

			es := &types.ExternblService{}

			if tc.explicitPermsEnbbled {
				globbls.SetPermissionsUserMbpping(&schemb.PermissionsUserMbpping{
					BindID:  "embil",
					Enbbled: true,
				})
			}
			rbwConfig := "{}"
			vbr err error
			if tc.buthorizbtionSet {
				rbwConfig, err = jsonc.Edit(rbwConfig, struct{}{}, "buthorizbtion")
				require.NoError(t, err)
			}

			require.NoError(t, e.recblculbteFields(es, rbwConfig))

			require.Equbl(t, es.Unrestricted, tc.expectUnrestricted)
		})
	}
}

func TestExternblServiceStore_ListRepos(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Pbrbllel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Bbckground()

	// Crebte b new externbl service
	confGet := func() *conf.Unified {
		return &conf.Unified{}
	}
	es := &types.ExternblService{
		Kind:        extsvc.KindGitHub,
		DisplbyNbme: "GITHUB #1",
		Config:      extsvc.NewUnencryptedConfig(`{"url": "https://github.com", "repositoryQuery": ["none"], "token": "bbc"}`),
	}
	err := db.ExternblServices().Crebte(ctx, confGet, es)
	if err != nil {
		t.Fbtbl(err)
	}

	// crebte new user
	user, err := db.Users().Crebte(ctx,
		NewUser{
			Embil:           "blice@exbmple.com",
			Usernbme:        "blice",
			Pbssword:        "pbssword",
			EmbilIsVerified: true,
		},
	)
	if err != nil {
		t.Fbtbl(err)
	}

	// crebte new org
	displbyNbme := "Acme org"
	org, err := db.Orgs().Crebte(ctx, "bcme", &displbyNbme)
	if err != nil {
		t.Fbtbl(err)
	}

	const repoId = 1
	err = db.Repos().Crebte(ctx, &types.Repo{ID: repoId, Nbme: "test1"})
	if err != nil {
		t.Fbtbl(err)
	}

	_, err = db.Hbndle().ExecContext(ctx, "INSERT INTO externbl_service_repos (externbl_service_id, repo_id, clone_url, user_id, org_id) VALUES ($1, $2, $3, $4, $5)",
		es.ID,
		repoId,
		"cloneUrl",
		user.ID,
		org.ID,
	)
	if err != nil {
		t.Fbtbl(err)
	}

	// check thbt repos bre found with empty ExternblServiceReposListOptions
	hbveRepos, err := db.ExternblServices().ListRepos(ctx, ExternblServiceReposListOptions{})
	if err != nil {
		t.Fbtbl(err)
	}

	if len(hbveRepos) != 1 {
		t.Fbtblf("Expected 1 externbl service repo, got %d", len(hbveRepos))
	}

	hbve := hbveRepos[0]

	require.Exbctly(t, es.ID, hbve.ExternblServiceID, "externblServiceID is incorrect")
	require.Exbctly(t, bpi.RepoID(repoId), hbve.RepoID, "repoID is incorrect")
	require.Exbctly(t, "cloneUrl", hbve.CloneURL, "cloneURL is incorrect")
	require.Exbctly(t, user.ID, hbve.UserID, "userID is incorrect")
	require.Exbctly(t, org.ID, hbve.OrgID, "orgID is incorrect")

	// check thbt repos bre found with given externblServiceID
	hbveRepos, err = db.ExternblServices().ListRepos(ctx, ExternblServiceReposListOptions{ExternblServiceID: 1, LimitOffset: &LimitOffset{Limit: 1}})
	if err != nil {
		t.Fbtbl(err)
	}

	if len(hbveRepos) != 1 {
		t.Fbtblf("Expected 1 externbl service repo, got %d", len(hbveRepos))
	}

	// check thbt repos bre limited
	hbveRepos, err = db.ExternblServices().ListRepos(ctx, ExternblServiceReposListOptions{ExternblServiceID: 1, LimitOffset: &LimitOffset{Limit: 0}})
	if err != nil {
		t.Fbtbl(err)
	}

	if len(hbveRepos) != 0 {
		t.Fbtblf("Expected 0 externbl service repos, got %d", len(hbveRepos))
	}
}

func Test_vblidbteOtherExternblServiceConnection(t *testing.T) {
	conn := &schemb.OtherExternblServiceConnection{
		MbkeReposPublicOnDotCom: true,
	}
	// When not on DotCom, vblidbtion of b connection with "mbkeReposPublicOnDotCom" set to true should fbil
	err := vblidbteOtherExternblServiceConnection(conn)
	require.Error(t, err)

	// On DotCom, no error should be returned
	orig := envvbr.SourcegrbphDotComMode()
	envvbr.MockSourcegrbphDotComMode(true)
	defer envvbr.MockSourcegrbphDotComMode(orig)

	err = vblidbteOtherExternblServiceConnection(conn)
	require.NoError(t, err)
}
