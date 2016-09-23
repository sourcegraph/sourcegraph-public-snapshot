// Package inventory scans a directory tree to identify the
// programming languages, etc., in use.
package inventory

import (
	"os"
	"sort"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/inventory/filelang"

	"context"

	"github.com/kr/fs"
)

// Scan performs an inventory of the tree at fs.
//
// Scan respects the ctx's deadline. If it exceeds the deadline,
// it will return a partial inventory and the error value
// context.DeadlineExceeded.
func Scan(ctx context.Context, vfs fs.FileSystem) (*Inventory, error) {
	langs := map[string]uint64{}
	var err error

	w := fs.WalkFS("/", vfs)
Outer:
	for w.Step() {
		select {
		case <-ctx.Done():
			err = ctx.Err()
			// Carry through this error to the final "return"
			// statement, so that we return a partial result.
			break Outer
		default:
		}

		if err := w.Err(); err != nil {
			if w.Path() != "/" && (os.IsNotExist(err) || os.IsPermission(err)) {
				continue
			}
			return nil, err
		}

		fi := w.Stat()
		if filelang.IsVendored(w.Path(), w.Stat().Mode().IsDir()) {
			w.SkipDir()
			continue
		}
		if fi.Mode().IsRegular() {
			matchedLangs := filelang.Langs.ByFilename(fi.Name())
			if len(matchedLangs) > 0 {
				langs[matchedLangs[0].Name] += uint64(fi.Size())
			}
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

	return &inv, err
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

func (v langsByTotalBytes) Len() int           { return len(v) }
func (v langsByTotalBytes) Less(i, j int) bool { return v[i].TotalBytes < v[j].TotalBytes }
func (v langsByTotalBytes) Swap(i, j int)      { v[i], v[j] = v[j], v[i] }

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
