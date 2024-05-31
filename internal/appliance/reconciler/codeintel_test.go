package reconciler

func (suite *ApplianceTestSuite) TestDeployCodeIntel() {
	for _, tc := range []struct {
		name string
	}{
		{name: "codeintel/default"},
	} {
		suite.Run(tc.name, func() {
			namespace := suite.createConfigMap(tc.name)
			suite.awaitReconciliation(namespace)
			suite.makeGoldenAssertions(namespace, tc.name)
		})
	}
}
