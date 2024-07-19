package reconciler

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (suite *ApplianceTestSuite) TestDeployFrontend() {
	for _, tc := range []struct {
		name string
	}{
		{name: "frontend/default"},
		{name: "frontend/with-blobstore"},
		{name: "frontend/with-ingress"},
		{name: "frontend/with-ingress-optional-fields"},
		{name: "frontend/with-overrides"},
	} {
		suite.Run(tc.name, func() {
			namespace := suite.createConfigMapAndAwaitReconciliation(tc.name)
			suite.makeGoldenAssertions(namespace, tc.name)
		})
	}
}

func (suite *ApplianceTestSuite) TestFrontendDeploymentRollsWhenPGSecretChanges() {
	// Create the frontend before the PGSQL secret exists. In general, this
	// might happen, depending on the order of the reconcile loop. If we
	// introducce concurrency to this, we'll have little control over what
	// happens first.
	namespace := suite.createConfigMapAndAwaitReconciliation("frontend/default")

	// Create the PGSQL secret.
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: pgsqlSecretName,
		},
		StringData: map[string]string{
			"host":     "example.com",
			"port":     "5432",
			"user":     "alice",
			"password": "letmein",
			"database": "sg",
		},
	}
	_, err := suite.k8sClient.CoreV1().Secrets(namespace).Create(suite.ctx, secret, metav1.CreateOptions{})
	suite.Require().NoError(err)

	// We have to make a config change to trigger the reconcile loop
	suite.awaitReconciliation(namespace, func() {
		cfgMap := suite.newConfigMap(namespace, "frontend/default")
		cfgMap.GetAnnotations()["force-reconcile"] = "1"
		_, err := suite.k8sClient.CoreV1().ConfigMaps(namespace).Update(suite.ctx, cfgMap, metav1.UpdateOptions{})
		suite.Require().NoError(err)
	})

	suite.makeGoldenAssertions(namespace, "frontend/after-create-pg-secret")
}
