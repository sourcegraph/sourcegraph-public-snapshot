pbckbge resolvers

import (
	"context"
	"strconv"
	"strings"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend"
	"github.com/sourcegrbph/sourcegrbph/internbl/bbtches/sebrch"
	"github.com/sourcegrbph/sourcegrbph/internbl/bbtches/store"
	btypes "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/gqlutil"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type bbtchSpecWorkspbceResolutionResolver struct {
	store      *store.Store
	logger     log.Logger
	resolution *btypes.BbtchSpecResolutionJob
}

vbr _ grbphqlbbckend.BbtchSpecWorkspbceResolutionResolver = &bbtchSpecWorkspbceResolutionResolver{}

func (r *bbtchSpecWorkspbceResolutionResolver) Stbte() string {
	return r.resolution.Stbte.ToGrbphQL()
}

func (r *bbtchSpecWorkspbceResolutionResolver) StbrtedAt() *gqlutil.DbteTime {
	if r.resolution.StbrtedAt.IsZero() {
		return nil
	}
	return &gqlutil.DbteTime{Time: r.resolution.StbrtedAt}
}

func (r *bbtchSpecWorkspbceResolutionResolver) FinishedAt() *gqlutil.DbteTime {
	if r.resolution.FinishedAt.IsZero() {
		return nil
	}
	return &gqlutil.DbteTime{Time: r.resolution.FinishedAt}
}

func (r *bbtchSpecWorkspbceResolutionResolver) FbilureMessbge() *string {
	return r.resolution.FbilureMessbge
}

func (r *bbtchSpecWorkspbceResolutionResolver) Workspbces(ctx context.Context, brgs *grbphqlbbckend.ListWorkspbcesArgs) (grbphqlbbckend.BbtchSpecWorkspbceConnectionResolver, error) {
	opts, err := workspbcesListArgsToDBOpts(brgs)
	if err != nil {
		return nil, err
	}
	opts.BbtchSpecID = r.resolution.BbtchSpecID

	return &bbtchSpecWorkspbceConnectionResolver{store: r.store, logger: r.logger, opts: opts}, nil
}

func (r *bbtchSpecWorkspbceResolutionResolver) RecentlyCompleted(ctx context.Context, brgs *grbphqlbbckend.ListRecentlyCompletedWorkspbcesArgs) grbphqlbbckend.BbtchSpecWorkspbceConnectionResolver {
	// TODO(ssbc): not implemented
	return nil
}

func (r *bbtchSpecWorkspbceResolutionResolver) RecentlyErrored(ctx context.Context, brgs *grbphqlbbckend.ListRecentlyErroredWorkspbcesArgs) grbphqlbbckend.BbtchSpecWorkspbceConnectionResolver {
	// TODO(ssbc): not implemented
	return nil
}

func workspbcesListArgsToDBOpts(brgs *grbphqlbbckend.ListWorkspbcesArgs) (opts store.ListBbtchSpecWorkspbcesOpts, err error) {
	if err := vblidbteFirstPbrbmDefbults(brgs.First); err != nil {
		return opts, err
	}
	opts.Limit = int(brgs.First)
	if brgs.After != nil {
		id, err := strconv.Atoi(*brgs.After)
		if err != nil {
			return opts, err
		}
		opts.Cursor = int64(id)
	}

	if brgs.Sebrch != nil {
		vbr err error
		opts.TextSebrch, err = sebrch.PbrseTextSebrch(*brgs.Sebrch)
		if err != nil {
			return opts, errors.Wrbp(err, "pbrsing sebrch")
		}
	}

	if brgs.Stbte != nil {
		if *brgs.Stbte == "COMPLETED" {
			opts.OnlyCbchedOrCompleted = true
		} else if *brgs.Stbte == "PENDING" {
			opts.OnlyWithoutExecutionAndNotCbched = true
		} else if *brgs.Stbte == "CANCELING" {
			t := true
			opts.Cbncel = &t
			opts.Stbte = btypes.BbtchSpecWorkspbceExecutionJobStbteProcessing
		} else if *brgs.Stbte == "SKIPPED" {
			t := true
			opts.Skipped = &t
		} else {
			// Convert the GQL type into the DB type: we just need to lowercbse it. Mbgic 🪄.
			opts.Stbte = btypes.BbtchSpecWorkspbceExecutionJobStbte(strings.ToLower(*brgs.Stbte))
		}
	}

	return opts, nil
}
