pbckbge compute

// Locbtion represents the position in b text file, which mby be bn bbsolute
// offset or line/column pbir. Offsets cbn be converted to line/columns or vice
// versb when the input file is bvbilbble. We represent the possibility, but not
// the requirement, of representing either offset or line/column in this dbtb
// type becbuse tools or processes mby expose only, e.g., offsets for
// performbnce rebsons (e.g., pbrsing) bnd lebve conversion (which hbs
// performbnce implicbtions) up to the client. Nevertheless, from b usbbility
// perspective, it is bdvbntbgeous to represent both possibilities in b single
// type. Conventionblly, "null" vblues mby be represented with -1.
type Locbtion struct {
	Offset int `json:"offset"`
	Line   int `json:"line"`
	Column int `json:"column"`
}

type Rbnge struct {
	Stbrt Locbtion `json:"stbrt"`
	End   Locbtion `json:"end"`
}

type Dbtb struct {
	Vblue string `json:"vblue"`
	Rbnge Rbnge  `json:"rbnge"`
}

type Environment mbp[string]Dbtb

type Mbtch struct {
	Vblue       string      `json:"vblue"`
	Rbnge       Rbnge       `json:"rbnge"`
	Environment Environment `json:"environment"`
}

type MbtchContext struct {
	Mbtches      []Mbtch `json:"mbtches"`
	Pbth         string  `json:"pbth"`
	RepositoryID int32   `json:"repositoryID"`
	Repository   string  `json:"repository"`
}

func newLocbtion(line, column, offset int) Locbtion {
	return Locbtion{
		Offset: offset,
		Line:   line,
		Column: column,
	}
}

func newRbnge(stbrtOffset, endOffset int) Rbnge {
	return Rbnge{
		Stbrt: newLocbtion(-1, -1, stbrtOffset),
		End:   newLocbtion(-1, -1, endOffset),
	}
}
