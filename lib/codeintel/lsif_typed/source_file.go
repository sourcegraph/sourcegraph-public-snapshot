package lsif_typed

import (
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

type SourceFile struct {
	AbsolutePath string
	RelativePath string
	Text         string
	Lines        []string
}

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
		sourceFile, err := ReadSourceFile(path, relativePath)
		if err != nil {
			return err
		}
		result = append(result, sourceFile)
		return nil
	})
	return result, err
}

func ReadSourceFile(absolutePath, relativePath string) (*SourceFile, error) {
	text, err := os.ReadFile(absolutePath)
	if err != nil {
		return nil, err
	}
	return NewSourceFile(absolutePath, relativePath, string(text)), nil
}

func NewSourceFile(absolutePath, relativePath, code string) *SourceFile {
	return &SourceFile{
		AbsolutePath: absolutePath,
		RelativePath: relativePath,
		Text:         code,
		Lines:        strings.Split(code, "\n"),
	}
}

func (d *SourceFile) String() string {
	data, err := json.Marshal(&d)
	if err != nil {
		panic(err)
	}
	return string(data)
}

func (d *SourceFile) SlicePosition(position RangePosition) string {
	result := strings.Builder{}
	for line := position.Start.Line; line < position.End.Line; line++ {
		start := position.Start.Character
		if line > position.Start.Line {
			result.WriteString("\n")
			start = 0
		}
		end := position.End.Character
		if line < position.End.Line {
			end = len(d.Lines[line])
		}
		result.WriteString(d.Lines[line][start:end])
	}
	return result.String()
}
