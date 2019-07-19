package campaigns

type dbMocks struct {
	campaigns        mockCampaigns
	campaignsThreads mockCampaignsThreads
}

var mocks dbMocks

func resetMocks() {
	mocks = dbMocks{}
}
