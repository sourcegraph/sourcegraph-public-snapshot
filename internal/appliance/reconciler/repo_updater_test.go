package reconciler

func (suite *ApplianceTestSuite) TestDeployRepoUpdater() {
	for _, tc := range []struct {
		name string
	}{
		{name: "repo-updater/default"},
	} {
		suite.Run(tc.name, func() {
			namespace := suite.createConfigMapAndAwaitReconciliation(tc.name)
			suite.makeGoldenAssertions(namespace, tc.name)
		})
	}
}
