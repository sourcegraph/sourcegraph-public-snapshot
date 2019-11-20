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

	// NewFileReader is called to get an io.ReadCloser from the file at path.
	NewFileReader func(ctx context.Context, path string) (io.ReadCloser, error)

	// CacheGet, if set, returns the cached inventory and true for the given tree, or false for a cache miss.
	CacheGet func(os.FileInfo) (Inventory, bool)

	// CacheSet, if set, stores the inventory in the cache for the given tree.
	CacheSet func(os.FileInfo, Inventory)
}

// fileReadBufferSize is the size of the buffer we'll use while reading file contents
const fileReadBufferSize = 16 * 1024

// Tree computes the inventory of languages for a tree. It caches the inventories of subtrees.
func (c *Context) Tree(ctx context.Context, tree os.FileInfo) (inv Inventory, err error) {
	buf := make([]byte, fileReadBufferSize)
	return c.tree(ctx, tree, buf)
}

func (c *Context) tree(ctx context.Context, tree os.FileInfo, buf []byte) (inv Inventory, err error) {
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
	langStats := map[string]*Lang{} // language name -> stats
	for _, e := range entries {
		switch {
		case e.Mode().IsRegular(): // file
			lang, err := getLang(ctx, e, buf, c.NewFileReader)
			if err != nil {
				return Inventory{}, errors.Wrapf(err, "inventory file %q", e.Name())
			}
			if lang.Name != "" {
				l := langStats[lang.Name]
				if l == nil {
					l = &Lang{
						Name: lang.Name,
					}
				}
				l.TotalBytes += lang.TotalBytes
				l.TotalLines += lang.TotalLines
				langStats[lang.Name] = l
			}

		case e.Mode().IsDir(): // subtree
			entryInv, err := c.tree(ctx, e, buf)
			if err != nil {
				return Inventory{}, errors.Wrapf(err, "inventory tree %q", e.Name())
			}
			for _, lang := range entryInv.Languages {
				l := langStats[lang.Name]
				if l == nil {
					l = &Lang{
						Name: lang.Name,
					}
				}
				l.TotalBytes += lang.TotalBytes
				l.TotalLines += lang.TotalLines
				langStats[lang.Name] = l
			}

		default:
			// Skip symlinks, submodules, etc.
		}
	}
	return sum(langStats), nil
}

func sum(langStats map[string]*Lang) Inventory {
	sum := Inventory{Languages: make([]Lang, 0, len(langStats))}
	for name := range langStats {
		stats := langStats[name]
		stats.Name = name
		sum.Languages = append(sum.Languages, *stats)
	}
	sort.Slice(sum.Languages, func(i, j int) bool {
		return sum.Languages[i].TotalBytes > sum.Languages[j].TotalBytes || (sum.Languages[i].TotalBytes == sum.Languages[j].TotalBytes && sum.Languages[i].Name < sum.Languages[j].Name)
	})
	return sum
}
