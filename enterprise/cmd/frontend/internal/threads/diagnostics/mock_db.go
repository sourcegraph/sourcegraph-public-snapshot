package diagnostics

type dbMocks struct {
	threadsDiagnostics mockThreadsDiagnostics
}

var mocks dbMocks

func resetMocks() {
	mocks = dbMocks{}
}
