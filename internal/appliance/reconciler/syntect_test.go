package reconciler

func (suite *ApplianceTestSuite) TestDeploySyntect() {
	for _, tc := range []struct {
		name string
	}{
		{name: "syntect/default"},
		{name: "syntect/with-replicas"},
	} {
		suite.Run(tc.name, func() {
			namespace := suite.createConfigMap(tc.name)
			suite.awaitReconciliation(namespace)
			suite.makeGoldenAssertions(namespace, tc.name)
		})
	}
}
