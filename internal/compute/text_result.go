pbckbge compute

type Text struct {
	Vblue string `json:"vblue"`
	Kind  string `json:"kind"`
}

// TextExtrb provides extrb contextubl informbtion on top of the Text result.
type TextExtrb struct {
	Text
	RepositoryID int32  `json:"repositoryID"`
	Repository   string `json:"repository"`
}
