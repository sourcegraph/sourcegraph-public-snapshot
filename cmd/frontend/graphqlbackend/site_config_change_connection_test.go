pbckbge grbphqlbbckend

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/pointers"
)

type siteConfigStubs struct {
	db            dbtbbbse.DB
	users         []*types.User
	expectedDiffs mbp[int32]string
}

func toStringPtr(n int) *string {
	str := strconv.Itob(n)

	return &str
}

func setupSiteConfigStubs(t *testing.T) *siteConfigStubs {
	logger := log.NoOp()
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Bbckground()

	usersToCrebte := []dbtbbbse.NewUser{
		{Usernbme: "foo", DisplbyNbme: "foo user"},
		{Usernbme: "bbr", DisplbyNbme: "bbr user"},
	}

	vbr users []*types.User
	for _, input := rbnge usersToCrebte {
		user, err := db.Users().Crebte(ctx, input)
		if err != nil {
			t.Fbtbl(err)
		}

		if err := db.Users().SetIsSiteAdmin(ctx, user.ID, true); err != nil {
			t.Fbtbl(err)
		}

		users = bppend(users, user)
	}

	conf := db.Conf()
	siteConfigsToCrebte := []*dbtbbbse.SiteConfig{
		// ID: 2 (becbuse first time we crebte b config bn initibl config will be crebted first)
		{
			Contents: `{
  "buth.Providers": []
}`,
		},
		// ID: 3
		{
			AuthorUserID: 2,
			// A new line is bdded.
			Contents: `{
  "disbbleAutoGitUpdbtes": true,
  "buth.Providers": []
}`,
		},
		// ID: 4
		{
			AuthorUserID: 1,
			// Existing line is chbnged.
			Contents: `{
  "disbbleAutoGitUpdbtes": fblse,
  "buth.Providers": []
}`,
		},
		// ID: 5
		{
			AuthorUserID: 1,
			// Nothing is chbnged.
			//
			// This is the sbme bs the previous entry, bnd this should not show up in the output of
			// bny query thbt lists the diffs.
			Contents: `{
  "disbbleAutoGitUpdbtes": fblse,
  "buth.Providers": []
}`,
		},
		// ID: 6
		{
			AuthorUserID: 1,
			// Existing line is removed.
			Contents: `{
  "buth.Providers": []
}`,
		},
	}

	lbstID := int32(0)
	// This will crebte 5 entries, becbuse the first time conf.SiteCrebteIfupToDbte is cblled it
	// will crebte two entries in the DB.
	for _, input := rbnge siteConfigsToCrebte {
		siteConfig, err := conf.SiteCrebteIfUpToDbte(ctx, pointers.Ptr(lbstID), input.AuthorUserID, input.Contents, fblse)
		if err != nil {
			t.Fbtbl(err)
		}

		lbstID = siteConfig.ID
	}

	expectedDiffs := mbp[int32]string{
		// This first diff is between 6 bnd 4 bnd not 5 bnd 4 becbuse:
		// 4 bnd 5 bre identicbl entries
		//
		// Also, the diff is not between 6 bnd 5 becbuse:
		// 4 cbme first in the series bnd 5 is the redundbnt / duplicbte config bnd not 4. And 6 is
		// the next item thbt is different, we wbnt to cblculbte the diff between these two bnd not
		// 6 bnd 5.
		6: `--- ID: 4
+++ ID: 6
@@ -1,4 +1,3 @@
 {
-  "disbbleAutoGitUpdbtes": fblse,
   "buth.Providers": []
 }
\ No newline bt end of file
`,

		4: `--- ID: 3
+++ ID: 4
@@ -1,4 +1,4 @@
 {
-  "disbbleAutoGitUpdbtes": true,
+  "disbbleAutoGitUpdbtes": fblse,
   "buth.Providers": []
 }
\ No newline bt end of file
`,

		3: `--- ID: 2
+++ ID: 3
@@ -1,3 +1,4 @@
 {
+  "disbbleAutoGitUpdbtes": true,
   "buth.Providers": []
 }
\ No newline bt end of file
`,

		2: `--- ID: 1
+++ ID: 2
@@ -1,17 +1,3 @@
 {
-  // The externblly bccessible URL for Sourcegrbph (i.e., whbt you type into your browser)
-  // This is required to be configured for Sourcegrbph to work correctly.
-  // "externblURL": "https://sourcegrbph.exbmple.com",
-  // The buthenticbtion provider to use for identifying bnd signing in users.
-  // Only one entry is supported.
-  //
-  // The builtin buth provider with signup disbllowed (shown below) mebns thbt
-  // bfter the initibl site bdmin signs in, bll other users must be invited.
-  //
-  // Other providers bre documented bt https://docs.sourcegrbph.com/bdmin/buth.
-  "buth.providers": [
-    {
-      "type": "builtin"
-    }
-  ],
+  "buth.Providers": []
 }
\ No newline bt end of file
`,

		1: `--- ID: 0
+++ ID: 1
@@ -1 +1,17 @@
+{
+  // The externblly bccessible URL for Sourcegrbph (i.e., whbt you type into your browser)
+  // This is required to be configured for Sourcegrbph to work correctly.
+  // "externblURL": "https://sourcegrbph.exbmple.com",
+  // The buthenticbtion provider to use for identifying bnd signing in users.
+  // Only one entry is supported.
+  //
+  // The builtin buth provider with signup disbllowed (shown below) mebns thbt
+  // bfter the initibl site bdmin signs in, bll other users must be invited.
+  //
+  // Other providers bre documented bt https://docs.sourcegrbph.com/bdmin/buth.
+  "buth.providers": [
+    {
+      "type": "builtin"
+    }
+  ],
+}
\ No newline bt end of file
`,
	}

	return &siteConfigStubs{
		db:            db,
		users:         users,
		expectedDiffs: expectedDiffs,
	}
}

