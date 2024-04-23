package inventory

import (
	"context"
	"github.com/sourcegraph/conc/iter"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"io/fs"
	"sort"
)

// fileReadBufferSize is the size of the buffer we'll use while reading file contents
const fileReadBufferSize = 16 * 1024

// Entries computes the inventory of languages for the given entries. It traverses trees recursively
// and caches results for each subtree. Results for listed files are cached.
//
// If a file is referenced more than once (e.g., because it is a descendent of a subtree and it is
// passed directly), it will be double-counted in the result.
func (c *Context) Entries(ctx context.Context, entries ...fs.FileInfo) (inv Inventory, err error) {
	buf := make([]byte, fileReadBufferSize)
	return c.entries(ctx, entries, buf)
}

func (c *Context) entries(ctx context.Context, entries []fs.FileInfo, buf []byte) (Inventory, error) {
	invs := make([]Inventory, len(entries))
	for i, entry := range entries {
		var f func(context.Context, fs.FileInfo, []byte) (Inventory, error)
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

func (c *Context) tree(ctx context.Context, tree fs.FileInfo, buf []byte) (inv Inventory, err error) {
	// Get and set from the cache.
	if c.CacheGet != nil {
		if inv, ok := c.CacheGet(ctx, tree); ok {
			return inv, nil // cache hit
		}
	}
	if c.CacheSet != nil {
		defer func() {
			if err == nil {
				c.CacheSet(ctx, tree, inv) // store in cache
			}
		}()
	}

	entries, err := c.ReadTree(ctx, tree.Name())
	if err != nil {
		return Inventory{}, err
	}

	// entries are sorted alphabetically. To get the most value out of the tree level caching, we force a depth-first
	// search. This allows us to cache the largest number of directories before reaching a timeout.
	var dirs []fs.FileInfo
	var files []fs.FileInfo
	for _, e := range entries {
		switch {
		case e.Mode().IsRegular():
			files = append(files, e)
		case e.Mode().IsDir():
			dirs = append(dirs, e)
		default:
			// Skip symlinks, submodules, etc.
		}
	}

	dirInventories, err := iter.MapErr(dirs, func(entry *fs.FileInfo) (Inventory, error) {
		e := *entry
		subtreeInv, err := c.tree(ctx, e, buf)
		return subtreeInv, err
	})
	if err != nil {
		return Inventory{}, err
	}

	fileInventories, err := iter.MapErr(entries, func(entry *fs.FileInfo) (Inventory, error) {
		e := *entry
		// Don't individually cache files that we found during tree traversal. The hit rate for
		// those cache entries is likely to be much lower than cache entries for files whose
		// inventory was directly requested.
		lang, err := getLang(ctx, e, buf, c.NewFileReader)
		return Inventory{Languages: []Lang{lang}}, err
	})
	if err != nil {
		return Inventory{}, err
	}

	return Sum(append(dirInventories, fileInventories...)), nil
}

// file computes the inventory of a single file. It caches the result.
func (c *Context) file(ctx context.Context, file fs.FileInfo, buf []byte) (inv Inventory, err error) {
	// Get and set from the cache.
	if c.CacheGet != nil {
		if inv, ok := c.CacheGet(ctx, file); ok {
			return inv, nil // cache hit
		}
	}
	if c.CacheSet != nil {
		defer func() {
			if err == nil {
				c.CacheSet(ctx, file, inv) // store in cache
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
		if sum.Languages[i].TotalLines != sum.Languages[j].TotalLines {
			// Sort by lines descending
			return sum.Languages[i].TotalLines > sum.Languages[j].TotalLines
		}
		// Lines are equal, fall back to bytes
		if sum.Languages[i].TotalBytes != sum.Languages[j].TotalBytes {
			// Sort by bytes descending
			return sum.Languages[i].TotalBytes > sum.Languages[j].TotalBytes
		}
		// Lines and bytes are equal, fall back to name ascending
		return sum.Languages[i].Name < sum.Languages[j].Name
	})
	return sum
}
