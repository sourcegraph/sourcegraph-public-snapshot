package integrationtest

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"sigs.k8s.io/yaml"

	"github.com/sourcegraph/sourcegraph/internal/appliance/config"
	"github.com/sourcegraph/sourcegraph/internal/k8s/resource/configmap"
)

// Note - this whole package is early WIP. These tests are skipped in CI,
// because nothing sets APPLIANCE_SMOKE_TESTS_RUN. See
// https://linear.app/sourcegraph/issue/REL-23/developer-smoke-testing-setup-with-a-real-kubernetes
// and any follow-on work for plans in this area.

// This test assumes that the appliance/reconciler is already running in the
// background.
func TestSmokeTestApplianceInstallation(t *testing.T) {
	if os.Getenv("APPLIANCE_SMOKE_TESTS_RUN") == "" {
		t.Skip("APPLIANCE_SMOKE_TESTS_RUN is not set, skipping")
	}

	ctx := context.Background()
	k8sClient := k8sClient(t)
	namespace, cleanup := createTestNamespace(t, ctx, k8sClient)
	if os.Getenv("APPLIANCE_SMOKE_TESTS_SKIP_TEARDOWN") == "" {
		defer cleanup()
	}

	defaultConfig := config.SourcegraphSpec{
		RequestedVersion: "5.3.9104",
	}
	createConfigMap(t, namespace, k8sClient, defaultConfig)

	// TODO wait for services to be up
	//
	// TODO src-validate exits with zero under some scenerios, including when
	// there are no pods at all:
	// https://linear.app/sourcegraph/issue/REL-42/src-validate-should-not-exit-with-zero-when-no-pods-are-available-at
	//
	// TODO add black-box validations to src-validate that exercise core flows,
	// not just check processes are running and sockets are listening. Perhaps
	// put this behind a flag:
	// https://linear.app/sourcegraph/issue/REL-43/src-validate-should-perform-black-box-validations-of-core-flows
	srcValidate(t, namespace)
}

func srcValidate(t *testing.T, namespace string) {
	cmd := exec.Command("src", "validate", "kube", "--namespace", namespace)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	require.NoError(t, cmd.Run())
}

func createConfigMap(t *testing.T, namespace string, k8sClient *kubernetes.Clientset, cfg config.SourcegraphSpec) {
	sg := &config.Sourcegraph{
		Spec: cfg,
	}
	sg.SetLocalDevMode()
	cfgBytes, err := yaml.Marshal(sg)
	require.NoError(t, err)
	cfgMap := configmap.NewConfigMap("sg", namespace)
	cfgMap.Data = map[string]string{
		"spec": string(cfgBytes),
	}
	cfgMap.SetAnnotations(map[string]string{
		"appliance.sourcegraph.com/managed": "true",
	})
	_, err = k8sClient.CoreV1().ConfigMaps(namespace).Create(context.Background(), &cfgMap, metav1.CreateOptions{})
	require.NoError(t, err)
}

func createTestNamespace(t *testing.T, ctx context.Context, k8sClient *kubernetes.Clientset) (string, func()) {
	namespaceName := "appliance-integration-test-" + randomSlug(t)
	namespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespaceName,
		},
	}
	_, err := k8sClient.CoreV1().Namespaces().Create(ctx, namespace, metav1.CreateOptions{})
	require.NoError(t, err)

	return namespaceName, func() {
		err := k8sClient.CoreV1().Namespaces().Delete(ctx, namespaceName, metav1.DeleteOptions{})
		require.NoError(t, err)
	}
}

func k8sClient(t *testing.T) *kubernetes.Clientset {
	kubeConfigPath := os.Getenv("KUBECONFIG")
	if kubeConfigPath == "" {
		kubeConfigPath = filepath.Join(homedir.HomeDir(), ".kube", "config")
	}

	k8sCfg, err := clientcmd.BuildConfigFromFlags("", kubeConfigPath)
	require.NoError(t, err)
	return kubernetes.NewForConfigOrDie(k8sCfg)
}

func randomSlug(t *testing.T) string {
	buf := make([]byte, 3)
	_, err := rand.Read(buf)
	require.NoError(t, err)
	return hex.EncodeToString(buf)
}
