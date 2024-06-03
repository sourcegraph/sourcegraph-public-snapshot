package reconciler

func (suite *ApplianceTestSuite) TestDeploySymbols() {
	for _, tc := range []struct {
		name string
	}{
		{name: "symbols/default"},

		// This service does some logic on the storage quantity, so we can't
		// just rely on the standard config test for storage amounts/classes.
		{name: "symbols/with-storage"},
	} {
		suite.Run(tc.name, func() {
			namespace := suite.createConfigMapAndAwaitReconciliation(tc.name)
			suite.makeGoldenAssertions(namespace, tc.name)
		})
	}
}
