package appliance

import "time"

// Use this file to test features available in StandardConfig (see
// development.md and config subpackage).

func (suite *ApplianceTestSuite) TestStandardFeatures() {
	for _, tc := range []struct {
		name string
	}{
		{name: "repo-updater-with-pod-template-config"},
		{name: "repo-updater-with-resources"},
		{name: "repo-updater-with-no-resources"},
		{name: "repo-updater-with-sa-annotations"},
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
	namespace := suite.createConfigMap("blobstore-default")
	suite.Require().Eventually(func() bool {
		return suite.getConfigMapReconcileEventCount(namespace) > 0
	}, time.Second*10, time.Millisecond*200)

	eventsSeenSoFar := suite.getConfigMapReconcileEventCount(namespace)
	suite.updateConfigMap(namespace, "everything-disabled")
	suite.Require().Eventually(func() bool {
		return suite.getConfigMapReconcileEventCount(namespace) > eventsSeenSoFar
	}, time.Second*10, time.Millisecond*200)

	suite.makeGoldenAssertions(namespace, "blobstore-subsequent-disable")
}
