package api

import (
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
)

type Input struct {
	Filepath string
	Text     string
	Lines    []string
	Bytes    []byte
}

func NewInput(filename string, textBytes []byte) *Input {
	text := string(textBytes)
	return &Input{
		Filepath: filename,
		Text:     text,
		Lines:    strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n"),
		Bytes:    textBytes,
	}
}

func (i *Input) Format(n *sitter.Node) string {
	if n == nil {
		return "<nil>"
	}
	lineIndex := int(n.StartPoint().Row)
	if lineIndex < 0 {
		return "lineIndex<0"
	}
	if lineIndex >= len(i.Lines) {
		return "lineIndex>len(lines)"
	}
	line := i.Lines[lineIndex]
	character := int(n.StartPoint().Column)
	if character < 0 {
		return "lineIndex<0"
	}
	if character >= len(line) {
		return "lineIndex>len(lines)"
	}
	length := int(n.EndPoint().Column) - character
	if n.StartPoint().Row != n.EndPoint().Row {
		length = len(line) - character
	}
	return strings.ReplaceAll(line, "\t", " ") + "\n" + strings.Repeat(" ", character) + strings.Repeat("^", length)
}

func (i *Input) Substring(n *sitter.Node) string {
	start := int(n.StartByte())
	if start < 0 {
		start = 0
	}
	end := int(n.EndByte())
	if end > len(i.Bytes) {
		end = len(i.Bytes)
	}
	return string(i.Bytes[start:end])
}

func (i *Input) Uri() string {
	return "file:///" + i.Filepath
}
