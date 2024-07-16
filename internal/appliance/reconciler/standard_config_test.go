package reconciler

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/sourcegraph/sourcegraph/internal/appliance/k8senvtest"
)

// Use this file to test features available in StandardConfig (see
// development.md and config subpackage).

func (suite *ApplianceTestSuite) TestStandardFeatures() {
	for _, tc := range []struct {
		name string
	}{
		{name: "standard/blobstore-with-named-storage-class"},
		{name: "standard/frontend-with-no-cpu-memory-resources"},
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

func (suite *ApplianceTestSuite) TestDoesNotDeleteUnownedResources() {
	namespace, err := k8senvtest.NewRandomNamespace("test-appliance")
	suite.Require().NoError(err)
	_, err = suite.k8sClient.CoreV1().Namespaces().Create(suite.ctx, namespace, metav1.CreateOptions{})
	suite.Require().NoError(err)

	// Example: the admin configures a pgsql secret that references an external
	// database, and therefore disables pgsql in appliance config.
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
	_, err = suite.k8sClient.CoreV1().Secrets(namespace.Name).Create(suite.ctx, secret, metav1.CreateOptions{})
	suite.Require().NoError(err)

	suite.awaitReconciliation(namespace.Name, func() {
		// This is an artificial test fixture that disables everything except
		// frontend, but this is representative of disabling pgsql.
		cfgMap := suite.newConfigMap(namespace.Name, "frontend/default")
		_, err := suite.k8sClient.CoreV1().ConfigMaps(namespace.GetName()).Create(suite.ctx, cfgMap, metav1.CreateOptions{})
		suite.Require().NoError(err)
	})

	secretStillPresent, err := suite.k8sClient.CoreV1().Secrets(namespace.Name).Get(suite.ctx, pgsqlSecretName, metav1.GetOptions{})
	suite.Require().NoError(err)
	suite.Require().Equal("example.com", string(secretStillPresent.Data["host"]))
}
