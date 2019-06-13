package rules

type dbMocks struct {
	Rules mockRules
}

var Mocks dbMocks

func ResetMocks() {
	Mocks = dbMocks{}
}
