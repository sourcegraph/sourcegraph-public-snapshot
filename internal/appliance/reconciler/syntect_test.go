package reconciler

func (suite *ApplianceTestSuite) TestDeploySyntect() {
	for _, tc := range []struct {
		name string
	}{
		{name: "syntect/default"},
		{name: "syntect/with-replicas"},
	} {
		suite.Run(tc.name, func() {
			namespace := suite.createConfigMapAndAwaitReconciliation(tc.name)
			suite.makeGoldenAssertions(namespace, tc.name)
		})
	}
}
