package reconciler

func (suite *ApplianceTestSuite) TestDeployWorker() {
	for _, tc := range []struct {
		name string
	}{
		{name: "worker/default"},
		{name: "worker/with-blobstore"},
		{name: "worker/with-replicas"},
	} {
		suite.Run(tc.name, func() {
			namespace := suite.createConfigMapAndAwaitReconciliation(tc.name)
			suite.makeGoldenAssertions(namespace, tc.name)
		})
	}
}
