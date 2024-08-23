package common

import (
	"encoding/json"

	"github.com/bevzzz/nb/schema"
)

type Notebook struct {
	VersionMajor int             `json:"nbformat"`
	VersionMinor int             `json:"nbformat_minor"`
	Metadata     json.RawMessage `json:"metadata"` // TODO: omitempty
}

func (n *Notebook) Version() schema.Version {
	return schema.Version{
		Major: n.VersionMajor,
		Minor: n.VersionMinor,
	}
}

const (
	PlainText    = "text/plain"
	MarkdownText = "text/markdown"
	Stdout       = "application/vnd.jupyter.stdout" // Custom mime-type for stream output to stdout.
	Stderr       = "application/vnd.jupyter.stderr" // Custom mime-type for stream output to stderr.
)

// Markdown defines the schema for a "markdown" cell.
type Markdown struct {
	Source MultilineString `json:"source"`
}

var _ schema.Cell = (*Markdown)(nil)

func (md *Markdown) Type() schema.CellType {
	return schema.Markdown
}

func (md *Markdown) MimeType() string {
	return MarkdownText
}

func (md *Markdown) Text() []byte {
	return md.Source.Text()
}

// Raw defines the schema for a "raw" cell.
type Raw struct {
	Source   MultilineString `json:"source"`
	Metadata RawCellMetadata `json:"metadata"`
}

var _ schema.Cell = (*Raw)(nil)

func (raw *Raw) Type() schema.CellType {
	return schema.Raw
}

func (raw *Raw) MimeType() string {
	return raw.Metadata.MimeType()
}

func (raw *Raw) Text() []byte {
	return raw.Source.Text()
}

// RawCellMetadata may specify a target conversion format.
type RawCellMetadata struct {
	Format      *string `json:"format"`
	RawMimeType *string `json:"raw_mimetype"`
}

// MimeType returns a more specific mime-type if one is provided and "text/plain" otherwise.
func (raw *RawCellMetadata) MimeType() string {
	switch {
	case raw.Format != nil:
		return *raw.Format
	case raw.RawMimeType != nil:
		return *raw.RawMimeType
	default:
		return PlainText
	}
}
