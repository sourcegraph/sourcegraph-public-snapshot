package shared

import "time"

type Summary struct {
	GraphKey                string
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

type RankingDefinitions struct {
	UploadID         int
	ExportedUploadID int
	SymbolName       string
	DocumentPath     string
}

type RankingReferences struct {
	UploadID         int
	ExportedUploadID int
	SymbolNames      []string
}
