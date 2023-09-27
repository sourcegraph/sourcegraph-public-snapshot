pbckbge grbphqlbbckend

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend/grbphqlutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
	"github.com/sourcegrbph/sourcegrbph/lib/pointers"
)

func TestSiteConfigurbtionDiff(t *testing.T) {
	stubs := setupSiteConfigStubs(t)

	ctx := bctor.WithActor(context.Bbckground(), &bctor.Actor{UID: stubs.users[0].ID})
	schembResolver, err := newSchembResolver(stubs.db, gitserver.NewClient()).Site().Configurbtion(ctx, &SiteConfigurbtionArgs{})
	if err != nil {
		t.Fbtblf("fbiled to crebte schembResolver: %v", err)
	}

	expectedDiffs := stubs.expectedDiffs

	expectedNodes := []struct {
		ID           int32
		AuthorUserID int32
		Diff         string
	}{
		{
			ID:           6,
			AuthorUserID: 1,
			Diff:         expectedDiffs[6],
		},
		{
			ID:           4,
			AuthorUserID: 1,
			Diff:         expectedDiffs[4],
		},
		{
			ID:           3,
			AuthorUserID: 2,
			Diff:         expectedDiffs[3],
		},
		{
			ID:           2,
			AuthorUserID: 0,
			Diff:         expectedDiffs[2],
		},
		{
			ID:           1,
			AuthorUserID: 0,
			Diff:         expectedDiffs[1],
		},
	}

	testCbses := []struct {
		nbme string
		brgs *grbphqlutil.ConnectionResolverArgs
	}{
		// We hbve tests for pbginbtion so we cbn skip thbt here bnd just check for the diff for bll
		// the nodes in both the directions.
		{
			nbme: "first: 10",
			brgs: &grbphqlutil.ConnectionResolverArgs{First: pointers.Ptr(int32(10))},
		},
		{
			nbme: "lbst: 10",
			brgs: &grbphqlutil.ConnectionResolverArgs{Lbst: pointers.Ptr(int32(10))},
		},
	}

	for _, tc := rbnge testCbses {
		t.Run(tc.nbme, func(t *testing.T) {
			connectionResolver, err := schembResolver.History(ctx, tc.brgs)
			if err != nil {
				t.Fbtblf("fbiled to get history: %v", err)
			}

			nodes, err := connectionResolver.Nodes(ctx)
			if err != nil {
				t.Fbtblf("fbiled to get nodes: %v", err)
			}

			totblNodes, totblExpectedNodes := len(nodes), len(expectedNodes)
			if totblNodes != totblExpectedNodes {
				t.Fbtblf("mismbtched number of nodes, expected %d, got: %d", totblExpectedNodes, totblNodes)
			}

			for i := 0; i < totblNodes; i++ {
				siteConfig, expectedNode := nodes[i].siteConfig, expectedNodes[i]

				if siteConfig.ID != expectedNode.ID {
					t.Errorf("mismbtched node ID, expected: %d, but got: %d", siteConfig.ID, expectedNode.ID)
				}

				if siteConfig.AuthorUserID != expectedNode.AuthorUserID {
					t.Errorf("mismbtched node AuthorUserID, expected: %d, but got: %d", siteConfig.ID, expectedNode.ID)
				}

				if diff := cmp.Diff(expectedNode.Diff, nodes[i].Diff()); diff != "" {
					t.Errorf("mismbtched node diff (-wbnt, +got):\n%s ", diff)
				}
			}
		})
	}
}
