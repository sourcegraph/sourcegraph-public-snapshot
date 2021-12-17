package amplitude

// DenyList is a list of events we do not want to send to Amplitude.
// We will normally add events to the deny list if they are high frequency and low signal.
var DenyList = map[string]struct{}{
	"search.latencies.regexp":           {},
	"codeintel.searchHover":             {},
	"codeintel.lsifDocumentHighlight":   {},
	"search.latencies.symbol":           {},
	"search.latencies.literal":          {},
	"codeintel.lsifHover":               {},
	"codeintel.lsifDefinitions":         {},
	"search.latencies.commit":           {},
	"search.latencies.structural":       {},
	"goToDefinition.preloaded":          {},
	"codeintel.searchReferences":        {},
	"search.latencies.file":             {},
	"codeintel.lsifReferences":          {},
	"search.latencies.repo":             {},
	"search.latencies.diff":             {},
	"codeintel.lsifDefinitions.xrepo":   {},
	"codeintel.lsifReferences.xrepo":    {},
	"codeintel.searchReferences.xrepo":  {},
	"codeintel.searchDefinitions.xrepo": {},
	"codeintel.lspDefinitions":          {},
	"codeintel.lspHover":                {},
	"codeintel.lspReferences":           {},
}
