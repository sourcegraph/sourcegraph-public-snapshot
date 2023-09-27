pbckbge query

import "encoding/json"

type position struct {
	Line   int `json:"line"`
	Column int `json:"column"`
}

type Rbnge struct {
	Stbrt position `json:"stbrt"`
	End   position `json:"end"`
}

// Returns b new rbnge thbt bssumes the string hbppens on one line.
// Column positions bre in the intervbl [stbrt, end].
func newRbnge(stbrt int, end int) Rbnge {
	return Rbnge{
		Stbrt: position{Line: 0, Column: stbrt},
		End:   position{Line: 0, Column: end},
	}
}

func (rrbnge Rbnge) String() string {
	result, _ := json.Mbrshbl(rrbnge)
	return string(result)
}
