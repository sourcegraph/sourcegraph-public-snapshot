package inventory

import (
	"context"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"golang.org/x/sync/semaphore"
	"io/fs"
	"os"
	"sort"
	"sync"
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
	// todo bahrmichael: explain reasoning for picked value, and make it configurable
	sem := semaphore.NewWeighted(5)
	return c.entries(ctx, entries, buf, sem)
}

func (c *Context) entries(ctx context.Context, entries []fs.FileInfo, buf []byte, sem *semaphore.Weighted) (Inventory, error) {
	invs := make([]Inventory, len(entries))
	for i, entry := range entries {
		var f func(context.Context, fs.FileInfo, []byte, *semaphore.Weighted) (Inventory, error)
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
		invs[i], err = f(ctx, entry, buf, sem)
		if err != nil {
			return Inventory{}, err
		}
	}

	return Sum(invs), nil
}

type treeIteratorResult struct {
	index     int
	inventory Inventory
	err       error
}

func (c *Context) tree(ctx context.Context, tree fs.FileInfo, buf []byte, sem *semaphore.Weighted) (inv Inventory, err error) {
	// Get and set from the cache.
	if err := sem.Acquire(ctx, 1); err != nil {
		return Inventory{}, err
	}
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
	sem.Release(1)
	if err != nil {
		return Inventory{}, err
	}
	invs := make([]Inventory, len(entries))
	results := make(chan treeIteratorResult, len(entries)) // Buffer the channel to the number of entries
	var wg sync.WaitGroup

	for i, e := range entries {
		wg.Add(1)
		go func(i int, e os.FileInfo) {
			defer wg.Done()

			switch {
			case e.Mode().IsRegular(): // file
				// Don't individually cache files that we found during tree traversal. The hit rate for
				// those cache entries is likely to be much lower than cache entries for files whose
				// inventory was directly requested.
				if err := sem.Acquire(ctx, 1); err != nil {
					results <- treeIteratorResult{i, Inventory{}, err}
				}
				lang, err := getLang(ctx, e, buf, c.NewFileReader)
				sem.Release(1)
				if err != nil {
					results <- treeIteratorResult{i, Inventory{Languages: []Lang{lang}}, err}
				}
				results <- treeIteratorResult{i, Inventory{Languages: []Lang{lang}}, nil}

			case e.Mode().IsDir(): // subtree
				subtreeInv, err := c.tree(ctx, e, buf, sem)
				if err != nil {
					results <- treeIteratorResult{i, subtreeInv, err}
				}
				results <- treeIteratorResult{i, subtreeInv, nil}

			default:
				// Skip symlinks, submodules, etc.
			}
		}(i, e)
	}

	// Wait for all goroutines to finish
	wg.Wait()
	close(results)

	for res := range results {
		if res.err != nil {
			return Inventory{}, res.err
		}
		invs[res.index] = res.inventory
	}

	return Sum(invs), nil
}

// file computes the inventory of a single file. It caches the result.
func (c *Context) file(ctx context.Context, file fs.FileInfo, buf []byte, sem *semaphore.Weighted) (inv Inventory, err error) {
	// Get and set from the cache.
	if err := sem.Acquire(ctx, 1); err != nil {
		return Inventory{}, err
	}
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
	sem.Release(1)
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
