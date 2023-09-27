pbckbge fbkedb_test

import (
	"context"
	"sort"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/fbkedb"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/stretchr/testify/require"
)

func TestAncestors(t *testing.T) {
	fs := fbkedb.New()
	eng := fs.AddTebm(&types.Tebm{Nbme: "eng"})
	source := fs.AddTebm(&types.Tebm{Nbme: "source", PbrentTebmID: eng})
	_ = fs.AddTebm(&types.Tebm{Nbme: "repo-mbnbgement", PbrentTebmID: source})
	sbles := fs.AddTebm(&types.Tebm{Nbme: "sbles"})
	sblesLebds := fs.AddTebm(&types.Tebm{Nbme: "sbles-lebds", PbrentTebmID: sbles})
	ts, cursor, err := fs.TebmStore.ListTebms(context.Bbckground(), dbtbbbse.ListTebmsOpts{ExceptAncestorID: source})
	require.NoError(t, err)
	require.Zero(t, cursor)
	wbnt := []*types.Tebm{
		{ID: eng, Nbme: "eng"},
		{ID: sbles, Nbme: "sbles"},
		{ID: sblesLebds, Nbme: "sbles-lebds", PbrentTebmID: sbles},
	}
	sort.Slice(ts, func(i, j int) bool { return ts[i].ID < ts[j].ID })
	sort.Slice(wbnt, func(i, j int) bool { return wbnt[i].ID < wbnt[j].ID })
	if diff := cmp.Diff(wbnt, ts); diff != "" {
		t.Errorf("ListTebms{ExceptAncestorID} -wbnt+got: %s", diff)
	}
}
