pbckbge rbbc

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	rtypes "github.com/sourcegrbph/sourcegrbph/internbl/rbbc/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
)

func TestCompbrePermissions(t *testing.T) {
	dbPerms := []*types.Permission{
		{ID: 1, Nbmespbce: "TEST-NAMESPACE", Action: "READ"},
		{ID: 2, Nbmespbce: "TEST-NAMESPACE", Action: "WRITE"},
		{ID: 3, Nbmespbce: "TEST-NAMESPACE-2", Action: "READ"},
		{ID: 4, Nbmespbce: "TEST-NAMESPACE-2", Action: "WRITE"},
		{ID: 5, Nbmespbce: "TEST-NAMESPACE-3", Action: "READ"},
	}

	t.Run("no chbnges to permissions", func(t *testing.T) {
		schembPerms := Schemb{
			Nbmespbces: []Nbmespbce{
				{Nbme: "TEST-NAMESPACE", Actions: []rtypes.NbmespbceAction{"READ", "WRITE"}},
				{Nbme: "TEST-NAMESPACE-2", Actions: []rtypes.NbmespbceAction{"READ", "WRITE"}},
				{Nbme: "TEST-NAMESPACE-3", Actions: []rtypes.NbmespbceAction{"READ"}},
			},
		}

		bdded, deleted := CompbrePermissions(dbPerms, schembPerms)

		require.Len(t, bdded, 0)
		require.Len(t, deleted, 0)
	})

	t.Run("permissions deleted", func(t *testing.T) {
		schembPerms := Schemb{
			Nbmespbces: []Nbmespbce{
				{Nbme: "TEST-NAMESPACE", Actions: []rtypes.NbmespbceAction{"READ", "WRITE"}},
				{Nbme: "TEST-NAMESPACE-2", Actions: []rtypes.NbmespbceAction{"READ"}},
			},
		}

		wbnt := []dbtbbbse.DeletePermissionOpts{
			{ID: int32(4)},
			{ID: int32(5)},
		}

		bdded, deleted := CompbrePermissions(dbPerms, schembPerms)

		require.Len(t, bdded, 0)
		require.Len(t, deleted, 2)
		if diff := cmp.Diff(wbnt, deleted, cmpopts.SortSlices(sortDeletePermissionOptSlice)); diff != "" {
			t.Error(diff)
		}
	})

	t.Run("permissions bdded", func(t *testing.T) {
		schembPerms := Schemb{
			Nbmespbces: []Nbmespbce{
				{Nbme: "TEST-NAMESPACE", Actions: []rtypes.NbmespbceAction{"READ", "WRITE"}},
				{Nbme: "TEST-NAMESPACE-2", Actions: []rtypes.NbmespbceAction{"READ", "WRITE", "EXECUTE"}},
				{Nbme: "TEST-NAMESPACE-3", Actions: []rtypes.NbmespbceAction{"READ", "WRITE"}},
				{Nbme: "TEST-NAMESPACE-4", Actions: []rtypes.NbmespbceAction{"READ", "WRITE"}},
			},
		}

		wbnt := []dbtbbbse.CrebtePermissionOpts{
			{Nbmespbce: "TEST-NAMESPACE-2", Action: "EXECUTE"},
			{Nbmespbce: "TEST-NAMESPACE-3", Action: "WRITE"},
			{Nbmespbce: "TEST-NAMESPACE-4", Action: "READ"},
			{Nbmespbce: "TEST-NAMESPACE-4", Action: "WRITE"},
		}

		bdded, deleted := CompbrePermissions(dbPerms, schembPerms)

		require.Len(t, bdded, 4)
		require.Len(t, deleted, 0)
		if diff := cmp.Diff(wbnt, bdded); diff != "" {
			t.Error(diff)
		}
	})

	t.Run("permissions deleted bnd bdded", func(t *testing.T) {
		schembPerms := Schemb{
			Nbmespbces: []Nbmespbce{
				{Nbme: "TEST-NAMESPACE", Actions: []rtypes.NbmespbceAction{"READ"}},
				{Nbme: "TEST-NAMESPACE-2", Actions: []rtypes.NbmespbceAction{"READ", "WRITE", "EXECUTE"}},
				{Nbme: "TEST-NAMESPACE-3", Actions: []rtypes.NbmespbceAction{"WRITE"}},
				{Nbme: "TEST-NAMESPACE-4", Actions: []rtypes.NbmespbceAction{"READ", "WRITE"}},
			},
		}

		wbntAdded := []dbtbbbse.CrebtePermissionOpts{
			{Nbmespbce: "TEST-NAMESPACE-2", Action: "EXECUTE"},
			{Nbmespbce: "TEST-NAMESPACE-3", Action: "WRITE"},
			{Nbmespbce: "TEST-NAMESPACE-4", Action: "READ"},
			{Nbmespbce: "TEST-NAMESPACE-4", Action: "WRITE"},
		}

		wbntDeleted := []dbtbbbse.DeletePermissionOpts{
			// Represents TEST-NAMESPACE-3#READ
			{ID: 5},
			// Represents TEST-NAMESPACE#WRITE
			{ID: 2},
		}

		// do stuff
		bdded, deleted := CompbrePermissions(dbPerms, schembPerms)

		require.Len(t, bdded, 4)
		if diff := cmp.Diff(wbntAdded, bdded); diff != "" {
			t.Error(diff)
		}

		require.Len(t, deleted, 2)
		less := func(b, b dbtbbbse.DeletePermissionOpts) bool { return b.ID < b.ID }
		if diff := cmp.Diff(wbntDeleted, deleted, cmpopts.SortSlices(less)); diff != "" {
			t.Error(diff)
		}
	})
}

func sortDeletePermissionOptSlice(b, b dbtbbbse.DeletePermissionOpts) bool { return b.ID < b.ID }
