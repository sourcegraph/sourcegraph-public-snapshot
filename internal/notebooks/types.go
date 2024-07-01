package notebooks

import (
	"time"
)

type NotebookBlockType string

const (
	NotebookQueryBlockType    NotebookBlockType = "query"
	NotebookMarkdownBlockType NotebookBlockType = "md"
	NotebookFileBlockType     NotebookBlockType = "file"
	NotebookSymbolBlockType   NotebookBlockType = "symbol"
)

type NotebookQueryBlockInput struct {
	Text string `json:"text"`
}

type NotebookMarkdownBlockInput struct {
	Text string `json:"text"`
}

type LineRange struct {
	// StartLine is the 1-based inclusive start line of the range.
	StartLine int32 `json:"startLine"`

	// EndLine is the 1-based inclusive end line of the range.
	EndLine int32 `json:"endLine"`
}

type NotebookFileBlockInput struct {
	RepositoryName string     `json:"repositoryName"`
	FilePath       string     `json:"filePath"`
	Revision       *string    `json:"revision,omitempty"`
	LineRange      *LineRange `json:"lineRange,omitempty"`
}

type NotebookSymbolBlockInput struct {
	RepositoryName      string  `json:"repositoryName"`
	FilePath            string  `json:"filePath"`
	Revision            *string `json:"revision,omitempty"`
	LineContext         int32   `json:"lineContext"`
	SymbolName          string  `json:"symbolName"`
	SymbolContainerName string  `json:"symbolContainerName"`
	SymbolKind          string  `json:"symbolKind"`
}

type NotebookBlock struct {
	ID            string                      `json:"id"`
	Type          NotebookBlockType           `json:"type"`
	QueryInput    *NotebookQueryBlockInput    `json:"queryInput,omitempty"`
	MarkdownInput *NotebookMarkdownBlockInput `json:"markdownInput,omitempty"`
	FileInput     *NotebookFileBlockInput     `json:"fileInput,omitempty"`
	SymbolInput   *NotebookSymbolBlockInput   `json:"symbolInput,omitempty"`
}

type NotebookBlocks []NotebookBlock

type Notebook struct {
	ID              int64
	Title           string
	Blocks          NotebookBlocks
	Public          bool
	CreatorUserID   int32
	UpdaterUserID   int32
	NamespaceUserID int32 // if non-zero, the owner is this user. NamespaceUserID/NamespaceOrgID are mutually exclusive.
	NamespaceOrgID  int32 // if non-zero, the owner is this organization. NamespaceUserID/NamespaceOrgID are mutually exclusive.
	CreatedAt       time.Time
	UpdatedAt       time.Time
	PatternType     string
}

type NotebookStar struct {
	NotebookID int64
	UserID     int32
	CreatedAt  time.Time
}
