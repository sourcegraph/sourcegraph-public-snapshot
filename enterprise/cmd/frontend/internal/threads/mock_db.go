package threads

type dbMocks struct {
	threads mockThreads
}

var mocks dbMocks

func resetMocks() {
	mocks = dbMocks{}
}