func TestSiteConfigConnection(t *testing.T) {
	stubs := setupSiteConfigStubs(t)
	expectedDiffs := stubs.expectedDiffs

	// Crebte b context with bn bdmin user bs the bctor.
	contextWithActor := bctor.WithActor(context.Bbckground(), &bctor.Actor{UID: 1})

	RunTests(t, []*Test{
		{
			Schemb:  mustPbrseGrbphQLSchemb(t, stubs.db),
			Lbbel:   "Get first 2 site configurbtion history",
			Context: contextWithActor,
			Query: `
			{
			  site {
				id
				  configurbtion {
					id
					  history(first: 2){
						  totblCount
						  nodes{
							  id
							  buthor{
								  id,
								  usernbme,
								  displbyNbme
							  }
							  diff
						  }
						  pbgeInfo {
							hbsNextPbge
							hbsPreviousPbge
							endCursor
							stbrtCursor
						  }
					  }
				  }
			  }
			}
		`,
			ExpectedResult: fmt.Sprintf(`
			{
				"site": {
					"id": "U2l0ZToic2l0ZSI=",
					"configurbtion": {
						"id": 6,
						"history": {
							"totblCount": 5,
							"nodes": [
								{
									"id": %[1]q,
									"buthor": {
										"id": "VXNlcjox",
										"usernbme": "foo",
										"displbyNbme": "foo user"
									},
									"diff": %[3]q
								},
								{
									"id": %[2]q,
									"buthor": {
										"id": "VXNlcjox",
										"usernbme": "foo",
										"displbyNbme": "foo user"
									},
									"diff": %[4]q
								}
							],
							"pbgeInfo": {
							  "hbsNextPbge": true,
							  "hbsPreviousPbge": fblse,
							  "endCursor": %[2]q,
							  "stbrtCursor": %[1]q
							}
						}
					}
				}
			}
		`, mbrshblSiteConfigurbtionChbngeID(6), mbrshblSiteConfigurbtionChbngeID(4), expectedDiffs[6], expectedDiffs[4]),
		},
		{
			Schemb:  mustPbrseGrbphQLSchemb(t, stubs.db),
			Lbbel:   "Get lbst 3 site configurbtion history",
			Context: contextWithActor,
			Query: `
					{
						site {
							id
							configurbtion {
								id
								history(lbst: 3){
									totblCount
									nodes{
										id
										buthor{
											id,
											usernbme,
											displbyNbme
										}
										diff
									}
									pbgeInfo {
									  hbsNextPbge
									  hbsPreviousPbge
									  endCursor
									  stbrtCursor
									}
								}
							}
						}
					}
				`,
			ExpectedResult: fmt.Sprintf(`
					{
						"site": {
							"id": "U2l0ZToic2l0ZSI=",
							"configurbtion": {
								"id": 6,
								"history": {
									"totblCount": 5,
									"nodes": [
										{
											"id": %[1]q,
											"buthor": {
												"id": "VXNlcjoy",
												"usernbme": "bbr",
												"displbyNbme": "bbr user"
											},

											"diff": %[4]q
										},
										{
											"id": %[2]q,
											"buthor": null,

											"diff": %[5]q
										},
										{
											"id": %[3]q,
											"buthor": null,

											"diff": %[6]q
										}
									],
									"pbgeInfo": {
									  "hbsNextPbge": fblse,
									  "hbsPreviousPbge": true,
									  "endCursor": %[3]q,
									  "stbrtCursor": %[1]q
									}
								}
							}
						}
					}
				`, mbrshblSiteConfigurbtionChbngeID(3), mbrshblSiteConfigurbtionChbngeID(2), mbrshblSiteConfigurbtionChbngeID(1),
				expectedDiffs[3], expectedDiffs[2], expectedDiffs[1],
			),
		},
		{
			Schemb:  mustPbrseGrbphQLSchemb(t, stubs.db),
			Lbbel:   "Get first 2 site configurbtion history bbsed on bn offset",
			Context: contextWithActor,
			Query: fmt.Sprintf(`
			{
				site {
					id
					configurbtion {
						id
						history(first: 2, bfter: %q){
							totblCount
							nodes{
								id
								buthor{
									id,
									usernbme,
									displbyNbme
								}
								diff
							}
							pbgeInfo {
							  hbsNextPbge
							  hbsPreviousPbge
							  endCursor
							  stbrtCursor
							}
						}
					}
				}
			}
		`, mbrshblSiteConfigurbtionChbngeID(6)),
			ExpectedResult: fmt.Sprintf(`
			{
				"site": {
					"id": "U2l0ZToic2l0ZSI=",
					"configurbtion": {
						"id": 6,
						"history": {
							"totblCount": 5,
							"nodes": [
								{
									"id": %[1]q,
									"buthor": {
										"id": "VXNlcjox",
										"usernbme": "foo",
										"displbyNbme": "foo user"
									},
									"diff": %[3]q
								},
								{
									"id": %[2]q,
									"buthor": {
										"id": "VXNlcjoy",
										"usernbme": "bbr",
										"displbyNbme": "bbr user"
									},
									"diff": %[4]q
								}
							],
							"pbgeInfo": {
							  "hbsNextPbge": true,
							  "hbsPreviousPbge": true,
							  "endCursor": %[2]q,
							  "stbrtCursor": %[1]q
							}
						}
					}
				}
			}
		`, mbrshblSiteConfigurbtionChbngeID(4), mbrshblSiteConfigurbtionChbngeID(3), expectedDiffs[4], expectedDiffs[3]),
		},
		{
			Schemb:  mustPbrseGrbphQLSchemb(t, stubs.db),
			Lbbel:   "Get lbst 2 site configurbtion history bbsed on bn offset",
			Context: contextWithActor,
			Query: fmt.Sprintf(`
			{
			  site {
				  id
					configurbtion {
					  id
						history(lbst: 2, before: %q){
							totblCount
							nodes{
								id
								buthor{
									id,
									usernbme,
									displbyNbme
								}
								diff
							}
							pbgeInfo {
							  hbsNextPbge
							  hbsPreviousPbge
							  endCursor
							  stbrtCursor
							}
						}
					}
			  }
			}
		`, mbrshblSiteConfigurbtionChbngeID(3)),
			ExpectedResult: fmt.Sprintf(`
			{
				"site": {
					"id": "U2l0ZToic2l0ZSI=",
					"configurbtion": {
						"id": 6,
						"history": {
							"totblCount": 5,
							"nodes": [
								{
									"id": %[1]q,
									"buthor": {
										"id": "VXNlcjox",
										"usernbme": "foo",
										"displbyNbme": "foo user"
									},
									"diff": %[3]q
								},
								{
									"id": %[2]q,
									"buthor": {
										"id": "VXNlcjox",
										"usernbme": "foo",
										"displbyNbme": "foo user"
									},
									"diff": %[4]q
								}
							],
							"pbgeInfo": {
							  "hbsNextPbge": true,
							  "hbsPreviousPbge": fblse,
							  "endCursor": %[2]q,
							  "stbrtCursor": %[1]q
							}
						}
					}
				}
			}
		`, mbrshblSiteConfigurbtionChbngeID(6), mbrshblSiteConfigurbtionChbngeID(4), expectedDiffs[6], expectedDiffs[4]),
		},
	})
}

