package reconciler

import (
	"os"
	"path/filepath"
	"regexp"
	"time"

	"github.com/life4/genesis/slices"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
	k8syaml "sigs.k8s.io/yaml"

	applianceyaml "github.com/sourcegraph/sourcegraph/internal/appliance/yaml"
)

// Test helpers

// creationTimestamp and uid need to be normalized
var magicTime = metav1.NewTime(time.Date(2024, time.April, 19, 0, 0, 0, 0, time.UTC))

var namespaceRegexp = regexp.MustCompile(`test\-appliance\-\w+`)

const normalizedString = "NORMALIZED_FOR_TESTING"

type goldenFile struct {
	Resources []client.Object `json:"resources"`
}

func (suite *ApplianceTestSuite) makeGoldenAssertions(namespace, goldenFileName string) {
	require := suite.Require()

	goldenFilePath := filepath.Join("testdata", "golden-fixtures", goldenFileName+".yaml")
	obtainedResources := goldenFile{Resources: suite.gatherResources(namespace)}
	obtainedBytes, err := k8syaml.Marshal(obtainedResources)
	require.NoError(err)
	obtainedBytes, err = applianceyaml.ConvertYAMLStringsToMultilineLiterals(obtainedBytes)
	require.NoError(err)

	if len(os.Args) > 0 && os.Args[len(os.Args)-1] == "appliance-update-golden-files" {
		err := os.MkdirAll(filepath.Dir(goldenFilePath), 0700)
		require.NoError(err)
		err = os.WriteFile(goldenFilePath, obtainedBytes, 0600)
		require.NoError(err)
	}

	goldenBytes, err := os.ReadFile(goldenFilePath)
	require.NoError(err)

	// testify prints a readable yaml diff
	require.Equal(string(goldenBytes), string(obtainedBytes))
}

