// Package inventory scans a directory tree to identify the
// programming languages, etc., in use.
package inventory

import (
	"context"
	"io"
	"io/fs"
	"log"
	"math"

	"github.com/go-enry/go-enry/v2"
	"github.com/go-enry/go-enry/v2/data"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/fileutil"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// Inventory summarizes a tree's contents (e.g., which programming
// languages are used).
type Inventory struct {
	// Languages are the programming languages used in the tree.
	Languages []Lang `json:"Languages,omitempty"`
}

// Lang represents a programming language used in a directory tree.
type Lang struct {
	// Name is the name of a programming language (e.g., "Go" or
	// "Java").
	Name string `json:"Name,omitempty"`
	// TotalBytes is the total number of bytes of code written in the
	// programming language.
	TotalBytes uint64 `json:"TotalBytes,omitempty"`
	// TotalLines is the total number of lines of code written in the
	// programming language.
	TotalLines uint64 `json:"TotalLines,omitempty"`
}

func getLang(ctx context.Context, db database.DB, repoID api.RepoID, file fs.FileInfo, commitID api.CommitID, getFileReader func(ctx context.Context, path string) (io.ReadCloser, error)) (Lang, error) {
	if file == nil {
		return Lang{}, nil
	}
	if !file.Mode().IsRegular() || enry.IsVendor(file.Name()) {
		return Lang{}, nil
	}

	// Check the cache first.
	// It is a two-level cache: redis backed by the database.
	metrics := db.FileMetrics().GetFileMetrics(ctx, repoID, commitID, file.Name())

	var err error

	if metrics == nil {
		// Nothing cached.
		// Calculate the metrics and cache them.
		var rc io.ReadCloser
		rc, err = getFileReader(ctx, file.Name())
		if err != nil {
			return Lang{}, errors.Wrap(err, "getting file reader")
		}
		if rc != nil {
			defer rc.Close()
		}

		metrics = &fileutil.FileMetrics{}

		// this might be slow if `getFileReader` is set up to call `gitserver`
		err = metrics.CalculateFileMetrics(file.Name(), rc)

		// don't make the client wait for the cache insert
		bgCtx, cancel := context.WithCancel(ctx)
		go func() {
			defer cancel()
			db.FileMetrics().SetFileMetrics(bgCtx, repoID, commitID, file.Name(), metrics, err == nil)
		}()

		// No way to read the file contents
		// Historically, in this situation, getLang returned the first language, lexographically,
		// determined from the file name/extension, and the file size as reported by the file info
		if rc == nil {
			return Lang{
				Name:       metrics.FirstLanguage(),
				TotalBytes: uint64(file.Size()),
			}, nil
		}
	}

	if len(metrics.Languages) != 1 && metrics.Bytes == 0 {
		// multiple language options and an empty file, so no way to refine the language
		// historically, in this situation, getLang returned nothing
		return Lang{}, nil
	}

	return Lang{
		Name:       metrics.FirstLanguage(),
		TotalBytes: uint64(math.Max(float64(file.Size()), float64(metrics.Bytes))),
		TotalLines: metrics.Lines,
	}, err
}

// GetLanguageByFilename returns the guessed language for the named file (and
// safe == true if this is very likely to be correct).
func GetLanguageByFilename(name string) (language string, safe bool) {
	language, safe = enry.GetLanguageByFilename(name)
	if language != "" {
		return language, safe
	}
	return enry.GetLanguageByExtension(name)
}

func GetLanguageByContent(name string, content []byte) (language string, safe bool) {
	return enry.GetLanguageByContent(name, content)
}

func GetLanguage(name string, content []byte) (language string) {
	return enry.GetLanguage(name, content)
}

func init() {
	// Treat .tsx and .jsx as TypeScript and JavaScript, respectively, instead of distinct languages
	// called "TSX" and "JSX". This is more consistent with user expectations.
	data.ExtensionsByLanguage["TypeScript"] = append(data.ExtensionsByLanguage["TypeScript"], ".tsx")
	data.LanguagesByExtension[".tsx"] = []string{"TypeScript"}
	data.ExtensionsByLanguage["JavaScript"] = append(data.ExtensionsByLanguage["JavaScript"], ".jsx")
	data.LanguagesByExtension[".jsx"] = []string{"JavaScript"}

	// Prefer more popular languages which share extensions
	preferLanguage("Markdown", ".md") // instead of GCC Machine Description
	preferLanguage("Rust", ".rs")     // instead of RenderScript
}

// preferLanguage updates LanguagesByExtension to have lang listed first for
// ext.
func preferLanguage(lang, ext string) {
	langs := data.LanguagesByExtension[ext]
	for i := range langs {
		if langs[i] == lang {
			// swap to front
			for ; i > 0; i-- {
				langs[i-1], langs[i] = langs[i], langs[i-1]
			}
			return
		}
	}
	log.Fatalf("%q not in %q: %q", lang, ext, langs)
}
