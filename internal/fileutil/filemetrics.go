package fileutil

import (
	"bufio"
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

	if len(f.Languages) != 1 && len(fileContents) > 0 {
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

	collector := newScanLinesPlusByteCounter(fileReadBufferSize)

	scanner := bufio.NewScanner(rc)
	scanner.Split(collector.ScanLines)
	for scanner.Scan() {
		line := scanner.Bytes()
		lineCount++
		wordCount += uint64(len(strings.Fields(string(line))))
	}
	byteCount = collector.byteCount
	beginningOfFile = collector.buffer[:collector.bufferSize]

	return
}

func newScanLinesPlusByteCounter(bufferSize int) *ScanLinesPlusByteCounter {
	return &ScanLinesPlusByteCounter{buffer: make([]byte, bufferSize)}
}

// create a data structure to hold the byte size of the file (really, the stream)
// along with the reading of the lines
type ScanLinesPlusByteCounter struct {
	byteCount  uint64
	buffer     []byte
	bufferSize int
}

// Entrypoint that gathers byte count of what's read so far
// and collects `x.bufferSize` number of bytes from the beginning of the stream.
// Requires `x.buffer` to be initialized in order to collect bytes.
func (x *ScanLinesPlusByteCounter) ScanLines(data []byte, atEOF bool) (advance int, token []byte, err error) {
	a, t, e := bufio.ScanLines(data, atEOF)
	x.byteCount += uint64(a)
	if a > 0 && x.bufferSize+a <= cap(x.buffer) {
		copy(x.buffer[x.bufferSize:], t)
		x.bufferSize += a
	}
	return a, t, e
}
