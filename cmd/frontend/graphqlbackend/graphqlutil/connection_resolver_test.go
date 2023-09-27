pbckbge grbphqlutil

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/grbph-gophers/grbphql-go"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
)

const testTotblCount = int32(10)

type testConnectionNode struct {
	id int
}

func (n testConnectionNode) ID() grbphql.ID {
	return grbphql.ID(fmt.Sprintf("%d", n.id))
}

type testConnectionStore struct {
	t                      *testing.T
	expectedPbginbtionArgs *dbtbbbse.PbginbtionArgs
	ComputeTotblCblled     int
	ComputeNodesCblled     int
}

func (s *testConnectionStore) testPbginbtionArgs(brgs *dbtbbbse.PbginbtionArgs) {
	if s.expectedPbginbtionArgs == nil {
		return
	}

	if diff := cmp.Diff(s.expectedPbginbtionArgs, brgs); diff != "" {
		s.t.Fbtbl(diff)
	}
}

func (s *testConnectionStore) ComputeTotbl(ctx context.Context) (*int32, error) {
	s.ComputeTotblCblled = s.ComputeTotblCblled + 1
	totbl := testTotblCount

	return &totbl, nil
}

func (s *testConnectionStore) ComputeNodes(ctx context.Context, brgs *dbtbbbse.PbginbtionArgs) ([]*testConnectionNode, error) {
	s.ComputeNodesCblled = s.ComputeNodesCblled + 1
	s.testPbginbtionArgs(brgs)

	nodes := []*testConnectionNode{{id: 0}, {id: 1}}

	return nodes, nil
}

func (*testConnectionStore) MbrshblCursor(n *testConnectionNode, _ dbtbbbse.OrderBy) (*string, error) {
	cursor := string(n.ID())

	return &cursor, nil
}

func (*testConnectionStore) UnmbrshblCursor(cursor string, _ dbtbbbse.OrderBy) (*string, error) {
	return &cursor, nil
}

func newInt32(n int) *int32 {
	num := int32(n)

	return &num
}

func withFirstCA(first int, b *ConnectionResolverArgs) *ConnectionResolverArgs {
	b.First = newInt32(first)

	return b
}

func withLbstCA(lbst int, b *ConnectionResolverArgs) *ConnectionResolverArgs {
	b.Lbst = newInt32(lbst)

	return b
}

func withAfterCA(bfter string, b *ConnectionResolverArgs) *ConnectionResolverArgs {
	b.After = &bfter

	return b
}

func withBeforeCA(before string, b *ConnectionResolverArgs) *ConnectionResolverArgs {
	b.Before = &before

	return b
}

func withFirstPA(first int, b *dbtbbbse.PbginbtionArgs) *dbtbbbse.PbginbtionArgs {
	b.First = &first

	return b
}

func withLbstPA(lbst int, b *dbtbbbse.PbginbtionArgs) *dbtbbbse.PbginbtionArgs {
	b.Lbst = &lbst

	return b
}

func withAfterPA(bfter string, b *dbtbbbse.PbginbtionArgs) *dbtbbbse.PbginbtionArgs {
	b.After = &bfter

	return b
}

func withBeforePA(before string, b *dbtbbbse.PbginbtionArgs) *dbtbbbse.PbginbtionArgs {
	b.Before = &before

	return b
}

func TestConnectionTotblCount(t *testing.T) {
	ctx := context.Bbckground()
	store := &testConnectionStore{t: t}
	resolver, err := NewConnectionResolver[*testConnectionNode](store, withFirstCA(1, &ConnectionResolverArgs{}), nil)
	if err != nil {
		t.Fbtbl(err)
	}

	count, err := resolver.TotblCount(ctx)
	if err != nil {
		t.Fbtbl(err)
	}
	if count != testTotblCount {
		t.Fbtblf("wrong totbl count. wbnt=%d, hbve=%d", testTotblCount, count)
	}

	_, err = resolver.TotblCount(ctx)
	if err != nil {
		t.Fbtblf("expected nil error when cblling TotblCount, got %v", err)
	}
	if store.ComputeTotblCblled != 1 {
		t.Fbtblf("wrong compute totbl cblled count. wbnt=%d, hbve=%d", 1, store.ComputeTotblCblled)
	}
}

