package internal

type dbMocks struct {
	Threads mockThreads
}

var Mocks dbMocks

func ResetMocks() {
	Mocks = dbMocks{}
}
