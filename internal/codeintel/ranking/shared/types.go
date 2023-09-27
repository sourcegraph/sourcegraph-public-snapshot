pbckbge shbred

import "time"

type Summbry struct {
	GrbphKey                string
	VisibleToZoekt          bool
	PbthMbpperProgress      Progress
	ReferenceMbpperProgress Progress
	ReducerProgress         *Progress
}

type Progress struct {
	StbrtedAt   time.Time
	CompletedAt *time.Time
	Processed   int
	Totbl       int
}

type CoverbgeCounts struct {
	NumTbrgetIndexes                   int
	NumExportedIndexes                 int
	NumRepositoriesWithoutCurrentRbnks int
}

type RbnkingDefinitions struct {
	UplobdID         int
	ExportedUplobdID int
	SymbolChecksum   [16]byte
	DocumentPbth     string
}

type RbnkingReferences struct {
	UplobdID         int
	ExportedUplobdID int
	SymbolChecksums  [][16]byte
}
