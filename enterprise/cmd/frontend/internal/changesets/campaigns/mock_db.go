package campaigns

type dbMocks struct {
	campaigns mockChangesetCampaigns
}

var mocks dbMocks

func resetMocks() {
	mocks = dbMocks{}
}
