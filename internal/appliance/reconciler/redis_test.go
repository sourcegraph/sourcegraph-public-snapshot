package reconciler

func (suite *ApplianceTestSuite) TestDeployRedis() {
	for _, tc := range []struct {
		name string
	}{
		{name: "redis/default"},
	} {
		suite.Run(tc.name, func() {
			namespace := suite.createConfigMap(tc.name)
			suite.awaitReconciliation(namespace)
			suite.makeGoldenAssertions(namespace, tc.name)
		})
	}
}
