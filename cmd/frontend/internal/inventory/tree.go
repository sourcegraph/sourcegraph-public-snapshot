package inventory

import (
	"context"
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

	// ReadFile is called to read the partial contents of the file at path. At least the specified
	// number of bytes must be returned (or the entire file, if it is smaller).
	ReadFile func(ctx context.Context, path string, minBytes int64) ([]byte, error)

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
	langTotalBytes := map[string]uint64{} // language name -> total bytes
	for _, e := range entries {
		switch {
		case e.Mode().IsRegular(): // file
			lang, err := detect(ctx, e, c.ReadFile)
			if err != nil {
				return Inventory{}, errors.Wrapf(err, "inventory file %q", e.Name())
			}
			if lang != "" {
				langTotalBytes[lang] += uint64(e.Size())
			}

		case e.Mode().IsDir(): // subtree
			entryInv, err := c.Tree(ctx, e)
			if err != nil {
				return Inventory{}, errors.Wrapf(err, "inventory tree %q", e.Name())
			}
			for _, lang := range entryInv.Languages {
				langTotalBytes[lang.Name] += lang.TotalBytes
			}

		default:
			// Skip symlinks, submodules, etc.
		}
	}
	return sum(langTotalBytes), nil
}

func sum(langTotalBytes map[string]uint64) Inventory {
	sum := Inventory{Languages: make([]Lang, 0, len(langTotalBytes))}
	for lang, totalBytes := range langTotalBytes {
		sum.Languages = append(sum.Languages, Lang{Name: lang, TotalBytes: totalBytes})
	}
	sort.Slice(sum.Languages, func(i, j int) bool {
		return sum.Languages[i].TotalBytes > sum.Languages[j].TotalBytes || (sum.Languages[i].TotalBytes == sum.Languages[j].TotalBytes && sum.Languages[i].Name < sum.Languages[j].Name)
	})
	return sum
}
