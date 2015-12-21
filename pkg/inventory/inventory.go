// Package inventory scans a directory tree to identify the
// programming languages, etc., in use.
package inventory

import (
	"path"
	"sort"
	"strings"
	"time"

	"github.com/kr/fs"
	"golang.org/x/net/context"
)

// TODO(sqs): Expand detection capabilities by using
// github.com/petermattis/linguist,
// https://godoc.org/github.com/sevki/goeylinguine or similar.
var extLangs = map[string]string{
	".go":    "Go",
	".java":  "Java",
	".py":    "Python",
	".rb":    "Ruby",
	".scala": "Scala",
	".js":    "JavaScript",
	".c":     "C",
}

var skipExts = []string{
	".min.js",
	".min.css",
}

func skipExt(filename string) bool {
	for _, ext := range skipExts {
		if strings.HasSuffix(filename, ext) {
			return true
		}
	}
	return false
}

var skipDirs = map[string]struct{}{
	"node_modules":     struct{}{},
	"bower_components": struct{}{},
	"vendor":           struct{}{},
	"Godeps":           struct{}{},
	"third_party":      struct{}{},
}

// Scan performs an inventory of the tree at fs.
//
// Scan attempts to respect the ctx's deadline. If it is nearing the
// deadline, it will return a partial inventory and the error value
// context.Canceled.
func Scan(ctx context.Context, vfs fs.FileSystem) (*Inventory, error) {
	langs := map[string]uint64{}

	// Respect deadline.
	//
	// TODO(sqs): Also support ctx cancelation.
	deadline, hasDeadline := ctx.Deadline()
	const finishTime = 15 * time.Millisecond
	var err error

	w := fs.WalkFS("/", vfs)
	for w.Step() {
		if hasDeadline && deadline.Sub(time.Now()) < finishTime {
			err = context.Canceled
			break
		}

		if err := w.Err(); err != nil {
			return nil, err
		}

		fi := w.Stat()
		name := fi.Name()
		mode := fi.Mode()
		switch {
		case mode.IsRegular():
			if skipExt(name) {
				continue
			}

			if lang := extLangs[path.Ext(name)]; lang != "" {
				langs[lang] += uint64(fi.Size())
			}

		case mode.IsDir():
			if _, skipDir := skipDirs[name]; skipDir {
				w.SkipDir()
			}
		}
	}

	var inv Inventory
	for lang, totalBytes := range langs {
		inv.Languages = append(inv.Languages, &Lang{Name: lang, TotalBytes: totalBytes})
	}
	sort.Sort(sort.Reverse(langsByTotalBytes(inv.Languages)))
	return &inv, err
}

type langsByTotalBytes []*Lang

func (v langsByTotalBytes) Len() int           { return len(v) }
func (v langsByTotalBytes) Less(i, j int) bool { return v[i].TotalBytes < v[j].TotalBytes }
func (v langsByTotalBytes) Swap(i, j int)      { v[i], v[j] = v[j], v[i] }
