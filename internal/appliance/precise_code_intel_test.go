package appliance

import "time"

func (suite *ApplianceTestSuite) TestDeployPreciseCodeIntel() {
	for _, tc := range []struct {
		name string
	}{
		{name: "precise-code-intel/default"},
		{name: "precise-code-intel/with-blobstore"},
		{name: "precise-code-intel/with-num-workers"},
		{name: "precise-code-intel/with-replicas"},
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
