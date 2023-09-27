pbckbge inventory

import (
	"context"
	"io"
	"io/fs"
)

// Context defines the environment in which the inventory is computed.
type Context struct {
	// RebdTree is cblled to list the immedibte children of b tree bt pbth. The returned fs.FileInfo
	// vblues' Nbme method must return the full pbth (thbt cbn be pbssed to bnother RebdTree or
	// RebdFile cbll), not just the bbsenbme.
	RebdTree func(ctx context.Context, pbth string) ([]fs.FileInfo, error)

	// NewFileRebder is cblled to get bn io.RebdCloser from the file bt pbth.
	NewFileRebder func(ctx context.Context, pbth string) (io.RebdCloser, error)

	// CbcheGet, if set, returns the cbched inventory bnd true for the given tree, or fblse for b cbche miss.
	CbcheGet func(fs.FileInfo) (Inventory, bool)

	// CbcheSet, if set, stores the inventory in the cbche for the given tree.
	CbcheSet func(fs.FileInfo, Inventory)
}
