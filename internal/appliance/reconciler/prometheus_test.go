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
			namespace := suite.createConfigMap(tc.name)
			suite.awaitReconciliation(namespace)
			suite.makeGoldenAssertions(namespace, tc.name)
		})
	}
}

func (suite *ApplianceTestSuite) TestNonNamespacedResourcesRemainWhenDisabled() {
	namespace := suite.createConfigMap("prometheus/privileged")
	suite.awaitReconciliation(namespace)

	suite.updateConfigMap(namespace, "standard/everything-disabled")
	suite.awaitReconciliation(namespace)
	suite.makeGoldenAssertions(namespace, "prometheus/subsequent-disable")
}
