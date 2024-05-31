package reconciler

func (suite *ApplianceTestSuite) TestDeployPGSQL() {
	for _, tc := range []struct {
		name string
	}{
		{name: "pgsql/default"},
	} {
		suite.Run(tc.name, func() {
			namespace := suite.createConfigMap(tc.name)
			suite.awaitReconciliation(namespace)
			suite.makeGoldenAssertions(namespace, tc.name)
		})
	}
}
