package main

type QueryResponse struct {
	Data struct {
		Repository struct {
			Commit struct {
				Blob struct {
					LSIF struct {
						Definitions     Definitions     `json:"definitions"`
						References      References      `json:"references"`
						Implementations Implementations `json:"implementations"`
						Prototypes      Prototypes      `json:"prototypes"`
					} `json:"lsif"`
				} `json:"blob"`
			} `json:"commit"`
		} `json:"repository"`
	} `json:"data"`
}

type Definitions struct {
	Nodes []Node `json:"nodes"`
}

type References struct {
	Nodes    []Node   `json:"nodes"`
	PageInfo PageInfo `json:"pageInfo"`
}

type Implementations struct {
	Nodes    []Node   `json:"nodes"`
	PageInfo PageInfo `json:"pageInfo"`
}

type Prototypes struct {
	Nodes    []Node   `json:"nodes"`
	PageInfo PageInfo `json:"pageInfo"`
}

type Node struct {
	Resource `json:"resource"`
	Range    `json:"range"`
}

type Resource struct {
	Path       string     `json:"path"`
	Repository Repository `json:"repository"`
	Commit     Commit     `json:"commit"`
}

type Repository struct {
	Name string `json:"name"`
}

type Commit struct {
	Oid string `json:"oid"`
}

type Range struct {
	Start Position `json:"start"`
	End   Position `json:"end"`
}

type Position struct {
	Line      int `json:"line"`
	Character int `json:"character"`
}

type PageInfo struct {
	EndCursor string `json:"endCursor"`
}
