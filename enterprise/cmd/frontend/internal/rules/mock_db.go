package rules

type dbMocks struct {
	rules mockRules
}

var mocks dbMocks

func resetMocks() {
	mocks = dbMocks{}
}
