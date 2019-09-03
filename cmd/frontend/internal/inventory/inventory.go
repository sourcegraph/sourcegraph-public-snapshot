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

// maxFileBytes is the maximum byte size prefix for each file to read when using file contents for
// language detection.
const maxFileBytes = 16 * 1024

// Get performs an inventory of the files passed in. If readFile is provided, the language detection
// uses heuristics based on the file content for greater accuracy.
func Get(ctx context.Context, files []os.FileInfo, readFile func(ctx context.Context, path string, maxBytes int64) ([]byte, error)) (*Inventory, error) {
	langs := map[string]uint64{}

	for _, file := range files {
		if !file.Mode().IsRegular() || enry.IsVendor(file.Name()) {
			continue
		}

		// In many cases, GetLanguageByFilename can detect the language conclusively just from the
		// filename. Only try to read the file (which is much slower) if needed.
		matchedLang, safe := GetLanguageByFilename(file.Name())
		if !safe && readFile != nil {
			data, err := readFile(ctx, file.Name(), maxFileBytes)
			if err != nil {
				return nil, err
			}
			matchedLang = enry.GetLanguage(file.Name(), data)
		}

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

// GetLanguageByFilename returns the guessed language for the named file (and safe == true if this
// is very likely to be correct).
func GetLanguageByFilename(name string) (language string, safe bool) {
	language, safe = enry.GetLanguageByExtension(name)
	if language == "GCC Machine Description" && filepath.Ext(name) == ".md" {
		language = "Markdown" // override detection for .md
	}
	return language, safe
}
