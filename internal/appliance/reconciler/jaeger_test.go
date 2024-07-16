package reconciler

func (suite *ApplianceTestSuite) TestDeployJaeger() {
	for _, tc := range []struct {
		name string
	}{
		{name: "jaeger/default"},
		{name: "jaeger/with-replicas"},
	} {
		suite.Run(tc.name, func() {
			namespace := suite.createConfigMapAndAwaitReconciliation(tc.name)
			suite.makeGoldenAssertions(namespace, tc.name)
		})
	}
}
