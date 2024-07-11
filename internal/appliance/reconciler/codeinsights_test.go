package reconciler

func (suite *ApplianceTestSuite) TestDeployCodeInsights() {
	for _, tc := range []struct {
		name string
	}{
		{name: "codeinsights/default"},
	} {
		suite.Run(tc.name, func() {
			namespace := suite.createConfigMapAndAwaitReconciliation(tc.name)
			suite.makeGoldenAssertions(namespace, tc.name)
		})
	}
}
