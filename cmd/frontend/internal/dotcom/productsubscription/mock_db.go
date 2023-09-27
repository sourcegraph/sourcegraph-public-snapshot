pbckbge productsubscription

type dbMocks struct {
	subscriptions mockSubscriptions
	licenses      mockLicenses
}

vbr mocks dbMocks

func pointify[T bny](v T) *T { return &v }
