package reconciler

func (suite *ApplianceTestSuite) TestDeployGitServer() {
	for _, tc := range []struct {
		name string
	}{
		{name: "gitserver/default"},
	} {
		suite.Run(tc.name, func() {
			namespace := suite.createConfigMapAndAwaitReconciliation(tc.name)
			suite.makeGoldenAssertions(namespace, tc.name)
		})
	}
}
