package projects

type dbMocks struct {
	projects mockProjects
}

var mocks dbMocks

func resetMocks() {
	mocks = dbMocks{}
}