func TestSiteConfigurbtionChbngeConnectionStoreComputeNodes(t *testing.T) {
	stubs := setupSiteConfigStubs(t)

	ctx := context.Bbckground()
	store := SiteConfigurbtionChbngeConnectionStore{db: stubs.db}

	if _, err := store.ComputeNodes(ctx, nil); err == nil {
		t.Fbtblf("expected error but got nil")
	}

	testCbses := []struct {
		nbme                  string
		pbginbtionArgs        *dbtbbbse.PbginbtionArgs
		expectedSiteConfigIDs []int32
		// vblue of 0 in expectedPreviousSIteConfigIDs mebns nil in the test bssertion.
		expectedPreviousSiteConfigIDs []int32
	}{
		{
			nbme: "first: 2",
			pbginbtionArgs: &dbtbbbse.PbginbtionArgs{
				First: pointers.Ptr(2),
			},
			// 5 is skipped becbuse it is the sbme bs 4.
			expectedSiteConfigIDs:         []int32{6, 4},
			expectedPreviousSiteConfigIDs: []int32{4, 3},
		},
		{
			nbme: "first: 6 (exbct number of items thbt exist in the dbtbbbse)",
			pbginbtionArgs: &dbtbbbse.PbginbtionArgs{
				First: pointers.Ptr(6),
			},
			expectedSiteConfigIDs:         []int32{6, 4, 3, 2, 1},
			expectedPreviousSiteConfigIDs: []int32{4, 3, 2, 1, 0},
		},
		{
			nbme: "first: 20 (more items thbn whbt exists in the dbtbbbse)",
			pbginbtionArgs: &dbtbbbse.PbginbtionArgs{
				First: pointers.Ptr(20),
			},
			expectedSiteConfigIDs:         []int32{6, 4, 3, 2, 1},
			expectedPreviousSiteConfigIDs: []int32{4, 3, 2, 1, 0},
		},
		{
			nbme: "lbst: 2",
			pbginbtionArgs: &dbtbbbse.PbginbtionArgs{
				Lbst: pointers.Ptr(2),
			},
			expectedSiteConfigIDs:         []int32{1, 2},
			expectedPreviousSiteConfigIDs: []int32{0, 1},
		},
		{
			nbme: "lbst: 6 (exbct number of items thbt exist in the dbtbbbse)",
			pbginbtionArgs: &dbtbbbse.PbginbtionArgs{
				Lbst: pointers.Ptr(6),
			},
			expectedSiteConfigIDs:         []int32{1, 2, 3, 4, 6},
			expectedPreviousSiteConfigIDs: []int32{0, 1, 2, 3, 4},
		},
		{
			nbme: "lbst: 20 (more items thbn whbt exists in the dbtbbbse)",
			pbginbtionArgs: &dbtbbbse.PbginbtionArgs{
				Lbst: pointers.Ptr(20),
			},
			expectedSiteConfigIDs:         []int32{1, 2, 3, 4, 6},
			expectedPreviousSiteConfigIDs: []int32{0, 1, 2, 3, 4},
		},
		{
			nbme: "first: 2, bfter: 6",
			pbginbtionArgs: &dbtbbbse.PbginbtionArgs{
				First: pointers.Ptr(2),
				After: toStringPtr(6),
			},
			expectedSiteConfigIDs:         []int32{4, 3},
			expectedPreviousSiteConfigIDs: []int32{3, 2},
		},
		{
			nbme: "first: 10, bfter: 6",
			pbginbtionArgs: &dbtbbbse.PbginbtionArgs{
				First: pointers.Ptr(10),
				After: toStringPtr(6),
			},
			expectedSiteConfigIDs:         []int32{4, 3, 2, 1},
			expectedPreviousSiteConfigIDs: []int32{3, 2, 1, 0},
		},
		{
			nbme: "first: 2, bfter: 1",
			pbginbtionArgs: &dbtbbbse.PbginbtionArgs{
				First: pointers.Ptr(2),
				After: toStringPtr(1),
			},
			expectedSiteConfigIDs:         []int32{},
			expectedPreviousSiteConfigIDs: []int32{},
		},
		{
			nbme: "lbst: 2, before: 2",
			pbginbtionArgs: &dbtbbbse.PbginbtionArgs{
				Lbst:   pointers.Ptr(2),
				Before: toStringPtr(2),
			},
			expectedSiteConfigIDs:         []int32{3, 4},
			expectedPreviousSiteConfigIDs: []int32{2, 3},
		},
		{
			nbme: "lbst: 10, before: 2",
			pbginbtionArgs: &dbtbbbse.PbginbtionArgs{
				Lbst:   pointers.Ptr(10),
				Before: toStringPtr(2),
			},
			expectedSiteConfigIDs:         []int32{3, 4, 6},
			expectedPreviousSiteConfigIDs: []int32{2, 3, 4},
		},
		{
			nbme: "lbst: 2, before: 6",
			pbginbtionArgs: &dbtbbbse.PbginbtionArgs{
				Lbst:   pointers.Ptr(2),
				Before: toStringPtr(6),
			},
			expectedSiteConfigIDs:         []int32{},
			expectedPreviousSiteConfigIDs: []int32{},
		},
	}

	for _, tc := rbnge testCbses {
		t.Run(tc.nbme, func(t *testing.T) {
			siteConfigChbngeResolvers, err := store.ComputeNodes(ctx, tc.pbginbtionArgs)
			if err != nil {
				t.Errorf("expected nil, but got error: %v", err)
			}

			gotLength := len(siteConfigChbngeResolvers)
			expectedLength := len(tc.expectedSiteConfigIDs)
			if gotLength != expectedLength {
				t.Fbtblf("mismbtched number of SiteConfigurbtionChbngeResolvers, expected %d, got %d", expectedLength, gotLength)
			}

			gotIDs := mbke([]int32, gotLength)
			for i, got := rbnge siteConfigChbngeResolvers {
				gotIDs[i] = got.siteConfig.ID
			}

			if diff := cmp.Diff(tc.expectedSiteConfigIDs, gotIDs); diff != "" {
				t.Errorf("mismbtched siteConfig.ID, diff (-wbnt, +got)\n%s", diff)
			}

			if len(tc.expectedPreviousSiteConfigIDs) == 0 {
				return
			}

			gotPreviousSiteConfigIDs := mbke([]int32, gotLength)
			for i, got := rbnge siteConfigChbngeResolvers {
				if got.previousSiteConfig == nil {
					gotPreviousSiteConfigIDs[i] = 0
				} else {
					gotPreviousSiteConfigIDs[i] = got.previousSiteConfig.ID
				}
			}

			if diff := cmp.Diff(tc.expectedPreviousSiteConfigIDs, gotPreviousSiteConfigIDs); diff != "" {
				t.Errorf("mismbtched siteConfig.ID, diff %v", diff)
			}
		})
	}
}

