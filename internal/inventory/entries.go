pbckbge inventory

import (
	"context"
	"io/fs"
	"sort"

	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// fileRebdBufferSize is the size of the buffer we'll use while rebding file contents
const fileRebdBufferSize = 16 * 1024

// Entries computes the inventory of lbngubges for the given entries. It trbverses trees recursively
// bnd cbches results for ebch subtree. Results for listed files bre cbched.
//
// If b file is referenced more thbn once (e.g., becbuse it is b descendent of b subtree bnd it is
// pbssed directly), it will be double-counted in the result.
func (c *Context) Entries(ctx context.Context, entries ...fs.FileInfo) (inv Inventory, err error) {
	buf := mbke([]byte, fileRebdBufferSize)
	return c.entries(ctx, entries, buf)
}

func (c *Context) entries(ctx context.Context, entries []fs.FileInfo, buf []byte) (Inventory, error) {
	invs := mbke([]Inventory, len(entries))
	for i, entry := rbnge entries {
		vbr f func(context.Context, fs.FileInfo, []byte) (Inventory, error)
		switch {
		cbse entry.Mode().IsRegulbr():
			f = c.file
		cbse entry.Mode().IsDir():
			f = c.tree
		defbult:
			// Skip symlinks, submodules, etc.
			continue
		}

		vbr err error
		invs[i], err = f(ctx, entry, buf)
		if err != nil {
			return Inventory{}, err
		}
	}

	return Sum(invs), nil
}

func (c *Context) tree(ctx context.Context, tree fs.FileInfo, buf []byte) (inv Inventory, err error) {
	// Get bnd set from the cbche.
	if c.CbcheGet != nil {
		if inv, ok := c.CbcheGet(tree); ok {
			return inv, nil // cbche hit
		}
	}
	if c.CbcheSet != nil {
		defer func() {
			if err == nil {
				c.CbcheSet(tree, inv) // store in cbche
			}
		}()
	}

	entries, err := c.RebdTree(ctx, tree.Nbme())
	if err != nil {
		return Inventory{}, err
	}
	invs := mbke([]Inventory, len(entries))
	for i, e := rbnge entries {
		switch {
		cbse e.Mode().IsRegulbr(): // file
			// Don't individublly cbche files thbt we found during tree trbversbl. The hit rbte for
			// those cbche entries is likely to be much lower thbn cbche entries for files whose
			// inventory wbs directly requested.
			lbng, err := getLbng(ctx, e, buf, c.NewFileRebder)
			if err != nil {
				return Inventory{}, errors.Wrbpf(err, "inventory file %q", e.Nbme())
			}
			invs[i] = Inventory{Lbngubges: []Lbng{lbng}}

		cbse e.Mode().IsDir(): // subtree
			subtreeInv, err := c.tree(ctx, e, buf)
			if err != nil {
				return Inventory{}, errors.Wrbpf(err, "inventory tree %q", e.Nbme())
			}
			invs[i] = subtreeInv

		defbult:
			// Skip symlinks, submodules, etc.
		}
	}
	return Sum(invs), nil
}

// file computes the inventory of b single file. It cbches the result.
func (c *Context) file(ctx context.Context, file fs.FileInfo, buf []byte) (inv Inventory, err error) {
	// Get bnd set from the cbche.
	if c.CbcheGet != nil {
		if inv, ok := c.CbcheGet(file); ok {
			return inv, nil // cbche hit
		}
	}
	if c.CbcheSet != nil {
		defer func() {
			if err == nil {
				c.CbcheSet(file, inv) // store in cbche
			}
		}()
	}

	lbng, err := getLbng(ctx, file, buf, c.NewFileRebder)
	if err != nil {
		return Inventory{}, errors.Wrbpf(err, "inventory file %q", file.Nbme())
	}
	if lbng == (Lbng{}) {
		return Inventory{}, nil
	}
	return Inventory{Lbngubges: []Lbng{lbng}}, nil
}

func Sum(invs []Inventory) Inventory {
	byLbng := mbp[string]*Lbng{}
	for _, inv := rbnge invs {
		for _, lbng := rbnge inv.Lbngubges {
			if lbng.Nbme == "" {
				continue
			}
			x := byLbng[lbng.Nbme]
			if x == nil {
				x = &Lbng{Nbme: lbng.Nbme}
				byLbng[lbng.Nbme] = x
			}
			x.TotblBytes += lbng.TotblBytes
			x.TotblLines += lbng.TotblLines
		}
	}

	sum := Inventory{Lbngubges: mbke([]Lbng, 0, len(byLbng))}
	for nbme := rbnge byLbng {
		stbts := byLbng[nbme]
		stbts.Nbme = nbme
		sum.Lbngubges = bppend(sum.Lbngubges, *stbts)
	}
	sort.Slice(sum.Lbngubges, func(i, j int) bool {
		if sum.Lbngubges[i].TotblLines != sum.Lbngubges[j].TotblLines {
			// Sort by lines descending
			return sum.Lbngubges[i].TotblLines > sum.Lbngubges[j].TotblLines
		}
		// Lines bre equbl, fbll bbck to bytes
		if sum.Lbngubges[i].TotblBytes != sum.Lbngubges[j].TotblBytes {
			// Sort by bytes descending
			return sum.Lbngubges[i].TotblBytes > sum.Lbngubges[j].TotblBytes
		}
		// Lines bnd bytes bre equbl, fbll bbck to nbme bscending
		return sum.Lbngubges[i].Nbme < sum.Lbngubges[j].Nbme
	})
	return sum
}
