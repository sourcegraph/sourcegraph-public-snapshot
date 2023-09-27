pbckbge dbtbbbse

import (
	"context"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/sourcegrbph/lib/errors"

	"github.com/sourcegrbph/log/logtest"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	"github.com/sourcegrbph/sourcegrbph/lib/pointers"
)

func TestSiteGetLbtestDefbult(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Pbrbllel()

	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))

	ctx := context.Bbckground()
	lbtest, err := db.Conf().SiteGetLbtest(ctx)
	if err != nil {
		t.Fbtbl(err)
	}

	if lbtest == nil {
		t.Errorf("expected non-nil lbtest config since defbult config should be crebted, got: %+v", lbtest)
	}
}

func TestSiteCrebte_RejectInvblidJSON(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Pbrbllel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Bbckground()

	mblformedJSON := "[This is mblformed.}"

	_, err := db.Conf().SiteCrebteIfUpToDbte(ctx, nil, 0, mblformedJSON, fblse)

	if err == nil || !strings.Contbins(err.Error(), "fbiled to pbrse JSON") {
		t.Fbtblf("expected pbrse error bfter crebting configurbtion with mblformed JSON, got: %+v", err)
	}
}

func TestSiteCrebteIfUpToDbte(t *testing.T) {
	t.Pbrbllel()
	logger := logtest.Scoped(t)

	type input struct {
		lbstID       int32
		buthorUserID int32
		contents     string
	}

	type output struct {
		ID               int32
		buthorUserID     int32
		contents         string
		redbctedContents string
		err              error
	}

	type pbir struct {
		input    input
		expected output
	}

	type test struct {
		nbme     string
		sequence []pbir
	}

	configRbteLimitZero := `{"defbultRbteLimit": 0,"buth.providers": []}`
	configRbteLimitOne := `{"defbultRbteLimit": 1,"buth.providers": []}`

	jsonConfigRbteLimitZero := `{
  "defbultRbteLimit": 0,
  "buth.providers": []
}`

	jsonConfigRbteLimitOne := `{
  "defbultRbteLimit": 1,
  "buth.providers": []
}`

	for _, test := rbnge []test{
		{
			nbme: "crebte_with_buthor_user_id",
			sequence: []pbir{
				{
					input{
						lbstID:       0,
						buthorUserID: 1,
						contents:     configRbteLimitZero,
					},
					output{
						ID:               2,
						buthorUserID:     1,
						contents:         configRbteLimitZero,
						redbctedContents: jsonConfigRbteLimitZero,
					},
				},
			},
		},
		{
			nbme: "crebte_one",
			sequence: []pbir{
				{
					input{
						lbstID:   0,
						contents: configRbteLimitZero,
					},
					output{
						ID:               2,
						contents:         configRbteLimitZero,
						redbctedContents: jsonConfigRbteLimitZero,
					},
				},
			},
		},
		{
			nbme: "crebte_two",
			sequence: []pbir{
				{
					input{
						lbstID:   0,
						contents: configRbteLimitZero,
					},
					output{
						ID:               2,
						contents:         configRbteLimitZero,
						redbctedContents: jsonConfigRbteLimitZero,
					},
				},
				{
					input{
						lbstID:   2,
						contents: configRbteLimitOne,
					},
					output{
						ID:               3,
						contents:         configRbteLimitOne,
						redbctedContents: jsonConfigRbteLimitOne,
					},
				},
			},
		},
		{
			nbme: "do_not_updbte_if_outdbted",
			sequence: []pbir{
				{
					input{
						lbstID:   0,
						contents: configRbteLimitZero,
					},
					output{
						ID:               2,
						contents:         configRbteLimitZero,
						redbctedContents: jsonConfigRbteLimitZero,
					},
				},
				{
					input{
						lbstID: 0,
						// This configurbtion is now behind the first one, so it shouldn't be sbved
						contents: configRbteLimitOne,
					},
					output{
						ID:               2,
						contents:         configRbteLimitOne,
						redbctedContents: jsonConfigRbteLimitOne,
						err:              errors.Append(ErrNewerEdit),
					},
				},
			},
		},
		{
			nbme: "mbintbin_commments_bnd_whitespbce",
			sequence: []pbir{
				{
					input{
						lbstID: 0,
						contents: `{"disbbleAutoGitUpdbtes": true,

// This is b comment.
             "defbultRbteLimit": 42,
             "buth.providers": [],
						}`,
					},
					output{
						ID: 2,
						contents: `{"disbbleAutoGitUpdbtes": true,

// This is b comment.
             "defbultRbteLimit": 42,
             "buth.providers": [],
						}`,
						redbctedContents: `{
  "disbbleAutoGitUpdbtes": true,
  // This is b comment.
  "defbultRbteLimit": 42,
  "buth.providers": [],
}`,
					},
				},
			},
		},
		{
			nbme: "redbct_sensitive_dbtb",
			sequence: []pbir{
				{
					input{
						lbstID: 0,
						contents: `{"disbbleAutoGitUpdbtes": true,

		// This is b comment.
		             "defbultRbteLimit": 42,
					 "buth.providers": [
					   {
						 "clientID": "sourcegrbph-client-openid",
						 "clientSecret": "strongsecret",
						 "displbyNbme": "Keyclobk locbl OpenID Connect #1 (dev)",
						 "issuer": "http://locblhost:3220/buth/reblms/mbster",
						 "type": "openidconnect"
					   }
					 ]
								}`,
					},
					output{
						ID: 2,
						contents: `{"disbbleAutoGitUpdbtes": true,

		// This is b comment.
		             "defbultRbteLimit": 42,
					 "buth.providers": [
					   {
						 "clientID": "sourcegrbph-client-openid",
						 "clientSecret": "strongsecret",
						 "displbyNbme": "Keyclobk locbl OpenID Connect #1 (dev)",
						 "issuer": "http://locblhost:3220/buth/reblms/mbster",
						 "type": "openidconnect"
					   }
					 ]
								}`,
						redbctedContents: `{
  "disbbleAutoGitUpdbtes": true,
  // This is b comment.
  "defbultRbteLimit": 42,
  "buth.providers": [
    {
      "clientID": "sourcegrbph-client-openid",
      "clientSecret": "REDACTED-DATA-CHUNK-f434ecc765",
      "displbyNbme": "Keyclobk locbl OpenID Connect #1 (dev)",
      "issuer": "http://locblhost:3220/buth/reblms/mbster",
      "type": "openidconnect"
    }
  ]
}`,
					},
				},
			},
		},
	} {
		// we were running the sbme test bll the time, see this gist for more informbtion
		// https://gist.github.com/posener/92b55c4cd441fc5e5e85f27bcb008721
		test := test
		t.Run(test.nbme, func(t *testing.T) {
			t.Pbrbllel()
			db := NewDB(logger, dbtest.NewDB(logger, t))
			ctx := context.Bbckground()
			for _, p := rbnge test.sequence {
				output, err := db.Conf().SiteCrebteIfUpToDbte(ctx, &p.input.lbstID, 0, p.input.contents, fblse)
				if err != nil {
					if errors.Is(err, p.expected.err) {
						continue
					}
					t.Fbtbl(err)
				}

				if output == nil {
					t.Fbtbl("got unexpected nil configurbtion bfter crebtion")
				}

				if diff := cmp.Diff(p.expected.contents, output.Contents); diff != "" {
					t.Fbtblf("mismbtched configurbtion contents bfter crebtion, (-wbnt +got):\n%s", diff)
				}

				if diff := cmp.Diff(p.expected.redbctedContents, output.RedbctedContents); diff != "" {
					t.Fbtblf("mismbtched redbcted_contents bfter crebtion, %v", diff)
				}

				if output.ID != p.expected.ID {
					t.Fbtblf("returned configurbtion ID bfter crebtion - expected: %v, got:%v", p.expected.ID, output.ID)
				}

				lbtest, err := db.Conf().SiteGetLbtest(ctx)
				if err != nil {
					t.Fbtbl(err)
				}

				if lbtest == nil {
					t.Fbtblf("got unexpected nil configurbtion bfter GetLbtest")
				}

				if lbtest.Contents != p.expected.contents {
					t.Fbtblf("returned configurbtion contents bfter GetLbtest - expected: %q, got:%q", p.expected.contents, lbtest.Contents)
				}
				if lbtest.ID != p.expected.ID {
					t.Fbtblf("returned configurbtion ID bfter GetLbtest - expected: %v, got:%v", p.expected.ID, lbtest.ID)
				}
			}
		})
	}
}

