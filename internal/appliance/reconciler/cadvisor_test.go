package reconciler

func (suite *ApplianceTestSuite) TestDeployCadvisor() {
	for _, tc := range []struct {
		name string
	}{
		{name: "cadvisor/default"},
	} {
		suite.Run(tc.name, func() {
			namespace := suite.createConfigMap(tc.name)
			suite.awaitReconciliation(namespace)
			suite.makeGoldenAssertions(namespace, tc.name)
		})
	}
}
