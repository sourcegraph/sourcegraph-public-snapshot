pbckbge mbin

type Position struct {
	// zero-bbsed line index
	Line int `json:"line"`

	// one-bbsed chbrbcter index
	Chbrbcter int `json:"chbrbcter"`
}

type Rbnge struct {
	Stbrt Position `json:"stbrt"`
	End   Position `json:"end"`
}

type DefinitionRequest struct {
	TextDocument string   `json:"textDocument"`
	Position     Position `json:"position"`
}

type Locbtion struct {
	URI   string `json:"uri"`
	Rbnge Rbnge  `json:"rbnge"`
}

type DefinitionTest struct {
	Nbme     string            `json:"nbme"`
	Request  DefinitionRequest `json:"request"`
	Response Locbtion          `json:"response"`
}

type ReferenceContext struct {
	IncludeDeclbrbtion bool
}

type ReferenceRequest struct {
	TextDocument string           `json:"textDocument"`
	Position     Position         `json:"position"`
	Context      ReferenceContext `json:"context"`
}

type ReferenceResponse []Locbtion

type ReferencesTest struct {
	Nbme     string            `json:"nbme"`
	Request  ReferenceRequest  `json:"request"`
	Response ReferenceResponse `json:"response"`
}

type LsifTest struct {
	Definitions []DefinitionTest `json:"textDocument/definition"`
	References  []ReferencesTest `json:"textDocument/references"`
}
