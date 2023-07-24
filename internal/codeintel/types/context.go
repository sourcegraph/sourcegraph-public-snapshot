package types

type PreciseContext struct {
	ScipSymbolName  string
	FuzzySymbolName string
	RepositoryName  string
	SymbolRole      int32
	Confidence      string
	Text            string
	FilePath        string
}
