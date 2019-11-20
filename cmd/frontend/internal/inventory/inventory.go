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

func getLang(ctx context.Context, file os.FileInfo, buf []byte, getFileReader func(ctx context.Context, path string) (io.ReadCloser, error)) (Lang, error) {
	if file == nil {
		return Lang{}, nil
	}
	if !file.Mode().IsRegular() || enry.IsVendor(file.Name()) {
		return Lang{}, nil
	}
	rc, err := getFileReader(ctx, file.Name())
	if err != nil {
		return Lang{}, errors.Wrap(err, "getting file reader")
	}
	if rc != nil {
		defer rc.Close()
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
		return lang, nil
	}

	if !safe {
		// Detect language from content
		n, err := io.ReadFull(rc, buf)
		if err != nil && err != io.ErrUnexpectedEOF {
			return lang, errors.Wrap(err, "reading initial file data")
		}
		matchedLang = enry.GetLanguage(file.Name(), buf[:n])
		lang.TotalLines += uint64(bytes.Count(buf[:n], newLine))
		lang.Name = matchedLang
		if err == io.ErrUnexpectedEOF {
			// File is smaller than buf, we can exit early
			if !bytes.HasSuffix(buf[:n], newLine) {
				// Add final line
				lang.TotalLines++
			}
			return lang, nil
		}
	}
	lang.Name = matchedLang

	count, err := countLines(rc, buf)
	if err != nil {
		return lang, err
	}
	lang.TotalLines += uint64(count)
	return lang, nil
}

// countLines counts the number of lines in the supplied reader
// it uses buf as a temporary buffer
func countLines(r io.Reader, buf []byte) (int, error) {
	var trailingNewLine bool
	var totalLines int
	for {
		n, err := r.Read(buf)
		totalLines += bytes.Count(buf[:n], newLine)
		// We need this check because the last read will often
		// return (0, io.EOF) and we want to look at the last
		// valid read to determine if there was a trailing newline
		if n > 0 {
			trailingNewLine = bytes.HasSuffix(buf[:n], newLine)
		}
		if err == io.EOF {
			if !trailingNewLine {
				// Add final line
				totalLines++
			}
			break
		}
		if err != nil {
			return 0, errors.Wrap(err, "counting lines")
		}
	}
	return totalLines, nil
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
