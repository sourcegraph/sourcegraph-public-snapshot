package reconciler

// Use this file to test features available in StandardConfig (see
// development.md and config subpackage).

func (suite *ApplianceTestSuite) TestStandardFeatures() {
	for _, tc := range []struct {
		name string
	}{
		{name: "standard/blobstore-with-named-storage-class"},
		{name: "standard/precise-code-intel-with-env-vars"},
		{name: "standard/redis-with-multiple-custom-images"},
		{name: "standard/redis-with-storage"},
		{name: "standard/repo-updater-with-no-resources"},
		{name: "standard/repo-updater-with-pod-template-config"},
		{name: "standard/repo-updater-with-resources"},
		{name: "standard/repo-updater-with-sa-annotations"},
		{name: "standard/symbols-with-custom-image"},
	} {
		suite.Run(tc.name, func() {
			namespace := suite.createConfigMapAndAwaitReconciliation(tc.name)
			suite.makeGoldenAssertions(namespace, tc.name)
		})
	}
}

// More complex test cases involving updates to the configmap can have their own
// test blocks
func (suite *ApplianceTestSuite) TestResourcesDeletedWhenDisabled() {
	namespace := suite.createConfigMapAndAwaitReconciliation("blobstore/default")

	suite.updateConfigMapAndAwaitReconciliation(namespace, "standard/everything-disabled")
	suite.makeGoldenAssertions(namespace, "standard/blobstore-subsequent-disable")
}
