package reconciler

func (suite *ApplianceTestSuite) TestDeployGitServer() {
	for _, tc := range []struct {
		name string
	}{
		{name: "gitserver/default"},

		{name: "gitserver/default-5.4"},
		{name: "gitserver/with-cache-size-5.4"},

		{name: "gitserver/default-5.5"},
	} {
		suite.Run(tc.name, func() {
			namespace := suite.createConfigMapAndAwaitReconciliation(tc.name)
			suite.makeGoldenAssertions(namespace, tc.name)
		})
	}
}

func (suite *ApplianceTestSuite) TestDeployGitServerUpgrade_5_3_9104_to_5_4_X() {
	namespace := suite.createConfigMapAndAwaitReconciliation("gitserver/default")
	suite.updateConfigMapAndAwaitReconciliation(namespace, "gitserver/default-5.4")
	suite.makeGoldenAssertions(namespace, "gitserver/upgrade/5.3.9104-to-5.4.X")
}

func (suite *ApplianceTestSuite) TestDeployGitServerUpgrade_5_4_1234_to_5_5_X() {
	namespace := suite.createConfigMapAndAwaitReconciliation("gitserver/default-5.4")
	suite.updateConfigMapAndAwaitReconciliation(namespace, "gitserver/default-5.5")
	suite.makeGoldenAssertions(namespace, "gitserver/upgrade/5.4.1234-to-5.5.X")
}
