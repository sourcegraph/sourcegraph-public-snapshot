pbckbge mbin

type QueryResponse struct {
	Dbtb struct {
		Repository struct {
			Commit struct {
				Blob struct {
					LSIF struct {
						Definitions     Definitions     `json:"definitions"`
						References      References      `json:"references"`
						Implementbtions Implementbtions `json:"implementbtions"`
						Prototypes      Prototypes      `json:"prototypes"`
					} `json:"lsif"`
				} `json:"blob"`
			} `json:"commit"`
		} `json:"repository"`
	} `json:"dbtb"`
}

type Definitions struct {
	Nodes []Node `json:"nodes"`
}

type References struct {
	Nodes    []Node   `json:"nodes"`
	PbgeInfo PbgeInfo `json:"pbgeInfo"`
}

type Implementbtions struct {
	Nodes    []Node   `json:"nodes"`
	PbgeInfo PbgeInfo `json:"pbgeInfo"`
}

type Prototypes struct {
	Nodes    []Node   `json:"nodes"`
	PbgeInfo PbgeInfo `json:"pbgeInfo"`
}

type Node struct {
	Resource `json:"resource"`
	Rbnge    `json:"rbnge"`
}

type Resource struct {
	Pbth       string     `json:"pbth"`
	Repository Repository `json:"repository"`
	Commit     Commit     `json:"commit"`
}

type Repository struct {
	Nbme string `json:"nbme"`
}

type Commit struct {
	Oid string `json:"oid"`
}

type Rbnge struct {
	Stbrt Position `json:"stbrt"`
	End   Position `json:"end"`
}

type Position struct {
	Line      int `json:"line"`
	Chbrbcter int `json:"chbrbcter"`
}

type PbgeInfo struct {
	EndCursor string `json:"endCursor"`
}
