package campaigns

type dbMocks struct {
	campaigns mockCampaigns
}

var mocks dbMocks

func resetMocks() {
	mocks = dbMocks{}
}
