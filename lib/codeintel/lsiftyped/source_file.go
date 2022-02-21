package lsiftyped

import (
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// SourceFile includes helper methods to deal with source files.
type SourceFile struct {
	AbsolutePath string
	RelativePath string
	Text         string
	Lines        []string
}

func NewSourceFile(absolutePath, relativePath, code string) *SourceFile {
	return &SourceFile{
		AbsolutePath: absolutePath,
		RelativePath: relativePath,
		Text:         code,
		Lines:        strings.Split(code, "\n"),
	}
}

// NewSourcesFromDirectory recursively walks the provided directory and creates a SourceFile for every regular file.
func NewSourcesFromDirectory(directory string) ([]*SourceFile, error) {
	var result []*SourceFile
	err := filepath.Walk(directory, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		relativePath, err := filepath.Rel(directory, path)
		if err != nil {
			return err
		}
		sourceFile, err := NewSourceFileFromPath(path, relativePath)
		if err != nil {
			return err
		}
		result = append(result, sourceFile)
		return nil
	})
	return result, err
}

// NewSourceFileFromPath reads the provided absolute path from disk and returns a SourceFile.
func NewSourceFileFromPath(absolutePath, relativePath string) (*SourceFile, error) {
	text, err := os.ReadFile(absolutePath)
	if err != nil {
		return nil, err
	}
	return NewSourceFile(absolutePath, relativePath, string(text)), nil
}

func (d *SourceFile) String() string {
	data, err := json.Marshal(&d)
	if err != nil {
		panic(err)
	}
	return string(data)
}

// RangeText returns the substring of the source file contents that enclose the provided range.
func (d *SourceFile) RangeText(position Range) string {
	result := strings.Builder{}
	for line := position.Start.Line; line < position.End.Line; line++ {
		start := position.Start.Character
		if line > position.Start.Line {
			result.WriteString("\n")
			start = 0
		}
		end := position.End.Character
		if line < position.End.Line {
			end = int32(len(d.Lines[line]))
		}
		result.WriteString(d.Lines[line][start:end])
	}
	return result.String()
}
