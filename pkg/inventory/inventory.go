// Package inventory scans a directory tree to identify the
// programming languages, etc., in use.
package inventory

import (
	"context"
	"os"
	"sort"
	"strings"

	"github.com/sourcegraph/sourcegraph/pkg/inventory/filelang"
)

// Inventory summarizes a tree's contents (e.g., which programming
// languages are used).
type Inventory struct {
	// Languages are the programming languages used in the tree.
	Languages []*Lang `json:"Languages,omitempty"`
}

// Constants that can be values in the Inventory.Languages slice.
const (
	LangGo   = "Go"
	LangJava = "Java"
)

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

// ConfigName matches the `langservers[].language` field in the site
// configuration.
func (l *Lang) ConfigName() string {
	switch l.Name {
	case "Shell":
		return "bash"
	case "C++":
		return "cpp"
	case "C#":
		return "cs"
	default:
		return strings.ToLower(l.Name)
	}
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

// PrimaryProgrammingLanguage returns the primary programming language
// discovered in the inventory (the language with the most
// non-vendored/non-skipped bytes of code). If there is none, the
// empty string is returned.
func (inv *Inventory) PrimaryProgrammingLanguage() string {
	langs := ProgrammingLangsOnly(inv.Languages)
	if len(langs) == 0 {
		return ""
	}
	return langs[0].Name
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

// LangsOfType returns only langs of the given type (matching the Type
// field).
func LangsOfType(langs []*Lang, typ string) []*Lang {
	var langs2 []*Lang
	for _, l := range langs {
		if l.Type == typ {
			langs2 = append(langs2, l)
		}
	}
	return langs2
}

// ProgrammingLangsOnly returns the subset of langs whose Type is
// "programming" (e.g., not "prose", "markup", etc.).
func ProgrammingLangsOnly(langs []*Lang) []*Lang {
	return LangsOfType(langs, "programming")
}
