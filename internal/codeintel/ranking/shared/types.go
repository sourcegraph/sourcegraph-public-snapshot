package shared

import "time"

type Summary struct {
	GraphKey                string
	VisibleToZoekt          bool
	PathMapperProgress      Progress
	ReferenceMapperProgress Progress
	ReducerProgress         *Progress
}

type Progress struct {
	StartedAt   time.Time
	CompletedAt *time.Time
	Processed   int
	Total       int
}

type CoverageCounts struct {
	NumTargetIndexes                   int
	NumExportedIndexes                 int
	NumRepositoriesWithoutCurrentRanks int
}

type RankingDefinitions struct {
	UploadID         int
	ExportedUploadID int
	SymbolChecksum   [16]byte
	DocumentPath     string
}

type RankingReferences struct {
	UploadID         int
	ExportedUploadID int
	SymbolChecksums  [][16]byte
}
