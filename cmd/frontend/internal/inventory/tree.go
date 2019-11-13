package inventory

import (
	"context"
	"io"
	"os"
	"sort"

	"github.com/pkg/errors"
)

// Context defines the environment in which the inventory is computed.
type Context struct {
	// ReadTree is called to list the immediate children of a tree at path. The returned os.FileInfo
	// values' Name method must return the full path (that can be passed to another ReadTree or
	// ReadFile call), not just the basename.
	ReadTree func(ctx context.Context, path string) ([]os.FileInfo, error)

	// GetFileReader is called to get an io.ReadCloser from the file at path.
	GetFileReader func(ctx context.Context, path string) (io.ReadCloser, error)

	// CacheGet, if set, returns the cached inventory and true for the given tree, or false for a cache miss.
	CacheGet func(os.FileInfo) (Inventory, bool)

	// CacheSet, if set, stores the inventory in the cache for the given tree.
	CacheSet func(os.FileInfo, Inventory)
}

// Tree computes the inventory of languages for a tree. It caches the inventories of subtrees.
func (c *Context) Tree(ctx context.Context, tree os.FileInfo) (inv Inventory, err error) {
	// Get and set from the cache.
	if c.CacheGet != nil {
		if inv, ok := c.CacheGet(tree); ok {
			return inv, nil // cache hit
		}
	}
	if c.CacheSet != nil {
		defer func() {
			if err == nil {
				c.CacheSet(tree, inv) // store in cache
			}
		}()
	}

	entries, err := c.ReadTree(ctx, tree.Name())
	if err != nil {
		return Inventory{}, err
	}
	totals := map[string]Lang{} // language name -> stats
	for _, e := range entries {
		switch {
		case e.Mode().IsRegular(): // file
			rc, err := c.GetFileReader(ctx, e.Name())
			if err != nil {
				return Inventory{}, errors.Wrap(err, "getting file reader")
			}
			lang, err := getLang(ctx, e, rc)
			if err != nil {
				return Inventory{}, errors.Wrapf(err, "inventory file %q", e.Name())
			}
			if lang.Name != "" {
				l := totals[lang.Name]
				lang.TotalBytes += l.TotalBytes
				lang.TotalLines += l.TotalLines
				totals[lang.Name] = lang
			}

		case e.Mode().IsDir(): // subtree
			entryInv, err := c.Tree(ctx, e)
			if err != nil {
				return Inventory{}, errors.Wrapf(err, "inventory tree %q", e.Name())
			}
			for _, lang := range entryInv.Languages {
				l := totals[lang.Name]
				l.TotalBytes += lang.TotalBytes
				l.TotalLines += lang.TotalLines
				totals[lang.Name] = l
			}

		default:
			// Skip symlinks, submodules, etc.
		}
	}
	return sum(totals), nil
}

func sum(langStats map[string]Lang) Inventory {
	sum := Inventory{Languages: make([]Lang, 0, len(langStats))}
	for name := range langStats {
		stats := langStats[name]
		stats.Name = name
		sum.Languages = append(sum.Languages, stats)
	}
	sort.Slice(sum.Languages, func(i, j int) bool {
		return sum.Languages[i].TotalBytes > sum.Languages[j].TotalBytes || (sum.Languages[i].TotalBytes == sum.Languages[j].TotalBytes && sum.Languages[i].Name < sum.Languages[j].Name)
	})
	return sum
}
