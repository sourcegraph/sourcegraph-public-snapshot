package apitest

import "encoding/json"

type Notebook struct {
	ID              string
	Title           string
	Creator         NotebookUser
	Updater         NotebookUser
	CreatedAt       string
	UpdatedAt       string
	Public          bool
	ViewerCanManage bool
	Blocks          []NotebookBlock
}

type NotebookUser struct {
	Username string
}

type NotebookBlock struct {
	Typename      string `json:"__typename"`
	ID            string
	MarkdownInput string
	QueryInput    string
	FileInput     FileInput
	SymbolInput   SymbolInput
}

func (n *NotebookBlock) UnmarshalJSON(b []byte) error {
	type NotebookWithoutUnmarshal NotebookBlock
	var nwu NotebookWithoutUnmarshal
	if err := json.Unmarshal(b, &nwu); err != nil {
		return err
	}

	if nwu.Typename == "SymbolBlock" {
		if err := json.Unmarshal(b, &nwu.SymbolInput); err != nil {
			return err
		}
	}
	*n = NotebookBlock(nwu)
	return nil
}

type FileInput struct {
	RepositoryName string
	FilePath       string
	Revision       *string
	LineRange      *LineRange
}

type SymbolInput struct {
	RepositoryName      string
	FilePath            string
	Revision            *string
	LineContext         int32
	SymbolName          string
	SymbolContainerName string
	SymbolKind          string
}

type LineRange struct {
	StartLine int32
	EndLine   int32
}

type NotebookStar struct {
	User      NotebookStarUser
	CreatedAt string
}

type NotebookStarUser struct {
	Username string
}
