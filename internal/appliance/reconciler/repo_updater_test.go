package reconciler

func (suite *ApplianceTestSuite) TestDeployRepoUpdater() {
	for _, tc := range []struct {
		name string
	}{
		{name: "repo-updater/default"},
	} {
		suite.Run(tc.name, func() {
			namespace := suite.createConfigMap(tc.name)
			suite.awaitReconciliation(namespace)
			suite.makeGoldenAssertions(namespace, tc.name)
		})
	}
}
