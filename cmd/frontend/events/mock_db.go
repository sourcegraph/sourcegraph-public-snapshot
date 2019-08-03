package events

type dbMocks struct {
	events        mockEvents
	eventsThreads mockEventsThreads
}

var mocks dbMocks

func resetMocks() {
	mocks = dbMocks{}
}
