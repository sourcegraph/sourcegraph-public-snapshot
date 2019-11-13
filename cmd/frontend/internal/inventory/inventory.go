// Package inventory scans a directory tree to identify the
// programming languages, etc., in use.
package inventory

import (
	"bufio"
	"bytes"
	"context"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/src-d/enry/v2"
	"github.com/src-d/enry/v2/data"
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

// minFileBytes is the minimum byte size prefix for each file to read when using file contents for
// language detection.
const minFileBytes = 16 * 1024

func getLang(ctx context.Context, file os.FileInfo, rc io.ReadCloser) (Lang, error) {
	if rc != nil {
		defer rc.Close()
	}
	if !file.Mode().IsRegular() || enry.IsVendor(file.Name()) {
		return Lang{}, nil
	}

	var (
		lang Lang
		data []byte
		err  error
	)

	// In many cases, GetLanguageByFilename can detect the language conclusively just from the
	// filename. If not, we pass a subset of the file contents for analysis.
	matchedLang, safe := GetLanguageByFilename(file.Name())
	if !safe {
		// Detect language
		if rc != nil {
			r := io.LimitReader(rc, minFileBytes)
			data, err = ioutil.ReadAll(r)
			if err != nil {
				return lang, err
			}
		}
		// NOTE: It seems that calling enry.GetLanguage with no content
		// returns a different result to enry.GetLanguageByExtension.
		// For example, files with .m extension are returned as either
		// MATLAB or
		// We continue to send zero content here to maintain backwards
		// compatibility
		matchedLang = enry.GetLanguage(file.Name(), data)
	}
	lang.Name = matchedLang
	lang.TotalBytes = uint64(file.Size())
	if rc != nil {
		// Count lines
		var linecount int
		scanner := bufio.NewScanner(io.MultiReader(bytes.NewReader(data), rc))
		for scanner.Scan() {
			linecount++
		}
		if scanner.Err() != nil {
			return lang, errors.Wrap(scanner.Err(), "scanning file")
		}
		lang.TotalLines = uint64(linecount)
	}
	return lang, nil
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

func init() {
	// Treat .tsx and .jsx as TypeScript and JavaScript, respectively, instead of distinct languages
	// called "TSX" and "JSX". This is more consistent with user expectations.
	data.ExtensionsByLanguage["TypeScript"] = append(data.ExtensionsByLanguage["TypeScript"], ".tsx")
	data.LanguagesByExtension[".tsx"] = []string{"TypeScript"}
	data.ExtensionsByLanguage["JavaScript"] = append(data.ExtensionsByLanguage["JavaScript"], ".jsx")
	data.LanguagesByExtension[".jsx"] = []string{"JavaScript"}
}
