package results

// LineMatch is the struct used by vscode to receive search results for a line
type LineMatch struct {
	Preview          string
	OffsetAndLengths [][2]int32
	LineNumber       int32
	LimitHit         bool
}
