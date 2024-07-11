package reconciler

func (suite *ApplianceTestSuite) TestDeploySearcher() {
	for _, tc := range []struct {
		name string
	}{
		{name: "searcher/default"},
		{name: "searcher/with-replicas"},

		// This service does some logic on the storage quantity, so we can't
		// just rely on the standard config test for storage amounts/classes.
		{name: "searcher/with-storage"},
	} {
		suite.Run(tc.name, func() {
			namespace := suite.createConfigMapAndAwaitReconciliation(tc.name)
			suite.makeGoldenAssertions(namespace, tc.name)
		})
	}
}
