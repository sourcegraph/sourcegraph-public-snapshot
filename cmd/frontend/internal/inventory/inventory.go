// Package inventory scans a directory tree to identify the
// programming languages, etc., in use.
package inventory

import (
	"context"
	"os"
	"sort"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/inventory/filelang"
)

// Inventory summarizes a tree's contents (e.g., which programming
// languages are used).
type Inventory struct {
	// Languages are the programming languages used in the tree.
	Languages []*Lang `json:"Languages,omitempty"`
}

// Lang represents a programming language used in a directory tree.
type Lang struct {
	// Name is the name of a programming language (e.g., "Go" or
	// "Java").
	Name string `json:"Name,omitempty"`
	// TotalBytes is the total number of bytes of code written in the
	// programming language.
	TotalBytes uint64 `json:"TotalBytes,omitempty"`
	// Type is either "data", "programming", "markup", "prose", or
	// empty.
	Type string `json:"Type,omitempty"`
}

var byFilename = filelang.Langs.CompileByFilename()

// Get performs an inventory of the files passed in.
func Get(ctx context.Context, files []os.FileInfo) (*Inventory, error) {
	langs := map[string]uint64{}

	for _, file := range files {
		// NOTE: We used to skip vendored files, but the
		// filelang.IsVendored function is slow (benchmark goes from
		// 160ms to 0.5ms without the check). Currently Inventory is
		// just used to determine which languages are in a repo, the
		// relative usage (TotalBytes) is not exposed or used. So
		// including vendored files should be fine for the aggregate
		// stats.
		matchedLangs := byFilename(file.Name())
		if len(matchedLangs) > 0 {
			langs[matchedLangs[0].Name] += uint64(file.Size())
		}
	}

	var inv Inventory
	for lang, totalBytes := range langs {
		inv.Languages = append(inv.Languages, &Lang{Name: lang, TotalBytes: totalBytes})
	}
	sort.Sort(sort.Reverse(langsByTotalBytes(inv.Languages)))

	// Set Type field.
	for _, il := range inv.Languages {
		for _, l := range filelang.Langs {
			if il.Name == l.Name {
				il.Type = l.Type
				break
			}
		}
	}

	return &inv, nil
}

type langsByTotalBytes []*Lang

func (v langsByTotalBytes) Len() int      { return len(v) }
func (v langsByTotalBytes) Swap(i, j int) { v[i], v[j] = v[j], v[i] }
func (v langsByTotalBytes) Less(i, j int) bool {
	if v[i].TotalBytes == v[j].TotalBytes {
		return v[i].Name < v[j].Name
	}
	return v[i].TotalBytes < v[j].TotalBytes
}
