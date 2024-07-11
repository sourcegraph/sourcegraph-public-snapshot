package reconciler

func (suite *ApplianceTestSuite) TestDeployPrometheus() {
	for _, tc := range []struct {
		name string
	}{
		{name: "prometheus/default"},
		{name: "prometheus/privileged"},
		{name: "prometheus/with-existing-configmap"},
	} {
		suite.Run(tc.name, func() {
			namespace := suite.createConfigMapAndAwaitReconciliation(tc.name)
			suite.makeGoldenAssertions(namespace, tc.name)
		})
	}
}

func (suite *ApplianceTestSuite) TestNonNamespacedResourcesRemainWhenDisabled() {
	namespace := suite.createConfigMapAndAwaitReconciliation("prometheus/privileged")
	suite.updateConfigMapAndAwaitReconciliation(namespace, "standard/everything-disabled")
	suite.makeGoldenAssertions(namespace, "prometheus/subsequent-disable")
}