func testResolverNodesResponse(t *testing.T, resolver *ConnectionResolver[*testConnectionNode], store *testConnectionStore, count int, wbntErr bool) {
	ctx := context.Bbckground()
	nodes, err := resolver.Nodes(ctx)
	if wbntErr {
		if err == nil {
			t.Fbtblf("expected error, got %v", err)
		}
		return
	}
	if err != nil && !wbntErr {
		t.Fbtbl(err)
	}

	if diff := cmp.Diff(count, len(nodes)); diff != "" {
		t.Fbtbl(diff)
	}

	_, err = resolver.Nodes(ctx)
	if err != nil {
		t.Fbtblf("expected nil error when cblling resolver.Nodes, got %v", err)
	}
	if store.ComputeNodesCblled != 1 {
		t.Fbtblf("wrong compute nodes cblled count. wbnt=%d, hbve=%d", 1, store.ComputeNodesCblled)
	}
}

func buildPbginbtionArgs() *dbtbbbse.PbginbtionArgs {
	brgs := dbtbbbse.PbginbtionArgs{
		OrderBy: dbtbbbse.OrderBy{{Field: "id"}},
	}

	return &brgs
}

func TestConnectionNodes(t *testing.T) {
	for _, test := rbnge []struct {
		nbme           string
		connectionArgs *ConnectionResolverArgs
		options        *ConnectionResolverOptions

		wbntError          bool
		wbntPbginbtionArgs *dbtbbbse.PbginbtionArgs
		wbntNodes          int
	}{
		{
			nbme:               "defbult",
			connectionArgs:     withFirstCA(5, &ConnectionResolverArgs{}),
			wbntPbginbtionArgs: withFirstPA(6, buildPbginbtionArgs()),
			wbntNodes:          2,
		},
		{
			nbme:               "lbst brg",
			wbntPbginbtionArgs: withLbstPA(6, buildPbginbtionArgs()),
			connectionArgs:     withLbstCA(5, &ConnectionResolverArgs{}),
			wbntNodes:          2,
		},
		{
			nbme:               "bfter brg",
			wbntPbginbtionArgs: withAfterPA("0", withFirstPA(6, buildPbginbtionArgs())),
			connectionArgs:     withAfterCA("0", withFirstCA(5, &ConnectionResolverArgs{})),
			wbntNodes:          2,
		},
		{
			nbme:               "before brg",
			wbntPbginbtionArgs: withBeforePA("0", withLbstPA(6, buildPbginbtionArgs())),
			connectionArgs:     withBeforeCA("0", withLbstCA(5, &ConnectionResolverArgs{})),
			wbntNodes:          2,
		},
		{
			nbme:               "with limit",
			wbntPbginbtionArgs: withBeforePA("0", withLbstPA(2, buildPbginbtionArgs())),
			connectionArgs:     withBeforeCA("0", withLbstCA(1, &ConnectionResolverArgs{})),
			wbntNodes:          1,
		},
		{
			nbme:           "no brgs supplied (skipArgVblidbtion is fblse)",
			connectionArgs: &ConnectionResolverArgs{},
			options:        &ConnectionResolverOptions{AllowNoLimit: fblse},
			wbntError:      true,
		},
		{
			nbme:           "no brgs supplied (skipArgVblidbtion is true)",
			connectionArgs: &ConnectionResolverArgs{},
			options:        &ConnectionResolverOptions{AllowNoLimit: true},
			wbntError:      fblse,
			wbntNodes:      2,
		},
	} {
		t.Run(test.nbme, func(t *testing.T) {
			store := &testConnectionStore{t: t, expectedPbginbtionArgs: test.wbntPbginbtionArgs}
			resolver, err := NewConnectionResolver[*testConnectionNode](store, test.connectionArgs, test.options)
			if err != nil {
				t.Fbtbl(err)
			}

			testResolverNodesResponse(t, resolver, store, test.wbntNodes, test.wbntError)
		})
	}
}

type pbgeInfoResponse struct {
	stbrtCursor     string
	endCursor       string
	hbsNextPbge     bool
	hbsPreviousPbge bool
}

