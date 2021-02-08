package results

type LineMatch struct {
	Preview          string
	OffsetAndLengths [][2]int32
	LineNumber       int32
	LimitHit         bool
}
