package reconciler

func (suite *ApplianceTestSuite) TestDeployOtelCollector() {
	for _, tc := range []struct {
		name string
	}{
		{name: "otel-collector/default"},
		{name: "otel-collector/with-exporters"},
		{name: "otel-collector/with-exporters-tls-secret-name"},
		{name: "otel-collector/with-jaeger"},
	} {
		suite.Run(tc.name, func() {
			namespace := suite.createConfigMapAndAwaitReconciliation(tc.name)
			suite.makeGoldenAssertions(namespace, tc.name)
		})
	}
}
