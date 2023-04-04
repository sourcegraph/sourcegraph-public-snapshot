package shared

type RankingDefinitions struct {
	UploadID     int
	SymbolName   string
	DocumentPath string
}

type RankingReferences struct {
	UploadID    int
	SymbolNames []string
}
