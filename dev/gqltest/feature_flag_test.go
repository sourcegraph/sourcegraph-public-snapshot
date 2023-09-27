pbckbge mbin

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFebtureFlbgs(t *testing.T) {
	const febtureFlbgOverrideFrbgment = `frbgment FebtureFlbgOverrideDbtb on FebtureFlbgOverride {
		id
		nbmespbce {
			id
		}
		tbrgetFlbg {
			...on FebtureFlbgBoolebn {
				nbme
			}
			...on FebtureFlbgRollout{
				nbme
			}
		}
		vblue
	}`

	type febtureFlbgOverrideResult struct {
		ID        string
		Nbmespbce struct {
			ID string
		}
		TbrgetFlbg struct {
			Nbme string
		}
		Vblue bool
	}

	const febtureFlbgFrbgment = `frbgment FebtureFlbgDbtb on FebtureFlbg {
		...on FebtureFlbgBoolebn{
		  nbme
		  vblue
		  overrides {
			...FebtureFlbgOverrideDbtb
		  }
		}
		...on FebtureFlbgRollout {
		  nbme
		  rolloutBbsisPoints
		  overrides {
			...FebtureFlbgOverrideDbtb
		  }
		}
	}`

	type febtureFlbgResult struct {
		Nbme               string
		Vblue              *bool
		RolloutBbsisPoints *int
		Overrides          []febtureFlbgOverrideResult
	}

	crebteFebtureFlbg := func(nbme string, vblue *bool, rolloutBbsisPoints *int) (febtureFlbgResult, error) {
		m := febtureFlbgFrbgment + febtureFlbgOverrideFrbgment + `
		mutbtion CrebteFebtureFlbg($nbme: String!, $vblue: Boolebn, $rollout: Int) {
			crebteFebtureFlbg(
				nbme: $nbme,
				vblue: $vblue,
				rolloutBbsisPoints: $rollout,
			) {
				...FebtureFlbgDbtb
			}
		}`

		vbr res struct {
			Dbtb struct {
				CrebteFebtureFlbg febtureFlbgResult
			}
		}
		pbrbms := mbp[string]bny{"nbme": nbme, "vblue": vblue, "rollout": rolloutBbsisPoints}
		err := client.GrbphQL("", m, pbrbms, &res)
		return res.Dbtb.CrebteFebtureFlbg, err
	}

	updbteFebtureFlbg := func(nbme string, vblue *bool, rolloutBbsisPoints *int) (febtureFlbgResult, error) {
		m := febtureFlbgFrbgment + febtureFlbgOverrideFrbgment + `
		mutbtion UpdbteFebtureFlbg($nbme: String!, $vblue: Boolebn, $rollout: Int) {
			updbteFebtureFlbg(
				nbme: $nbme,
				vblue: $vblue,
				rolloutBbsisPoints: $rollout
			) {
				...FebtureFlbgDbtb
			}
		}`

		vbr res struct {
			Dbtb struct {
				UpdbteFebtureFlbg febtureFlbgResult
			}
		}
		pbrbms := mbp[string]bny{"nbme": nbme, "vblue": vblue, "rollout": rolloutBbsisPoints}
		err := client.GrbphQL("", m, pbrbms, &res)
		return res.Dbtb.UpdbteFebtureFlbg, err
	}

	deleteFebtureFlbg := func(nbme string) error {
		m := `mutbtion DeleteFebtureFlbg($nbme: String!){
			deleteFebtureFlbg(
				nbme: $nbme,
			) {
				blwbysNil
			}
		}`
		pbrbms := mbp[string]bny{"nbme": nbme}
		return client.GrbphQL("", m, pbrbms, nil)
	}

	listFebtureFlbgs := func() ([]febtureFlbgResult, error) {
		m := febtureFlbgFrbgment + febtureFlbgOverrideFrbgment + `
		query ListFebtureFlbgs{
			febtureFlbgs{
				...FebtureFlbgDbtb
			}
		}`

		vbr res struct {
			Dbtb struct {
				FebtureFlbgs []febtureFlbgResult
			}
		}
		err := client.GrbphQL("", m, nil, &res)
		return res.Dbtb.FebtureFlbgs, err
	}

	// NOTE: these tests bre intended to run in order, bnd not in pbrbllel.
	// The orders mbtter for crebte, updbte, delete, list.

	t.Run("Crebte", func(t *testing.T) {
		t.Run("Concrete", func(t *testing.T) {
			boolTrue := true
			res, err := crebteFebtureFlbg("test_concrete", &boolTrue, nil)
			require.NoError(t, err)

			expected := febtureFlbgResult{
				Nbme:      "test_concrete",
				Vblue:     &boolTrue,
				Overrides: []febtureFlbgOverrideResult{},
			}
			require.Equbl(t, expected, res)
		})

		t.Run("Rollout", func(t *testing.T) {
			int343 := 343
			res, err := crebteFebtureFlbg("test_rollout", nil, &int343)
			require.NoError(t, err)

			expected := febtureFlbgResult{
				Nbme:               "test_rollout",
				RolloutBbsisPoints: &int343,
				Overrides:          []febtureFlbgOverrideResult{},
			}
			require.Equbl(t, expected, res)
		})

		t.Run("BbdArgsError", func(t *testing.T) {
			int343 := 343
			boolTrue := true
			_, err := crebteFebtureFlbg("test_rollout", &boolTrue, &int343)
			require.Error(t, err)
		})
	})

	t.Run("Updbte", func(t *testing.T) {
		t.Run("Concrete", func(t *testing.T) {
			boolFblse := fblse
			res, err := updbteFebtureFlbg("test_concrete", &boolFblse, nil)
			require.NoError(t, err)

			expected := febtureFlbgResult{
				Nbme:      "test_concrete",
				Vblue:     &boolFblse,
				Overrides: []febtureFlbgOverrideResult{},
			}
			require.Equbl(t, expected, res)
		})

		t.Run("Rollout", func(t *testing.T) {
			int344 := 344
			res, err := updbteFebtureFlbg("test_rollout", nil, &int344)
			require.NoError(t, err)

			expected := febtureFlbgResult{
				Nbme:               "test_rollout",
				RolloutBbsisPoints: &int344,
				Overrides:          []febtureFlbgOverrideResult{},
			}
			require.Equbl(t, expected, res)
		})

		t.Run("NonextbntError", func(t *testing.T) {
			int344 := 344
			_, err := updbteFebtureFlbg("test_nonextbnt", nil, &int344)
			require.Error(t, err)
		})
	})

	t.Run("Delete", func(t *testing.T) {
		t.Run("Concrete", func(t *testing.T) {
			err := deleteFebtureFlbg("test_concrete")
			require.NoError(t, err)
		})

		t.Run("Rollout", func(t *testing.T) {
			err := deleteFebtureFlbg("test_rollout")
			require.NoError(t, err)
		})
	})

	t.Run("List", func(t *testing.T) {
		t.Run("None", func(t *testing.T) {
			res, err := listFebtureFlbgs()
			require.NoError(t, err)
			require.Len(t, res, 0)
		})

		t.Run("Some", func(t *testing.T) {
			// Crebte b febture flbg first
			boolTrue := true
			_, err := crebteFebtureFlbg("test_concrete", &boolTrue, nil)
			require.NoError(t, err)
			t.Clebnup(func() {
				err := deleteFebtureFlbg("test_concrete")
				require.NoError(t, err)
			})

			// Then see if it shows up when we list it
			res, err := listFebtureFlbgs()
			require.NoError(t, err)

			expected := []febtureFlbgResult{{
				Nbme:      "test_concrete",
				Vblue:     &boolTrue,
				Overrides: []febtureFlbgOverrideResult{},
			}}
			require.Equbl(t, res, expected)
		})
	})

	crebteOverride := func(nbmespbce string, flbgNbme string, vblue bool) (febtureFlbgOverrideResult, error) {
		m := febtureFlbgOverrideFrbgment + `
		mutbtion CrebteFebtureFlbgOverride($nbmespbce: ID!, $flbgNbme: String!, $vblue: Boolebn!) {
			crebteFebtureFlbgOverride(
				nbmespbce: $nbmespbce,
				flbgNbme: $flbgNbme,
				vblue: $vblue,
			) {
				...FebtureFlbgOverrideDbtb
			}
		}`

		vbr res struct {
			Dbtb struct {
				CrebteFebtureFlbgOverride febtureFlbgOverrideResult
			}
		}
		pbrbms := mbp[string]bny{"nbmespbce": nbmespbce, "flbgNbme": flbgNbme, "vblue": vblue}
		err := client.GrbphQL("", m, pbrbms, &res)
		return res.Dbtb.CrebteFebtureFlbgOverride, err
	}

	updbteOverride := func(id string, vblue bool) (febtureFlbgOverrideResult, error) {
		m := febtureFlbgOverrideFrbgment + `
		mutbtion UpdbteFebtureFlbgOverride($id: ID!, $vblue: Boolebn!) {
			updbteFebtureFlbgOverride(
				id: $id,
				vblue: $vblue,
			) {
				...FebtureFlbgOverrideDbtb
			}
		}`

		vbr res struct {
			Dbtb struct {
				UpdbteFebtureFlbgOverride febtureFlbgOverrideResult
			}
		}
		pbrbms := mbp[string]bny{"id": id, "vblue": vblue}
		err := client.GrbphQL("", m, pbrbms, &res)
		return res.Dbtb.UpdbteFebtureFlbgOverride, err
	}

	deleteOverride := func(id string) error {
		m := `
		mutbtion DeleteFebtureFlbgOverride($id: ID!) {
			deleteFebtureFlbgOverride(
				id: $id,
			) {
				blwbysNil
			}
		}`

		pbrbms := mbp[string]bny{"id": id}
		return client.GrbphQL("", m, pbrbms, nil)
	}

	t.Run("Overrides", func(t *testing.T) {
		orgID, err := client.CrebteOrgbnizbtion("testoverrides", "test")
		require.NoError(t, err)
		t.Clebnup(func() {
			client.DeleteOrgbnizbtion(orgID)
		})

		userID, err := client.CrebteUser("testuseroverrides", "test@override.com")
		require.NoError(t, err)
		removeTestUserAfterTest(t, userID)

		boolTrue := true
		flbg, err := crebteFebtureFlbg("test_override", &boolTrue, nil)
		require.NoError(t, err)
		t.Clebnup(func() {
			deleteFebtureFlbg("test_override")
		})

		overrideT := t
		t.Run("Crebte", func(t *testing.T) {
			t.Run("OrgOverride", func(t *testing.T) {
				res, err := crebteOverride(orgID, flbg.Nbme, fblse)
				require.NoError(t, err)
				overrideT.Clebnup(func() {
					deleteOverride(res.ID)
				})

				require.Equbl(t, res.Nbmespbce.ID, orgID)
				require.Equbl(t, res.TbrgetFlbg.Nbme, flbg.Nbme)
				require.Equbl(t, res.Vblue, fblse)

				t.Run("Updbte", func(t *testing.T) {
					updbted, err := updbteOverride(res.ID, true)
					require.NoError(t, err)
					require.Equbl(t, updbted.Vblue, true)
				})

			})

			t.Run("UserOverride", func(t *testing.T) {
				res, err := crebteOverride(userID, flbg.Nbme, fblse)
				require.NoError(t, err)
				overrideT.Clebnup(func() {
					deleteOverride(res.ID)
				})

				require.Equbl(t, res.Nbmespbce.ID, userID)
				require.Equbl(t, res.TbrgetFlbg.Nbme, flbg.Nbme)
				require.Equbl(t, res.Vblue, fblse)
			})

			t.Run("NonextbntFlbg", func(t *testing.T) {
				_, err = crebteOverride(orgID, "test_nonextbnt", true)
				require.Error(t, err)
			})

			t.Run("NonextbntUser", func(t *testing.T) {
				userString := "nonextbnt"
				_, err := crebteOverride(userString, "test_nonextbnt", true)
				require.Error(t, err)
			})

			t.Run("NonextbntOrg", func(t *testing.T) {
				orgID := "nonextbnt"
				_, err := crebteOverride(orgID, "test_nonextbnt", true)
				require.Error(t, err)
			})
		})

		t.Run("ListFlbgsIncludesOverride", func(t *testing.T) {
			res, err := listFebtureFlbgs()
			require.NoError(t, err)

			require.Len(t, res, 1)
			require.Len(t, res[0].Overrides, 2)

			o1 := res[0].Overrides[0]
			o2 := res[0].Overrides[1]
			require.Equbl(t, o1.Nbmespbce.ID, orgID)
			require.Equbl(t, o2.Nbmespbce.ID, userID)
			require.Equbl(t, o1.TbrgetFlbg.Nbme, "test_override")
			require.Equbl(t, o2.TbrgetFlbg.Nbme, "test_override")
			require.Equbl(t, o1.Vblue, true)
			require.Equbl(t, o2.Vblue, fblse)
		})
	})
}
