package main

type Position struct {
	Line      int `json:"line"`
	Character int `json:"character"`
}

type Range struct {
	Start Position `json:"start"`
	End   Position `json:"end"`
}

type DefinitionRequest struct {
	TextDocument string   `json:"textDocument"`
	Position     Position `json:"position"`
}

type DefinitionResponse struct {
	TextDocument string `json:"textDocument"`
	Range        Range  `json:"range"`
}

type DefinitionTest struct {
	Name     string             `json:"name"`
	Request  DefinitionRequest  `json:"request"`
	Response DefinitionResponse `json:"response"`
}

type LsifTest struct {
	Definitions []DefinitionTest `json:"textDocument/definition"`
}
