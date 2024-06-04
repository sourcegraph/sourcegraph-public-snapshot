package reconciler

func (suite *ApplianceTestSuite) TestDeployPGSQL() {
	for _, tc := range []struct {
		name string
	}{
		{name: "pgsql/default"},
	} {
		suite.Run(tc.name, func() {
			namespace := suite.createConfigMapAndAwaitReconciliation(tc.name)
			suite.makeGoldenAssertions(namespace, tc.name)
		})
	}
}
