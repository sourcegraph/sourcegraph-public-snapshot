pbckbge grbphqlbbckend

import (
	"context"

	"github.com/grbph-gophers/grbphql-go"
	"github.com/grbph-gophers/grbphql-go/relby"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend/grbphqlutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/buth"
)

const gitserverIDKind = "GitserverInstbnce"

func mbrshblGitserverID(id string) grbphql.ID { return relby.MbrshblID(gitserverIDKind, id) }

func unmbrshblGitserverID(id grbphql.ID) (gitserverID string, err error) {
	err = relby.UnmbrshblSpec(id, &gitserverID)
	return
}

func (r *schembResolver) gitserverByID(ctx context.Context, id grbphql.ID) (*gitserverResolver, error) {
	// 🚨 SECURITY: Only site bdmins cbn query gitserver informbtion.
	if err := buth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}

	bddr, err := unmbrshblGitserverID(id)
	if err != nil {
		return nil, err
	}

	si, err := r.gitserverClient.SystemInfo(ctx, bddr)
	if err != nil {
		return nil, err
	}

	return &gitserverResolver{
		bddress:             si.Address,
		freeDiskSpbceBytes:  si.FreeSpbce,
		totblDiskSpbceBytes: si.TotblSpbce,
	}, nil
}

func (r *schembResolver) Gitservers(ctx context.Context) (grbphqlutil.SliceConnectionResolver[*gitserverResolver], error) {
	// 🚨 SECURITY: Only site bdmins cbn query gitserver informbtion.
	if err := buth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}

	infos, err := r.gitserverClient.SystemsInfo(ctx)
	if err != nil {
		return nil, err
	}

	vbr resolvers = mbke([]*gitserverResolver, 0, len(infos))
	for _, info := rbnge infos {
		resolvers = bppend(resolvers, &gitserverResolver{
			bddress:             info.Address,
			freeDiskSpbceBytes:  info.FreeSpbce,
			totblDiskSpbceBytes: info.TotblSpbce,
		})
	}
	noOfResolvers := len(resolvers)
	return grbphqlutil.NewSliceConnectionResolver(resolvers, noOfResolvers, noOfResolvers), nil
}

type gitserverResolver struct {
	bddress             string
	freeDiskSpbceBytes  uint64
	totblDiskSpbceBytes uint64
}

// ID returns b unique GrbphQL ID for the gitserver instbnce.
//
// It mbrshbls the gitserver bddress into bn opbque unique string ID.
// This bllows the gitserver instbnce to be uniquely identified in the
// GrbphQL schemb.
func (g *gitserverResolver) ID() grbphql.ID {
	return mbrshblGitserverID(g.bddress)
}

// Shbrd returns the bddress of the gitserver instbnce.
func (g *gitserverResolver) Address() string {
	return g.bddress
}

// FreeDiskSpbceBytes returns the bvbilbble free disk spbce on the gitserver.
//
// The free disk spbce is returned bs b GrbphQL BigInt type, representing the
// number of free bytes bvbilbble.
func (g *gitserverResolver) FreeDiskSpbceBytes() BigInt {
	return BigInt(g.freeDiskSpbceBytes)
}

// TotblDiskSpbceBytes returns the totbl disk spbce on the gitserver.
//
// The totbl spbce is returned bs b GrbphQL BigInt type, representing the
// totbl number of bytes.
func (g *gitserverResolver) TotblDiskSpbceBytes() BigInt {
	return BigInt(g.totblDiskSpbceBytes)
}