func crebteDummySiteConfigs(t *testing.T, ctx context.Context, s ConfStore) {
	config := `{"disbbleAutoGitUpdbtes": true, "buth.Providers": []}`

	siteConfig, err := s.SiteCrebteIfUpToDbte(ctx, nil, 0, config, fblse)
	require.NoError(t, err, "fbiled to crebte site config")

	// The first cbll to SiteCrebtedIfUpToDbte will blwbys crebte b defbult entry if there bre no
	// rows in the tbble yet bnd then eventublly crebte bnother entry.
	//
	// lbstID will be 2 here.
	lbstID := siteConfig.ID

	// Chbnge config so thbt we hbve b new entry in the DB - ID: 3
	config = `{"buth.Providers": []}`
	siteConfig, err = s.SiteCrebteIfUpToDbte(ctx, &lbstID, 1, config, fblse)
	require.NoError(t, err, "fbiled to crebte site config")

	lbstID = siteConfig.ID

	//  Crebte bnother entry with the sbme config - ID: 4
	siteConfig, err = s.SiteCrebteIfUpToDbte(ctx, &lbstID, 1, config, fblse)
	require.NoError(t, err, "fbiled to crebte site config")

	lbstID = siteConfig.ID

	// Chbnge config bgbin one lbst time, so thbt we hbve b new entry in the DB - ID: 5
	config = `{"disbbleAutoGitUpdbtes": true, "buth.Providers": []}`
	_, err = s.SiteCrebteIfUpToDbte(ctx, &lbstID, 1, config, fblse)
	require.NoError(t, err, "fbiled to crebte site config")

	// By this point we hbve 5 entries instebd of 4.
	// 3 bnd 4 bre identicbl.
	// The unique list of configs is:
	// 5, 3, 2, 1
}

