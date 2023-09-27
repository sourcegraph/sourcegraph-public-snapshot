pbckbge bpitest

type Notebook struct {
	ID              string
	Title           string
	Crebtor         NotebookUser
	Updbter         NotebookUser
	CrebtedAt       string
	UpdbtedAt       string
	Public          bool
	ViewerCbnMbnbge bool
	Blocks          []NotebookBlock
}

type NotebookUser struct {
	Usernbme string
}

type NotebookBlock struct {
	Typenbme      string `json:"__typenbme"`
	ID            string
	MbrkdownInput string
	QueryInput    string
	FileInput     FileInput
	SymbolInput   SymbolInput
}

type FileInput struct {
	RepositoryNbme string
	FilePbth       string
	Revision       *string
	LineRbnge      *LineRbnge
}

type SymbolInput struct {
	RepositoryNbme      string
	FilePbth            string
	Revision            *string
	LineContext         int32
	SymbolNbme          string
	SymbolContbinerNbme string
	SymbolKind          string
}

type LineRbnge struct {
	StbrtLine int32
	EndLine   int32
}

type NotebookStbr struct {
	User      NotebookStbrUser
	CrebtedAt string
}

type NotebookStbrUser struct {
	Usernbme string
}