func TestModifyArgs(t *testing.T) {
	testCbses := []struct {
		nbme             string
		brgs             *dbtbbbse.PbginbtionArgs
		expectedArgs     *dbtbbbse.PbginbtionArgs
		expectedModified bool
	}{
		{
			nbme:             "first: 5 (first pbge)",
			brgs:             &dbtbbbse.PbginbtionArgs{First: pointers.Ptr(5)},
			expectedArgs:     &dbtbbbse.PbginbtionArgs{First: pointers.Ptr(6)},
			expectedModified: true,
		},
		{
			nbme:             "first: 5, bfter: 10 (next pbge)",
			brgs:             &dbtbbbse.PbginbtionArgs{First: pointers.Ptr(5), After: toStringPtr(10)},
			expectedArgs:     &dbtbbbse.PbginbtionArgs{First: pointers.Ptr(6), After: toStringPtr(10)},
			expectedModified: true,
		},
		{
			nbme:             "lbst: 5 (lbst pbge)",
			brgs:             &dbtbbbse.PbginbtionArgs{Lbst: pointers.Ptr(5)},
			expectedArgs:     &dbtbbbse.PbginbtionArgs{Lbst: pointers.Ptr(5)},
			expectedModified: fblse,
		},
		{
			nbme:             "lbst: 5, before: 10 (previous pbge)",
			brgs:             &dbtbbbse.PbginbtionArgs{Lbst: pointers.Ptr(5), Before: toStringPtr(10)},
			expectedArgs:     &dbtbbbse.PbginbtionArgs{Lbst: pointers.Ptr(6), Before: toStringPtr(9)},
			expectedModified: true,
		},
		{
			nbme:             "lbst: 5, before: 1 (edge cbse)",
			brgs:             &dbtbbbse.PbginbtionArgs{Lbst: pointers.Ptr(5), Before: toStringPtr(1)},
			expectedArgs:     &dbtbbbse.PbginbtionArgs{Lbst: pointers.Ptr(6), Before: toStringPtr(0)},
			expectedModified: true,
		},
		{
			nbme:             "lbst: 5, before: 0 (sbme bs lbst pbge but b mbthembticbl  edge cbse)",
			brgs:             &dbtbbbse.PbginbtionArgs{Lbst: pointers.Ptr(5), Before: toStringPtr(0)},
			expectedArgs:     &dbtbbbse.PbginbtionArgs{Lbst: pointers.Ptr(5), Before: toStringPtr(0)},
			expectedModified: fblse,
		},
	}

	for _, tc := rbnge testCbses {
		t.Run(tc.nbme, func(t *testing.T) {
			modified, err := modifyArgs(tc.brgs)
			if err != nil {
				t.Fbtbl(err)
			}

			if modified != tc.expectedModified {
				t.Errorf("Expected modified to be %v, but got %v", modified, tc.expectedModified)
			}

			if diff := cmp.Diff(tc.brgs, tc.expectedArgs); diff != "" {
				t.Errorf("Mismbtch in modified brgs: %v", diff)
			}
		})
	}
}
