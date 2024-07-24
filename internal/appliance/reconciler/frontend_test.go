package reconciler

import (
	"fmt"
	"os"

	"github.com/sourcegraph/sourcegraph/internal/appliance/k8senvtest"
	"github.com/sourcegraph/sourcegraph/internal/k8s/resource/ingress"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
	corev1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
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

func (suite *ApplianceTestSuite) TestAdoptsHelmProvisionedFrontendResources() {
	namespace, err := k8senvtest.NewRandomNamespace("test-appliance")
	suite.Require().NoError(err)
	_, err = suite.k8sClient.CoreV1().Namespaces().Create(suite.ctx, namespace, metav1.CreateOptions{})
	suite.Require().NoError(err)
	testService := corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "sourcegraph-frontend",
			Namespace: namespace.Name,
		},
		Spec: corev1.ServiceSpec{
			Ports:    []corev1.ServicePort{{Name: "http", Port: 30080, TargetPort: intstr.FromString("http")}},
			Selector: map[string]string{"app": "sourcegraph-appliance"},
		},
	}
	_, err = suite.k8sClient.CoreV1().Services(namespace.Name).Create(suite.ctx, &testService, metav1.CreateOptions{})
	suite.Require().NoError(err)

	testIngress := ingress.NewIngress("sourcegraph-frontend", namespace.Name)
	testIngress.Spec.Rules = []netv1.IngressRule{{
		Host: "an-existing-hostname.com",
		IngressRuleValue: netv1.IngressRuleValue{
			HTTP: &netv1.HTTPIngressRuleValue{
				Paths: []netv1.HTTPIngressPath{{
					Path:     "/",
					PathType: pointers.Ptr(netv1.PathTypePrefix),
					Backend: netv1.IngressBackend{
						Service: &netv1.IngressServiceBackend{
							Name: "sourcegraph-frontend",
							Port: netv1.ServiceBackendPort{
								Number: 30081,
							},
						},
					},
				}},
			},
		},
	}}
	ingressClassName := "nginx"
	testIngress.Spec.IngressClassName = &ingressClassName
	_, err = suite.k8sClient.NetworkingV1().Ingresses(namespace.Name).Create(suite.ctx, &testIngress, metav1.CreateOptions{})
	suite.Require().NoError(err)

	cfgMap := suite.newConfigMap(namespace.GetName(), "frontend/with-ingress")
	suite.awaitReconciliation(namespace.GetName(), func() {
		_, err := suite.k8sClient.CoreV1().ConfigMaps(namespace.GetName()).Create(suite.ctx, cfgMap, metav1.CreateOptions{})
		suite.Require().NoError(err)
	})
	suite.makeGoldenAssertions(namespace.GetName(), "frontend/adopt-service")
}

func (suite *ApplianceTestSuite) TestFrontendDeploymentRollsWhenPGSecretsChange() {
	for _, tc := range []struct {
		secret string
	}{
		{secret: pgsqlSecretName},
		{secret: codeInsightsDBSecretName},
		{secret: codeIntelDBSecretName},
	} {
		suite.Run(tc.secret, func() {
			// Create the frontend before the PGSQL secret exists. In general, this
			// might happen, depending on the order of the reconcile loop. If we
			// introducce concurrency to this, we'll have little control over what
			// happens first.
			namespace := suite.createConfigMapAndAwaitReconciliation("frontend/default")

			// Create the PGSQL secret.
			secret := &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name: tc.secret,
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

			suite.makeGoldenAssertions(namespace, fmt.Sprintf("frontend/after-create-%s-secret", tc.secret))
		})
	}
}

func (suite *ApplianceTestSuite) TestFrontendDeploymentRollsWhenRedisSecretsChange() {
	for _, tc := range []struct {
		secret string
	}{
		{secret: redisCacheSecretName},
		{secret: redisStoreSecretName},
	} {
		suite.Run(tc.secret, func() {
			// Create the frontend before the PGSQL secret exists. In general, this
			// might happen, depending on the order of the reconcile loop. If we
			// introducce concurrency to this, we'll have little control over what
			// happens first.
			namespace := suite.createConfigMapAndAwaitReconciliation("frontend/default")

			// Create the PGSQL secret.
			secret := &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name: tc.secret,
				},
				StringData: map[string]string{
					"endpoint": "example.com",
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

			suite.makeGoldenAssertions(namespace, fmt.Sprintf("frontend/after-create-%s-secret", tc.secret))
		})
	}
}

func (suite *ApplianceTestSuite) TestMergeK8sObjects() {
	// Create a temporary JSON file
	tempFile, err := os.CreateTemp("", "test-service-*.json")
	suite.Require().NoError(err)
	defer os.Remove(tempFile.Name())

	// Define the JSON content
	jsonContent := `{
		"apiVersion": "v1",
		"kind": "Service",
		"metadata": {
			"name": "test-service",
			"labels": {
				"newLabel": "newValue"
			}
		},
		"spec": {
			"ports": [
				{
					"port": 8080,
					"targetPort": 8080
				}
			],
			"selector": {
				"app": "test"
			}
		}
	}`

	// Write JSON content to the temporary file
	err = os.WriteFile(tempFile.Name(), []byte(jsonContent), 0644)
	suite.Require().NoError(err)

	// Create an existing Service object
	existingService := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: "existing-service",
			Labels: map[string]string{
				"existingLabel": "existingValue",
			},
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Port:       80,
					TargetPort: intstr.FromInt(80),
				},
			},
		},
	}

	// Call the MergeK8sObjects function
	mergedObj, err := MergeK8sObjects(tempFile.Name(), existingService)
	suite.Require().NoError(err)

	// Assert that the merged object is a Service
	mergedService, ok := mergedObj.(*corev1.Service)
	suite.Require().True(ok, "Merged object should be a *corev1.Service")

	// Verify the merged results
	suite.Assert().Equal("test-service", mergedService.Name)
	suite.Assert().Equal(2, len(mergedService.Labels))
	suite.Assert().Equal("existingValue", mergedService.Labels["existingLabel"])
	suite.Assert().Equal("newValue", mergedService.Labels["newLabel"])
	suite.Assert().Equal(2, len(mergedService.Spec.Ports))
	suite.Assert().Equal(int32(80), mergedService.Spec.Ports[0].Port)
	suite.Assert().Equal(int32(8080), mergedService.Spec.Ports[1].Port)
	suite.Assert().Equal("test", mergedService.Spec.Selector["app"])

	// Verify that the original object was not modified
	suite.Assert().Equal("existing-service", existingService.Name)
	suite.Assert().Equal(1, len(existingService.Labels))
	suite.Assert().Equal(1, len(existingService.Spec.Ports))
}
