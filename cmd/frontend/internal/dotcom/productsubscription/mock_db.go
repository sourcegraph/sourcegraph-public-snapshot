package productsubscription

type dbMocks struct {
	subscriptions mockSubscriptions
	licenses      mockLicenses
}

var mocks dbMocks

func pointify[T any](v T) *T { return &v }
