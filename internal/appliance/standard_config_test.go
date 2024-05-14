package appliance

import "time"

// Use this file to test features available in StandardConfig (see
// development.md and config subpackage).

func (suite *ApplianceTestSuite) TestStandardFeatures() {
	for _, tc := range []struct {
		name string
	}{
		{name: "standard/repo-updater-with-pod-template-config"},
		{name: "standard/repo-updater-with-resources"},
		{name: "standard/repo-updater-with-no-resources"},
		{name: "standard/repo-updater-with-sa-annotations"},
		{name: "standard/symbols-with-custom-image"},
		{name: "standard/redis-with-multiple-custom-images"},
		{name: "standard/precise-code-intel-with-env-vars"},
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

// More complex test cases involving updates to the configmap can have their own
// test blocks
func (suite *ApplianceTestSuite) TestResourcesDeletedWhenDisabled() {
	namespace := suite.createConfigMap("blobstore/default")
	suite.Require().Eventually(func() bool {
		return suite.getConfigMapReconcileEventCount(namespace) > 0
	}, time.Second*10, time.Millisecond*200)

	eventsSeenSoFar := suite.getConfigMapReconcileEventCount(namespace)
	suite.updateConfigMap(namespace, "standard/everything-disabled")
	suite.Require().Eventually(func() bool {
		return suite.getConfigMapReconcileEventCount(namespace) > eventsSeenSoFar
	}, time.Second*10, time.Millisecond*200)

	suite.makeGoldenAssertions(namespace, "standard/blobstore-subsequent-disable")
}
