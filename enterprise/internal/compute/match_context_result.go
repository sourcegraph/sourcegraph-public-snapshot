package compute

// Location represents the position in a text file, which may be an absolute
// offset or line/column pair. Offsets can be converted to line/columns or vice
// versa when the input file is available. We represent the possibility, but not
// the requirement, of representing either offset or line/column in this data
// type because tools or processes may expose only, e.g., offsets for
// performance reasons (e.g., parsing) and leave conversion (which has
// performance implications) up to the client. Nevertheless, from a usability
// perspective, it is advantageous to represent both possibilities in a single
// type. Conventionally, "null" values may be represented with -1.
type Location struct {
	Offset int `json:"offset"`
	Line   int `json:"line"`
	Column int `json:"column"`
}

type Range struct {
	Start Location `json:"start"`
	End   Location `json:"end"`
}

type Data struct {
	Value string `json:"value"`
	Range Range  `json:"range"`
}

type Environment map[string]Data

type Match struct {
	Value       string      `json:"value"`
	Range       Range       `json:"range"`
	Environment Environment `json:"environment"`
}

type MatchContext struct {
	Matches      []Match `json:"matches"`
	Path         string  `json:"path"`
	RepositoryID int32   `json:"repositoryID"`
	Repository   string  `json:"repository"`
}

func newLocation(line, column, offset int) Location {
	return Location{
		Offset: offset,
		Line:   line,
		Column: column,
	}
}

func newRange(startOffset, endOffset int) Range {
	return Range{
		Start: newLocation(-1, -1, startOffset),
		End:   newLocation(-1, -1, endOffset),
	}
}
