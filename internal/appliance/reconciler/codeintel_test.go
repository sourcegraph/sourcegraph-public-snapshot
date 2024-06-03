package reconciler

func (suite *ApplianceTestSuite) TestDeployCodeIntel() {
	for _, tc := range []struct {
		name string
	}{
		{name: "codeintel/default"},
	} {
		suite.Run(tc.name, func() {
			namespace := suite.createConfigMapAndAwaitReconciliation(tc.name)
			suite.makeGoldenAssertions(namespace, tc.name)
		})
	}
}
