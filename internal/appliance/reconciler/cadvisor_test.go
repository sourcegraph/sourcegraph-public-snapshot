package reconciler

func (suite *ApplianceTestSuite) TestDeployCadvisor() {
	for _, tc := range []struct {
		name string
	}{
		{name: "cadvisor/default"},
	} {
		suite.Run(tc.name, func() {
			namespace := suite.createConfigMapAndAwaitReconciliation(tc.name)
			suite.makeGoldenAssertions(namespace, tc.name)
		})
	}
}