// When new owned types are declared in SetupWithManager() in reconcile.go, we
// must gather them here for golden testing to be reliable.
func (suite *ApplianceTestSuite) gatherResources(namespace string) []client.Object {
	var objs []client.Object

	// We set the GVK ourselves, as this is missing from the List response:
	// https://github.com/kubernetes/client-go/issues/861
	// This makes eyeballing golden file diffs a little easier, as we can see
	// which object is being changed.
	//
	// Certain common fields must be normalized in order to make golden testing
	// work, such as the creationTimestamp and UID, which would differ every
	// test run. Some resource-specific normalizations are also performed.
	deps, err := suite.k8sClient.AppsV1().Deployments(namespace).List(suite.ctx, metav1.ListOptions{})
	suite.Require().NoError(err)
	for _, obj := range deps.Items {
		obj := obj // see exportloopref
		obj.SetGroupVersionKind(schema.GroupVersionKind{Group: "apps", Version: "v1", Kind: "Deployment"})
		normalizeObj(&obj)
		objs = append(objs, &obj)
	}
	daemonsets, err := suite.k8sClient.AppsV1().DaemonSets(namespace).List(suite.ctx, metav1.ListOptions{})
	suite.Require().NoError(err)
	for _, obj := range daemonsets.Items {
		obj := obj
		obj.SetGroupVersionKind(schema.GroupVersionKind{Group: "apps", Version: "v1", Kind: "DaemonSet"})
		normalizeObj(&obj)
		objs = append(objs, &obj)
	}
	ssets, err := suite.k8sClient.AppsV1().StatefulSets(namespace).List(suite.ctx, metav1.ListOptions{})
	suite.Require().NoError(err)
	for _, obj := range ssets.Items {
		obj := obj
		for i := range obj.Spec.VolumeClaimTemplates {
			obj.Spec.VolumeClaimTemplates[i].Namespace = normalizedString
		}
		obj.SetGroupVersionKind(schema.GroupVersionKind{Group: "apps", Version: "v1", Kind: "StatefulSet"})
		normalizeObj(&obj)
		objs = append(objs, &obj)
	}

	// Cluster-scoped resources have to be qualified by something other than
	// metadata.namespace.
	clusterRoles, err := suite.k8sClient.RbacV1().ClusterRoles().List(suite.ctx, metav1.ListOptions{
		LabelSelector: "for-namespace=" + namespace,
	})
	suite.Require().NoError(err)
	for _, obj := range clusterRoles.Items {
		obj := obj
		obj.SetName(namespaceRegexp.ReplaceAllString(obj.Name, normalizedString))
		obj.Labels["for-namespace"] = normalizedString
		obj.SetGroupVersionKind(schema.GroupVersionKind{Group: "rbac.authorization.k8s.io", Version: "v1", Kind: "ClusterRole"})
		normalizeObj(&obj)
		objs = append(objs, &obj)
	}
	clusterRoleBindings, err := suite.k8sClient.RbacV1().ClusterRoleBindings().List(suite.ctx, metav1.ListOptions{
		LabelSelector: "for-namespace=" + namespace,
	})
	suite.Require().NoError(err)
	for _, obj := range clusterRoleBindings.Items {
		obj := obj
		obj.SetName(namespaceRegexp.ReplaceAllString(obj.Name, normalizedString))
		obj.Labels["for-namespace"] = normalizedString
		obj.RoleRef.Name = namespaceRegexp.ReplaceAllString(obj.RoleRef.Name, normalizedString)
		obj.Subjects = slices.Map(obj.Subjects, func(s rbacv1.Subject) rbacv1.Subject {
			s.Namespace = normalizedString
			return s
		})
		obj.SetGroupVersionKind(schema.GroupVersionKind{Group: "rbac.authorization.k8s.io", Version: "v1", Kind: "ClusterRoleBinding"})
		normalizeObj(&obj)
		objs = append(objs, &obj)
	}

	cmaps, err := suite.k8sClient.CoreV1().ConfigMaps(namespace).List(suite.ctx, metav1.ListOptions{})
	suite.Require().NoError(err)
	for _, obj := range cmaps.Items {
		obj := obj

		// Find and replace all instances of the randomly-namd namespace in
		// configmap data. Crude, but necessary for Prometheus config.
		for file, content := range obj.Data {
			obj.Data[file] = namespaceRegexp.ReplaceAllString(content, normalizedString)
		}

		obj.SetGroupVersionKind(schema.GroupVersionKind{Group: "", Version: "v1", Kind: "ConfigMap"})
		normalizeObj(&obj)
		objs = append(objs, &obj)
	}
	pvcs, err := suite.k8sClient.CoreV1().PersistentVolumeClaims(namespace).List(suite.ctx, metav1.ListOptions{})
	suite.Require().NoError(err)
	for _, obj := range pvcs.Items {
		obj := obj
		if obj.DeletionTimestamp != nil {
			obj.DeletionTimestamp = &magicTime
		}
		obj.SetGroupVersionKind(schema.GroupVersionKind{Group: "", Version: "v1", Kind: "PersistentVolumeClaim"})
		normalizeObj(&obj)
		objs = append(objs, &obj)
	}
	pods, err := suite.k8sClient.CoreV1().Pods(namespace).List(suite.ctx, metav1.ListOptions{})
	suite.Require().NoError(err)
	for _, obj := range pods.Items {
		obj := obj
		obj.SetGroupVersionKind(schema.GroupVersionKind{Group: "", Version: "v1", Kind: "Pod"})
		normalizeObj(&obj)
		objs = append(objs, &obj)
	}
	roles, err := suite.k8sClient.RbacV1().Roles(namespace).List(suite.ctx, metav1.ListOptions{})
	suite.Require().NoError(err)
	for _, obj := range roles.Items {
		obj := obj
		obj.SetGroupVersionKind(schema.GroupVersionKind{Group: "rbac.authorization.k8s.io", Version: "v1", Kind: "Role"})
		normalizeObj(&obj)
		objs = append(objs, &obj)
	}
	roleBindings, err := suite.k8sClient.RbacV1().RoleBindings(namespace).List(suite.ctx, metav1.ListOptions{})
	suite.Require().NoError(err)
	for _, obj := range roleBindings.Items {
		obj := obj
		obj.SetGroupVersionKind(schema.GroupVersionKind{Group: "rbac.authorization.k8s.io", Version: "v1", Kind: "RoleBinding"})
		obj.Subjects = slices.Map(obj.Subjects, func(s rbacv1.Subject) rbacv1.Subject {
			s.Namespace = normalizedString
			return s
		})
		normalizeObj(&obj)
		objs = append(objs, &obj)
	}

	// These are just test secrets, nothing truly sensitive should end up in the
	// golden files.
	secrets, err := suite.k8sClient.CoreV1().Secrets(namespace).List(suite.ctx, metav1.ListOptions{})
	suite.Require().NoError(err)
	for _, obj := range secrets.Items {
		obj := obj
		obj.SetGroupVersionKind(schema.GroupVersionKind{Group: "", Version: "v1", Kind: "Secret"})
		normalizeObj(&obj)
		objs = append(objs, &obj)
	}

	sas, err := suite.k8sClient.CoreV1().ServiceAccounts(namespace).List(suite.ctx, metav1.ListOptions{})
	suite.Require().NoError(err)
	for _, obj := range sas.Items {
		obj := obj
		obj.SetGroupVersionKind(schema.GroupVersionKind{Group: "", Version: "v1", Kind: "ServiceAccount"})
		normalizeObj(&obj)
		objs = append(objs, &obj)
	}
	svcs, err := suite.k8sClient.CoreV1().Services(namespace).List(suite.ctx, metav1.ListOptions{})
	suite.Require().NoError(err)
	for _, obj := range svcs.Items {
		obj := obj
		obj.Spec.ClusterIP = normalizedString
		obj.Spec.ClusterIPs = []string{normalizedString}
		obj.SetGroupVersionKind(schema.GroupVersionKind{Group: "", Version: "v1", Kind: "Service"})
		normalizeObj(&obj)
		objs = append(objs, &obj)
	}

	ingresses, err := suite.k8sClient.NetworkingV1().Ingresses(namespace).List(suite.ctx, metav1.ListOptions{})
	suite.Require().NoError(err)
	for _, obj := range ingresses.Items {
		obj := obj
		obj.SetGroupVersionKind(schema.GroupVersionKind{Group: "networking.k8s.io", Version: "v1", Kind: "Ingress"})
		normalizeObj(&obj)
		objs = append(objs, &obj)
	}

	return objs
}

func normalizeObj(obj client.Object) {
	obj.SetUID(normalizedString)
	obj.SetCreationTimestamp(magicTime)
	obj.SetManagedFields(nil)
	obj.SetNamespace(normalizedString)
	obj.SetResourceVersion(normalizedString)

	ownerRefs := obj.GetOwnerReferences()
	normalizedOwnerRefs := make([]metav1.OwnerReference, len(ownerRefs))
	for i, ownerRef := range ownerRefs {
		ownerRef.UID = normalizedString
		normalizedOwnerRefs[i] = ownerRef
	}
	obj.SetOwnerReferences(normalizedOwnerRefs)
}
