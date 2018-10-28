package registry

func resetMocks() {
	mocks = dbMocks{}
}

type dbMocks struct {
	extensions mockExtensions
	releases   mockReleases
}

var mocks dbMocks
