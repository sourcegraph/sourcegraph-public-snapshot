package changesets

type dbMocks struct {
	changesets mockChangesets
}

var mocks dbMocks

func resetMocks() {
	mocks = dbMocks{}
}
