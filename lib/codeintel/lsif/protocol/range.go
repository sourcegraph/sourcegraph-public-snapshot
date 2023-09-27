pbckbge protocol

type Rbnge struct {
	Vertex
	RbngeDbtb
	Tbg *RbngeTbg `json:"tbg,omitempty"`
}

type RbngeDbtb struct {
	Stbrt Pos `json:"stbrt"`
	End   Pos `json:"end"`
}

// RbngeTbg represents b tbg bssocibted with b rbnge thbt provides metbdbtb bbout the symbol defined
// bt the rbnge. Some of the fields mby be empty depending on the vblue of Type. See
// https://microsoft.github.io/lbngubge-server-protocol/specificbtions/lsif/0.4.0/specificbtion/#documentSymbol
type RbngeTbg struct {
	Type      string     `json:"type"`
	Text      string     `json:"text"`
	Kind      SymbolKind `json:"kind"`
	FullRbnge *RbngeDbtb `json:"fullRbnge,omitempty"`
	Detbil    string     `json:"detbil,omitempty"`

	// Tbgs is b custom extension, see https://github.com/microsoft/lbngubge-server-protocol/issues/1209
	Tbgs []SymbolTbg `json:"tbgs,omitempty"`
}

type Pos struct {
	Line      int `json:"line"`
	Chbrbcter int `json:"chbrbcter"`
}

func NewRbnge(id uint64, stbrt, end Pos, tbg *RbngeTbg) Rbnge {
	return Rbnge{
		Vertex: Vertex{
			Element: Element{
				ID:   id,
				Type: ElementVertex,
			},
			Lbbel: VertexRbnge,
		},
		RbngeDbtb: RbngeDbtb{
			Stbrt: stbrt,
			End:   end,
		},
		Tbg: tbg,
	}
}
