package types

// UploadLocation is a path and range pair from within a particular upload. The target commit
// denotes the target commit for which the location was set (the originally requested commit).
type UploadLocation struct {
	Dump         Dump
	Path         string
	TargetCommit string
	TargetRange  Range
}

type Range struct {
	Start Position
	End   Position
}

// Position is a unique position within a file.
type Position struct {
	Line      int
	Character int
}
