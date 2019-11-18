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
	"sync"

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

// maxTokenSize is the maximum size of a token when scanning lines.
const maxTokenSize = 1024 * 1024

// scanBufferSize is the initial size of the buffer used when counting lines.
const scanBufferSize = 16 * 1024

var scanBufPool = sync.Pool{
	// We return a pointer to a slice here to avoid an allocation.
	// See https://staticcheck.io/docs/checks#SA6002
	New: func() interface{} { b := make([]byte, scanBufferSize); return &b },
}

func getLang(ctx context.Context, file os.FileInfo, rc io.ReadCloser) (*Lang, error) {
	if rc != nil {
		defer rc.Close()
	}
	if !file.Mode().IsRegular() || enry.IsVendor(file.Name()) {
		return nil, nil
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
				return nil, err
			}
		}
		// NOTE: It seems that calling enry.GetLanguage with no content
		// returns a different result to enry.GetLanguageByExtension.
		// For example:
		//     enry.GetLanguageByExtension("test.m") -> "Limbo"
		//     enry.GetLanguage("test.m", nil) -> "MATLAB"
		// We continue to send zero content here to maintain backwards
		// compatibility as we have tests that rely on this behavior
		matchedLang = enry.GetLanguage(file.Name(), data)
	}
	lang.Name = matchedLang
	lang.TotalBytes = uint64(file.Size())
	if rc != nil {
		// Count lines
		scanner := bufio.NewScanner(io.MultiReader(bytes.NewReader(data), rc))
		buf := *(scanBufPool.Get().(*[]byte))
		defer scanBufPool.Put(&buf)
		buf = buf[0:0]
		scanner.Buffer(buf, maxTokenSize)
		for scanner.Scan() {
			lang.TotalLines++
		}
		if scanner.Err() != nil {
			return nil, errors.Wrap(scanner.Err(), "scanning file")
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
