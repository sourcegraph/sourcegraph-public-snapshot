package reconciler

func (suite *ApplianceTestSuite) TestDeployNodeExporter() {
	for _, tc := range []struct {
		name string
	}{
		{name: "nodeexporter/default"},
	} {
		suite.Run(tc.name, func() {
			namespace := suite.createConfigMapAndAwaitReconciliation(tc.name)
			suite.makeGoldenAssertions(namespace, tc.name)
		})
	}
}
