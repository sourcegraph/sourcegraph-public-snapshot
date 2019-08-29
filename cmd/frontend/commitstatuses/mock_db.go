package commitstatuses

type dbMocks struct {
	commitStatuses mockCommitStatuses
}

var mocks dbMocks

func resetMocks() {
	mocks = dbMocks{}
}
