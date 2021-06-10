package zoekt

type IndexedRequestType string

const (
	TextRequest   IndexedRequestType = "text"
	SymbolRequest IndexedRequestType = "symbol"
)
