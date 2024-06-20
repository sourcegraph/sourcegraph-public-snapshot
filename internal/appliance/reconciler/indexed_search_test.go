package reconciler

func (suite *ApplianceTestSuite) TestDeployIndexedSearch() {
	for _, tc := range []struct {
		name string
	}{
		{name: "indexed-search/default"},
		{name: "indexed-search/with-replicas"},
	} {
		suite.Run(tc.name, func() {
			namespace := suite.createConfigMapAndAwaitReconciliation(tc.name)
			suite.makeGoldenAssertions(namespace, tc.name)
		})
	}
}
