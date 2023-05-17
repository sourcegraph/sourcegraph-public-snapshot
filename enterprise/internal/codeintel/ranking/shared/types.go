package shared

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
