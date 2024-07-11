package reconciler

func (suite *ApplianceTestSuite) TestDeployPreciseCodeIntel() {
	for _, tc := range []struct {
		name string
	}{
		{name: "precise-code-intel/default"},
		{name: "precise-code-intel/with-blobstore"},
		{name: "precise-code-intel/with-num-workers"},
		{name: "precise-code-intel/with-replicas"},
	} {
		suite.Run(tc.name, func() {
			namespace := suite.createConfigMapAndAwaitReconciliation(tc.name)
			suite.makeGoldenAssertions(namespace, tc.name)
		})
	}
}
