package reconciler

func (suite *ApplianceTestSuite) TestDeployGrafana() {
	for _, tc := range []struct {
		name string
	}{
		{name: "grafana/default"},
		{name: "grafana/with-replicas"},
		{name: "grafana/with-existing-configmap"},
	} {
		suite.Run(tc.name, func() {
			namespace := suite.createConfigMapAndAwaitReconciliation(tc.name)
			suite.makeGoldenAssertions(namespace, tc.name)
		})
	}
}
