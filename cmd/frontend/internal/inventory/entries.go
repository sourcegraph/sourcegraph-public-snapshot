package inventory

import (
	"context"
	"os"
	"sort"

	"github.com/pkg/errors"
)

// fileReadBufferSize is the size of the buffer we'll use while reading file contents
const fileReadBufferSize = 16 * 1024

// Entries computes the inventory of languages for the given entries. It traverses trees recursively
// and caches results for each subtree. Results for listed files are cached.
//
// Known issue: If a listed entry's (os.FileInfo).Size() == 0, it is treated as empty. Callers must
// ensure that either the size is known. If the size is not known and the caller only cares about
// line counts, the caller must ensure that a nonzero size is reported and must ignore all
// TotalBytes counts.
//
// If a file is referenced more than once (e.g., because it is a descendent of a subtree and it is
// passed directly), it will be double-counted in the result.
func (c *Context) Entries(ctx context.Context, entries ...os.FileInfo) (inv Inventory, err error) {
	buf := make([]byte, fileReadBufferSize)
	return c.entries(ctx, entries, buf)
}

func (c *Context) entries(ctx context.Context, entries []os.FileInfo, buf []byte) (Inventory, error) {
	invs := make([]Inventory, len(entries))
	for i, entry := range entries {
		var f func(context.Context, os.FileInfo, []byte) (Inventory, error)
		switch {
		case entry.Mode().IsRegular():
			f = c.file
		case entry.Mode().IsDir():
			f = c.tree
		default:
			// Skip symlinks, submodules, etc.
			continue
		}

		var err error
		invs[i], err = f(ctx, entry, buf)
		if err != nil {
			return Inventory{}, err
		}
	}

	return Sum(invs), nil
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
	invs := make([]Inventory, len(entries))
	for i, e := range entries {
		switch {
		case e.Mode().IsRegular(): // file
			// Don't individually cache files that we found during tree traversal. The hit rate for
			// those cache entries is likely to be much lower than cache entries for files whose
			// inventory was directly requested.
			lang, err := getLang(ctx, e, buf, c.NewFileReader)
			if err != nil {
				return Inventory{}, errors.Wrapf(err, "inventory file %q", e.Name())
			}
			invs[i] = Inventory{Languages: []Lang{lang}}

		case e.Mode().IsDir(): // subtree
			subtreeInv, err := c.tree(ctx, e, buf)
			if err != nil {
				return Inventory{}, errors.Wrapf(err, "inventory tree %q", e.Name())
			}
			invs[i] = subtreeInv

		default:
			// Skip symlinks, submodules, etc.
		}
	}
	return Sum(invs), nil
}

// file computes the inventory of a single file. It caches the result.
func (c *Context) file(ctx context.Context, file os.FileInfo, buf []byte) (inv Inventory, err error) {
	// Get and set from the cache.
	if c.CacheGet != nil {
		if inv, ok := c.CacheGet(file); ok {
			return inv, nil // cache hit
		}
	}
	if c.CacheSet != nil {
		defer func() {
			if err == nil {
				c.CacheSet(file, inv) // store in cache
			}
		}()
	}

	lang, err := getLang(ctx, file, buf, c.NewFileReader)
	if err != nil {
		return Inventory{}, errors.Wrapf(err, "inventory file %q", file.Name())
	}
	if lang == (Lang{}) {
		return Inventory{}, nil
	}
	return Inventory{Languages: []Lang{lang}}, nil
}

func Sum(invs []Inventory) Inventory {
	byLang := map[string]*Lang{}
	for _, inv := range invs {
		for _, lang := range inv.Languages {
			if lang.Name == "" {
				continue
			}
			x := byLang[lang.Name]
			if x == nil {
				x = &Lang{Name: lang.Name}
				byLang[lang.Name] = x
			}
			x.TotalBytes += lang.TotalBytes
			x.TotalLines += lang.TotalLines
		}
	}

	sum := Inventory{Languages: make([]Lang, 0, len(byLang))}
	for name := range byLang {
		stats := byLang[name]
		stats.Name = name
		sum.Languages = append(sum.Languages, *stats)
	}
	sort.Slice(sum.Languages, func(i, j int) bool {
		return sum.Languages[i].TotalLines > sum.Languages[j].TotalLines || (sum.Languages[i].TotalLines == sum.Languages[j].TotalLines && sum.Languages[i].Name < sum.Languages[j].Name)
	})
	return sum
}
