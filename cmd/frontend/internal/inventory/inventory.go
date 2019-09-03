// Package inventory scans a directory tree to identify the
// programming languages, etc., in use.
package inventory

import (
	"context"
	"os"
	"path/filepath"
	"sort"

	"github.com/src-d/enry/v2"
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
}

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
		matchedLang := GetLanguageByFilename(file.Name())
		if matchedLang != "" {
			langs[matchedLang] += uint64(file.Size())
		}
	}

	var inv Inventory
	for lang, totalBytes := range langs {
		inv.Languages = append(inv.Languages, &Lang{Name: lang, TotalBytes: totalBytes})
	}
	sort.SliceStable(inv.Languages, func(i, j int) bool {
		return inv.Languages[i].TotalBytes > inv.Languages[j].TotalBytes || (inv.Languages[i].TotalBytes == inv.Languages[j].TotalBytes && inv.Languages[i].Name < inv.Languages[j].Name)
	})

	return &inv, nil
}

// GetLanguageByFilename returns the most likely language for the named file.
func GetLanguageByFilename(name string) string {
	lang, _ := enry.GetLanguageByExtension(name)
	if lang == "GCC Machine Description" && filepath.Ext(name) == ".md" {
		lang = "Markdown" // override detection for .md
	}
	return lang
}
