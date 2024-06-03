package reconciler

func (suite *ApplianceTestSuite) TestDeployRedis() {
	for _, tc := range []struct {
		name string
	}{
		{name: "redis/default"},
	} {
		suite.Run(tc.name, func() {
			namespace := suite.createConfigMapAndAwaitReconciliation(tc.name)
			suite.makeGoldenAssertions(namespace, tc.name)
		})
	}
}
