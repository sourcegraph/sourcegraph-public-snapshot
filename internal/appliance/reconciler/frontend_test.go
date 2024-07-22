package reconciler

import (
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

	cfgMap := suite.newConfigMap(namespace.GetName(), "frontend/adopt-service")
	suite.awaitReconciliation(namespace.GetName(), func() {
		_, err := suite.k8sClient.CoreV1().ConfigMaps(namespace.GetName()).Create(suite.ctx, cfgMap, metav1.CreateOptions{})
		suite.Require().NoError(err)
	})
	suite.makeGoldenAssertions(namespace.GetName(), "frontend/adopt-service")
}
