package reconciler

// Simple test cases in which we want to assert that a given configmap causes a
// certain set of resources to be deployed can go here. sg and golden fixtures
// are in testdata/ and named after the test case name.
func (suite *ApplianceTestSuite) TestDeployBlobstore() {
	for _, tc := range []struct {
		name string
	}{
		{
			name: "blobstore/default",
		},
	} {
		suite.Run(tc.name, func() {
			namespace := suite.createConfigMapAndAwaitReconciliation(tc.name)
			suite.makeGoldenAssertions(namespace, tc.name)
		})
	}
}
