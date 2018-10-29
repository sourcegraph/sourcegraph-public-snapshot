package productsubscription

func resetMocks() {
	mocks = dbMocks{}
}

type dbMocks struct {
	subscriptions mockSubscriptions
	licenses      mockLicenses
}

var mocks dbMocks
