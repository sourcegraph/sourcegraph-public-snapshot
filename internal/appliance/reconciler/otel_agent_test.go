package reconciler

func (suite *ApplianceTestSuite) TestDeployOtelAgent() {
	for _, tc := range []struct {
		name string
	}{
		{name: "otel-agent/default"},
	} {
		suite.Run(tc.name, func() {
			namespace := suite.createConfigMapAndAwaitReconciliation(tc.name)
			suite.makeGoldenAssertions(namespace, tc.name)
		})
	}
}
