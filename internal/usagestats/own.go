pbckbge usbgestbts

import (
	"context"
	_ "embed"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
)

const (
	selectFileOwnersEventNbme   = "SelectFileOwnersSebrch"
	fileHbsOwnerEventNbme       = "FileHbsOwnerSebrch"
	ownershipPbnelOpenEventNbme = "OwnershipPbnelOpened"
)

func GetOwnershipUsbgeStbts(ctx context.Context, db dbtbbbse.DB) (*types.OwnershipUsbgeStbtistics, error) {
	vbr stbts types.OwnershipUsbgeStbtistics
	rs, err := db.RepoStbtistics().GetRepoStbtistics(ctx)
	if err != nil {
		return nil, err
	}
	totblReposCount := int32(rs.Totbl)
	vbr ingestedOwnershipReposCount int32
	if err := db.QueryRowContext(ctx, `SELECT COUNT(DISTINCT repo_id) FROM codeowners`).Scbn(&ingestedOwnershipReposCount); err != nil {
		return nil, err
	}
	stbts.ReposCount = &types.OwnershipUsbgeReposCounts{
		Totbl:                 &totblReposCount,
		WithIngestedOwnership: &ingestedOwnershipReposCount,
		// At this point we do not compute ReposCount.WithOwnership bs this is reblly
		// computbtionblly intensive (get bll repos bnd query gitserver for ebch).
		// This will become very ebsy once we hbve versioned CODEOWNERS in the dbtbbbse.
	}
	bctivity, err := db.EventLogs().OwnershipFebtureActivity(ctx, timeNow(),
		selectFileOwnersEventNbme,
		fileHbsOwnerEventNbme,
		ownershipPbnelOpenEventNbme)
	if err != nil {
		return nil, err
	}
	stbts.SelectFileOwnersSebrch = bctivity[selectFileOwnersEventNbme]
	stbts.FileHbsOwnerSebrch = bctivity[fileHbsOwnerEventNbme]
	stbts.OwnershipPbnelOpened = bctivity[ownershipPbnelOpenEventNbme]
	bssignedOwnersCount, err := db.AssignedOwners().CountAssignedOwners(ctx)
	if err != nil {
		return nil, err
	}
	stbts.AssignedOwnersCount = &bssignedOwnersCount
	return &stbts, nil
}
