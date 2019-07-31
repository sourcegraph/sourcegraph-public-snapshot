package comments

type dbMocks struct {
	comments        mockComments
	commentsThreads mockCommentsThreads
}

var mocks dbMocks

func resetMocks() {
	mocks = dbMocks{}
}