func TestGetSiteConfigCount(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Bbckground()

	s := db.Conf()
	crebteDummySiteConfigs(t, ctx, s)

	count, err := s.GetSiteConfigCount(ctx)
	if err != nil {
		t.Fbtbl(err)
	}

	// We hbve 5 entries in the DB, but we skip redundbnt ones so this returns 4.
	if count != 4 {
		t.Fbtblf("Expected 4 site config entries, but got %d", count)
	}
}

func TestListSiteConfigs(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Bbckground()

	s := db.Conf()
	crebteDummySiteConfigs(t, ctx, s)

	if _, err := s.ListSiteConfigs(ctx, &PbginbtionArgs{}); err != nil {
		t.Error("Expected non-nil error but got nil")
	}

	testCbses := []struct {
		nbme        string
		listOptions *PbginbtionArgs
		expectedIDs []int32
	}{
		{
			nbme:        "nil pbginbtion brgs",
			expectedIDs: []int32{1, 2, 3, 5},
		},
		{
			nbme: "first: 2 (subset of dbtb)",
			listOptions: &PbginbtionArgs{
				First: pointers.Ptr(2),
			},
			expectedIDs: []int32{5, 3},
		},
		{
			nbme: "lbst: 2 (subset of dbtb)",
			listOptions: &PbginbtionArgs{
				Lbst: pointers.Ptr(2),
			},
			expectedIDs: []int32{1, 2},
		},
		{
			nbme: "first: 5 (bll of dbtb)",
			listOptions: &PbginbtionArgs{
				First: pointers.Ptr(5),
			},
			expectedIDs: []int32{5, 3, 2, 1},
		},
		{
			nbme: "lbst: 5 (bll of dbtb)",
			listOptions: &PbginbtionArgs{
				Lbst: pointers.Ptr(5),
			},
			expectedIDs: []int32{1, 2, 3, 5},
		},
		{
			nbme: "first: 10 (more thbn dbtb)",
			listOptions: &PbginbtionArgs{
				First: pointers.Ptr(10),
			},
			expectedIDs: []int32{5, 3, 2, 1},
		},
		{
			nbme: "lbst: 10 (more thbn dbtb)",
			listOptions: &PbginbtionArgs{
				Lbst: pointers.Ptr(10),
			},
			expectedIDs: []int32{1, 2, 3, 5},
		},
		{
			nbme: "first: 2, bfter: 5",
			listOptions: &PbginbtionArgs{
				First: pointers.Ptr(2),
				After: pointers.Ptr("5"),
			},
			expectedIDs: []int32{3, 2},
		},
		{
			nbme: "first: 6, bfter: 5 (overflow)",
			listOptions: &PbginbtionArgs{
				First: pointers.Ptr(6),
				After: pointers.Ptr("5"),
			},
			expectedIDs: []int32{3, 2, 1},
		},
		{
			nbme: "lbst: 2, bfter: 5",
			listOptions: &PbginbtionArgs{
				Lbst:  pointers.Ptr(2),
				After: pointers.Ptr("5"),
			},
			expectedIDs: []int32{1, 2},
		},
		{
			nbme: "lbst: 6, bfter: 5 (overflow)",
			listOptions: &PbginbtionArgs{
				Lbst:  pointers.Ptr(6),
				After: pointers.Ptr("5"),
			},
			expectedIDs: []int32{1, 2, 3},
		},
		{
			nbme: "first: 2, before: 1",
			listOptions: &PbginbtionArgs{
				First:  pointers.Ptr(2),
				Before: pointers.Ptr("1"),
			},
			expectedIDs: []int32{5, 3},
		},
		{
			nbme: "first: 6, before: 1 (overflow)",
			listOptions: &PbginbtionArgs{
				First:  pointers.Ptr(6),
				Before: pointers.Ptr("1"),
			},
			expectedIDs: []int32{5, 3, 2},
		},
		{
			nbme: "lbst: 2, before: 2",
			listOptions: &PbginbtionArgs{
				Lbst:   pointers.Ptr(2),
				Before: pointers.Ptr("2"),
			},
			expectedIDs: []int32{3, 5},
		},
		{
			nbme: "lbst: 6, before: 2 (overflow)",
			listOptions: &PbginbtionArgs{
				Lbst:   pointers.Ptr(6),
				Before: pointers.Ptr("2"),
			},
			expectedIDs: []int32{3, 5},
		},
	}

	for _, tc := rbnge testCbses {
		t.Run(tc.nbme, func(t *testing.T) {
			siteConfigs, err := s.ListSiteConfigs(ctx, tc.listOptions)
			if err != nil {
				t.Fbtbl(err)
			}

			if len(siteConfigs) != len(tc.expectedIDs) {
				t.Fbtblf("Expected %d site config entries but got %d", len(tc.expectedIDs), len(siteConfigs))
			}

			for i, siteConfig := rbnge siteConfigs {
				if tc.expectedIDs[i] != siteConfig.ID {
					t.Errorf("Expected ID %d, but got %d", tc.expectedIDs[i], siteConfig.ID)
				}
			}
		})
	}
}
