package reconciler

import "time"

func (suite *ApplianceTestSuite) TestDeployPrometheus() {
	for _, tc := range []struct {
		name string
	}{
		{name: "prometheus/default"},
		{name: "prometheus/privileged"},
		{name: "prometheus/with-existing-configmap"},
		{name: "prometheus/with-storage"},
	} {
		suite.Run(tc.name, func() {
			namespace := suite.createConfigMap(tc.name)

			// Wait for reconciliation to be finished.
			suite.Require().Eventually(func() bool {
				return suite.getConfigMapReconcileEventCount(namespace) > 0
			}, time.Second*10, time.Millisecond*200)

			suite.makeGoldenAssertions(namespace, tc.name)
		})
	}
}

func (suite *ApplianceTestSuite) TestNonNamespacedResourcesRemainWhenDisabled() {
	namespace := suite.createConfigMap("prometheus/privileged")
	suite.Require().Eventually(func() bool {
		return suite.getConfigMapReconcileEventCount(namespace) > 0
	}, time.Second*10, time.Millisecond*200)

	eventsSeenSoFar := suite.getConfigMapReconcileEventCount(namespace)
	suite.updateConfigMap(namespace, "standard/everything-disabled")
	suite.Require().Eventually(func() bool {
		return suite.getConfigMapReconcileEventCount(namespace) > eventsSeenSoFar
	}, time.Second*10, time.Millisecond*200)

	suite.makeGoldenAssertions(namespace, "prometheus/subsequent-disable")
}
