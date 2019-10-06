package internal

type dbMocks struct {
	Comments mockComments
}

var Mocks dbMocks

func ResetMocks() {
	Mocks = dbMocks{}
}
