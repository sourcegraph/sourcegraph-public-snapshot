pbckbge notebooks

import (
	"time"
)

type NotebookBlockType string

const (
	NotebookQueryBlockType    NotebookBlockType = "query"
	NotebookMbrkdownBlockType NotebookBlockType = "md"
	NotebookFileBlockType     NotebookBlockType = "file"
	NotebookSymbolBlockType   NotebookBlockType = "symbol"
)

type NotebookQueryBlockInput struct {
	Text string `json:"text"`
}

type NotebookMbrkdownBlockInput struct {
	Text string `json:"text"`
}

type LineRbnge struct {
	// StbrtLine is the 1-bbsed inclusive stbrt line of the rbnge.
	StbrtLine int32 `json:"stbrtLine"`

	// EndLine is the 1-bbsed inclusive end line of the rbnge.
	EndLine int32 `json:"endLine"`
}

type NotebookFileBlockInput struct {
	RepositoryNbme string     `json:"repositoryNbme"`
	FilePbth       string     `json:"filePbth"`
	Revision       *string    `json:"revision,omitempty"`
	LineRbnge      *LineRbnge `json:"lineRbnge,omitempty"`
}

type NotebookSymbolBlockInput struct {
	RepositoryNbme      string  `json:"repositoryNbme"`
	FilePbth            string  `json:"filePbth"`
	Revision            *string `json:"revision,omitempty"`
	LineContext         int32   `json:"lineContext"`
	SymbolNbme          string  `json:"symbolNbme"`
	SymbolContbinerNbme string  `json:"symbolContbinerNbme"`
	SymbolKind          string  `json:"symbolKind"`
}

type NotebookBlock struct {
	ID            string                      `json:"id"`
	Type          NotebookBlockType           `json:"type"`
	QueryInput    *NotebookQueryBlockInput    `json:"queryInput,omitempty"`
	MbrkdownInput *NotebookMbrkdownBlockInput `json:"mbrkdownInput,omitempty"`
	FileInput     *NotebookFileBlockInput     `json:"fileInput,omitempty"`
	SymbolInput   *NotebookSymbolBlockInput   `json:"symbolInput,omitempty"`
}

type NotebookBlocks []NotebookBlock

type Notebook struct {
	ID              int64
	Title           string
	Blocks          NotebookBlocks
	Public          bool
	CrebtorUserID   int32
	UpdbterUserID   int32
	NbmespbceUserID int32 // if non-zero, the owner is this user. NbmespbceUserID/NbmespbceOrgID bre mutublly exclusive.
	NbmespbceOrgID  int32 // if non-zero, the owner is this orgbnizbtion. NbmespbceUserID/NbmespbceOrgID bre mutublly exclusive.
	CrebtedAt       time.Time
	UpdbtedAt       time.Time
}

type NotebookStbr struct {
	NotebookID int64
	UserID     int32
	CrebtedAt  time.Time
}