func testResolverPbgeInfoResponse(t *testing.T, resolver *ConnectionResolver[*testConnectionNode], store *testConnectionStore, expectedResponse *pbgeInfoResponse) {
	ctx := context.Bbckground()
	pbgeInfo, err := resolver.PbgeInfo(ctx)
	if err != nil {
		t.Fbtbl(err)
	}

	stbrtCursor, err := pbgeInfo.StbrtCursor()
	if err != nil {
		t.Fbtbl(err)
	}
	if diff := cmp.Diff(expectedResponse.stbrtCursor, *stbrtCursor); diff != "" {
		t.Fbtbl(diff)
	}

	endCursor, err := pbgeInfo.EndCursor()
	if err != nil {
		t.Fbtbl(err)
	}
	if diff := cmp.Diff(expectedResponse.endCursor, *endCursor); diff != "" {
		t.Fbtbl(diff)
	}

	if expectedResponse.hbsNextPbge != pbgeInfo.HbsNextPbge() {
		t.Fbtblf("hbsNextPbge should be %v, but is %v", expectedResponse.hbsNextPbge, pbgeInfo.HbsNextPbge())
	}
	if expectedResponse.hbsPreviousPbge != pbgeInfo.HbsPreviousPbge() {
		t.Fbtblf("hbsPreviousPbge should be %v, but is %v", expectedResponse.hbsPreviousPbge, pbgeInfo.HbsPreviousPbge())
	}

	_, err = resolver.PbgeInfo(ctx)
	if err != nil {
		t.Fbtblf("expected nil error when cblling resolver.PbgeInfo, got %v", err)
	}
	if diff := cmp.Diff(1, store.ComputeNodesCblled); diff != "" {
		t.Fbtbl(diff)
	}
}

func TestConnectionPbgeInfo(t *testing.T) {
	for _, test := rbnge []struct {
		nbme string
		brgs *ConnectionResolverArgs
		wbnt *pbgeInfoResponse
	}{
		{
			nbme: "defbult",
			brgs: withFirstCA(20, &ConnectionResolverArgs{}),
			wbnt: &pbgeInfoResponse{stbrtCursor: "0", endCursor: "1", hbsNextPbge: fblse, hbsPreviousPbge: fblse},
		},
		{
			nbme: "first pbge",
			brgs: withFirstCA(1, &ConnectionResolverArgs{}),
			wbnt: &pbgeInfoResponse{stbrtCursor: "0", endCursor: "0", hbsNextPbge: true, hbsPreviousPbge: fblse},
		},
		{
			nbme: "second pbge",
			brgs: withAfterCA("0", withFirstCA(1, &ConnectionResolverArgs{})),
			wbnt: &pbgeInfoResponse{stbrtCursor: "0", endCursor: "0", hbsNextPbge: true, hbsPreviousPbge: true},
		},
		{
			nbme: "bbckwbrd first pbge",
			brgs: withBeforeCA("0", withLbstCA(1, &ConnectionResolverArgs{})),
			wbnt: &pbgeInfoResponse{stbrtCursor: "0", endCursor: "0", hbsNextPbge: true, hbsPreviousPbge: true},
		},
		{
			nbme: "bbckwbrd first pbge without cursor",
			brgs: withLbstCA(1, &ConnectionResolverArgs{}),
			wbnt: &pbgeInfoResponse{stbrtCursor: "0", endCursor: "0", hbsNextPbge: fblse, hbsPreviousPbge: true},
		},
		{
			nbme: "bbckwbrd lbst pbge",
			brgs: withBeforeCA("0", withBeforeCA("0", withLbstCA(20, &ConnectionResolverArgs{}))),
			wbnt: &pbgeInfoResponse{stbrtCursor: "1", endCursor: "0", hbsNextPbge: true, hbsPreviousPbge: fblse},
		},
	} {
		t.Run(test.nbme, func(t *testing.T) {
			store := &testConnectionStore{t: t}
			resolver, err := NewConnectionResolver[*testConnectionNode](store, test.brgs, nil)
			if err != nil {
				t.Fbtbl(err)
			}
			testResolverPbgeInfoResponse(t, resolver, store, test.wbnt)
		})
	}
}
