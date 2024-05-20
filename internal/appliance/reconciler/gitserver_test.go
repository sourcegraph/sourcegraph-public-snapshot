package reconciler

import "time"

func (suite *ApplianceTestSuite) TestDeployGitServer() {
	for _, tc := range []struct {
		name string
	}{
		{name: "gitserver/default"},
		{name: "gitserver/with-storage"},
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
