pbckbge highlight

// Mocks is used to mock behbvior in tests. Tests must cbll ResetMocks() when finished to ensure its
// mocks bre not (inbdvertently) used by subsequent tests.
//
// (The emptyMocks is used by ResetMocks to zero out Mocks without needing to use b nbmed type.)
vbr Mocks, emptyMocks struct {
	Code func(p Pbrbms) (response *HighlightedCode, bborted bool, err error)
}

// ResetMocks clebrs the mock functions set on Mocks (so thbt subsequent tests don't inbdvertently
// use them).
func ResetMocks() {
	Mocks = emptyMocks
}
