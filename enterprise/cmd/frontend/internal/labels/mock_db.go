package labels

type dbMocks struct {
	labels        mockLabels
	labelsObjects mockLabelsObjects
}

var mocks dbMocks

func resetMocks() {
	mocks = dbMocks{}
}
