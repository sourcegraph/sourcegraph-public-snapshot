package appliance

import (
	"time"
)

func (suite *ApplianceTestSuite) TestDeployRepoUpdater() {
	for _, tc := range []struct {
		name string
	}{
		{name: "repo-updater-default"},
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

func (suite *ApplianceTestSuite) TestRepoUpdaterResourcesDeletedWhenDisabled() {
	namespace := suite.createConfigMap("repo-updater-default")
	suite.Require().Eventually(func() bool {
		return suite.getConfigMapReconcileEventCount(namespace) > 0
	}, time.Second*10, time.Millisecond*200)

	eventsSeenSoFar := suite.getConfigMapReconcileEventCount(namespace)
	suite.updateConfigMap(namespace, "everything-disabled")
	suite.Require().Eventually(func() bool {
		return suite.getConfigMapReconcileEventCount(namespace) > eventsSeenSoFar
	}, time.Second*10, time.Millisecond*200)

	suite.makeGoldenAssertions(namespace, "repo-updater-subsequent-disable")
}
