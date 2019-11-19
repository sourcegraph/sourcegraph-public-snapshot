// Package inventory scans a directory tree to identify the
// programming languages, etc., in use.
package inventory

import (
	"bytes"
	"context"
	"io"
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

var newLine = []byte{'\n'}

func getLang(ctx context.Context, file os.FileInfo, buf []byte, rc io.ReadCloser) (*Lang, error) {
	if rc != nil {
		defer rc.Close()
	}
	if !file.Mode().IsRegular() || enry.IsVendor(file.Name()) {
		return nil, nil
	}
	lang := Lang{
		TotalBytes: uint64(file.Size()),
	}

	// In many cases, GetLanguageByFilename can detect the language conclusively just from the
	// filename. If not, we pass a subset of the file contents for analysis.
	matchedLang, safe := GetLanguageByFilename(file.Name())

	// No content
	if rc == nil || lang.TotalBytes == 0 {
		lang.Name = matchedLang
		return &lang, nil
	}

	if !safe {
		// Detect language from content
		n, err := io.ReadFull(rc, buf)
		if err != nil && err != io.ErrUnexpectedEOF {
			return nil, errors.Wrap(err, "reading initial file data")
		}
		matchedLang = enry.GetLanguage(file.Name(), buf[:n])
		lang.TotalLines += uint64(bytes.Count(buf[:n], newLine))
		lang.Name = matchedLang
		// File is smaller than buf, we can exit early
		if err == io.ErrUnexpectedEOF {
			if !bytes.HasSuffix(buf[:n], newLine) {
				// Add final line
				lang.TotalLines++
			}
			return &lang, nil
		}
	}
	lang.Name = matchedLang

	// Count lines
	var trailingNewLine bool
	for {
		n, err := rc.Read(buf)
		lang.TotalLines += uint64(bytes.Count(buf[:n], newLine))
		if n > 0 {
			trailingNewLine = bytes.HasSuffix(buf[:n], newLine)
		}
		if err == io.EOF {
			if !trailingNewLine {
				// Add final line
				lang.TotalLines++
			}
			break
		}
		if err != nil {
			return nil, errors.Wrap(err, "reading lines")
		}
	}
	return &lang, nil
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
