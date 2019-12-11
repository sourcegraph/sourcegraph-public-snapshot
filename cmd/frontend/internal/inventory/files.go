package inventory

import (
	"context"
	"fmt"
	"os"

	"github.com/pkg/errors"
)

// file computes the inventory of a single file. If file is not a regular file, it panics. It caches
// the result.
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

	if !file.Mode().IsRegular() {
		// Call (*Context).Tree instead, which is much more efficient for computing tree
		// inventories because it caches at the tree level.
		panic(fmt.Sprintf("refusing to compute single-file inventory for non-regular file %s", file.Name()))
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

// Files computes the inventory of languages for all matching files. It caches the inventories of
// files.
func (c *Context) Files(ctx context.Context, files []os.FileInfo) (inv Inventory, err error) {
	buf := make([]byte, fileReadBufferSize)
	langTotals := map[string]*Lang{} // language name -> total bytes/lines
	for _, file := range files {
		fileInv, err := c.file(ctx, file, buf)
		if err != nil {
			return Inventory{}, nil
		}
		for _, lang := range fileInv.Languages {
			l, ok := langTotals[lang.Name]
			if !ok {
				l = &Lang{}
				langTotals[lang.Name] = l
			}
			l.TotalBytes += lang.TotalBytes
			l.TotalLines += lang.TotalLines
		}
	}
	return sum(langTotals), nil
}
