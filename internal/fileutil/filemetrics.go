package fileutil

import (
	"bufio"
	"bytes"
	"context"
	"io"
	"strings"

	"github.com/go-enry/go-enry/v2"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type FileMetrics struct {
	Languages []string
	Lines     uint64
	Words     uint64
	Bytes     uint64
}

const fileReadBufferSize = 16 * 1024

var newLine = []byte{'\n'}

// Return the first laguage in the list.
// If the language list is empty, return an empty string.
func (f *FileMetrics) FirstLanguage() (language string) {
	language = ""
	if len(f.Languages) > 0 {
		language = f.Languages[0]
	}
	return language
}

// Guesses the language based on the file name and extension.
// Somewhat unreliable, given the long history of programming languages
// and the limited number of "nice" extensions. But it's fast.
func (f *FileMetrics) LanguagesByFileNameAndExtension(filePath string) {
	f.Languages = enry.GetLanguagesByFilename(filePath, nil, nil)
	if len(f.Languages) == 0 {
		f.Languages = enry.GetLanguagesByExtension(filePath, nil, nil)
	}
}

// Guesses the language based on the file content.
// Generally more reliable than `LanguagesByFileNameAndExtension`.
// But can still be amiguous, so still returns multiple possible languages.
func (f *FileMetrics) LanguagesByFileContent(filePath string, fileContent []byte) {
	f.Languages = enry.GetLanguagesByContent(filePath, fileContent, nil)
}

// Calculate the metrics for a file.
// Grab the languages based on the file name/extension first
// and if the file reader function gets a valid file reader,
// use the file contents to refine the language detection,
// and count the lines and bytes.
func (f *FileMetrics) CalculateFileMetrics(ctx context.Context, filePath string, fileReaderFunc func(ctx context.Context, path string) (io.ReadCloser, error)) (err error) {
	// start with what can be calculated without involving the file contents
	f.LanguagesByFileNameAndExtension(filePath)

	var fileContents []byte

	fileContents, f.Lines, f.Words, f.Bytes, err = scan(ctx, filePath, fileReaderFunc)
	if err != nil {
		return errors.Wrapf(err, "scanning %s", filePath)
	}

	if len(f.Languages) != 1 {
		// indeterminate language detection
		// involve the file contents
		// the initial chunk of file contents should be enough to detect a language
		f.Languages = enry.GetLanguagesByContent(filePath, fileContents, nil)
	}
	return nil
}

// Read the file contents, counting lines, words and bytes.
// Return the first `fileReadBufferSize` bytes along with the counts
// so that the caller can use the contents to refine language detection
func scan(ctx context.Context, filePath string, fileReaderFunc func(ctx context.Context, path string) (io.ReadCloser, error)) (beginningOfFile []byte, lineCount, wordCount, byteCount uint64, err error) {

	rc, err := fileReaderFunc(ctx, filePath)
	if err != nil {
		err = errors.Wrap(err, "getting file reader")
		return
	}
	if rc == nil {
		// no file contents to scan
		err = nil
		return
	}
	defer rc.Close()

	buf := make([]byte, fileReadBufferSize)
	var copied int

	scanner := bufio.NewScanner(rc)
	scanner.Split(scanNewLines)
	for scanner.Scan() {
		line := scanner.Bytes()
		lineCount++
		wordCount += uint64(len(strings.Fields(string(line))))
		byteCount += uint64(len(line) + 1)
		if lineLen := len(line); copied+lineLen+1 <= fileReadBufferSize {
			copy(buf[copied:], line)
			copied += lineLen
			copy(buf[copied:], newLine)
			copied++
		}
	}
	beginningOfFile = buf[:copied]

	return
}

// clone ScanLines, but without the `dropCR` so that byte size will be reliable
func scanNewLines(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	if i := bytes.IndexByte(data, '\n'); i >= 0 {
		// We have a full newline-terminated line.
		return i + 1, data, nil
	}
	// If we're at EOF, we have a final, non-terminated line. Return it.
	if atEOF {
		return len(data), data, nil
	}
	// Request more data.
	return 0, nil, nil
}
