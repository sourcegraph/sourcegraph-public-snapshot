package appliance

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	stdlog "log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bazelbuild/rules_go/go/runfiles"
	"github.com/go-logr/stdr"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
)

// Test helpers

type ApplianceTestSuite struct {
	suite.Suite

	ctx       context.Context
	cancelCtx context.CancelFunc

	testEnv     *envtest.Environment
	ctrlMgrDone chan struct{}

	k8sClient *kubernetes.Clientset
}

func TestApplianceTestSuite(t *testing.T) {
	suite.Run(t, new(ApplianceTestSuite))
}

func (suite *ApplianceTestSuite) SetupSuite() {
	suite.ctx, suite.cancelCtx = context.WithCancel(context.Background())
	suite.setupEnvtest()
}

func (suite *ApplianceTestSuite) TearDownSuite() {
	t := suite.T()
	suite.cancelCtx()
	require.NoError(t, suite.testEnv.Stop())
	<-suite.ctrlMgrDone
}

func (suite *ApplianceTestSuite) setupEnvtest() {
	t := suite.T()
	logger := stdr.New(stdlog.New(os.Stderr, "", stdlog.LstdFlags))

	suite.testEnv = &envtest.Environment{
		AttachControlPlaneOutput: true,
		BinaryAssetsDirectory:    suite.kubebuilderAssetPath(),
	}
	apiServerCfg := suite.testEnv.ControlPlane.GetAPIServer()
	apiServerCfg.Configure().Set("bind-address", "127.0.0.1")
	apiServerCfg.Configure().Set("advertise-address", "127.0.0.1")
	cfg, err := suite.testEnv.Start()
	require.NoError(t, err)

	// If we had CRDs, this is where we'd add them to the scheme

	ctrlMgr, err := ctrl.NewManager(cfg, ctrl.Options{
		Logger: logger,
		Scheme: scheme.Scheme,

		// On macos with the built-in firewall enabled, every test run will
		// cause a firewall dialog box to pop. That's incredibly annoying. This
		// can be solved by binding the metrics server to localhost, but we
		// don't need it, so we disable it altogether.
		Metrics: metricsserver.Options{BindAddress: "0"},
	})
	require.NoError(t, err)
	suite.k8sClient, err = kubernetes.NewForConfig(cfg)
	require.NoError(t, err)

	reconciler := &Reconciler{
		Client:   ctrlMgr.GetClient(),
		Scheme:   ctrlMgr.GetScheme(),
		Recorder: ctrlMgr.GetEventRecorderFor("appliance"),
	}
	require.NoError(t, reconciler.SetupWithManager(ctrlMgr))

	// Start controller manager async. We'll stop it with context cancellation
	// later.
	suite.ctrlMgrDone = make(chan struct{})
	go func() {
		require.NoError(t, ctrlMgr.Start(suite.ctx))
		close(suite.ctrlMgrDone)
	}()
}

func (suite *ApplianceTestSuite) kubebuilderAssetPath() string {
	if os.Getenv("BAZEL_TEST") == "" {
		return suite.kubebuilderAssetPathLocalDev()
	}

	assetPaths := strings.Split(os.Getenv("KUBEBUILDER_ASSET_PATHS"), " ")
	suite.Require().Greater(len(assetPaths), 0)
	arbAssetPath, err := runfiles.Rlocation(assetPaths[0])
	suite.Require().NoError(err)
	return filepath.Dir(arbAssetPath)
}

// In the hermetic bazel environment, we skip setup-envtest, handling the assets
// directly in a hermetic and cachable way.
// If we're using `go test`, which can be convenient for local dev, we fall back
// on expecting setup-envtest to be present on the developer machine.
func (suite *ApplianceTestSuite) kubebuilderAssetPathLocalDev() string {
	setupEnvTestCmd := exec.Command("setup-envtest", "use", "1.28.0", "--bin-dir", "/tmp/envtest", "-p", "path")
	var envtestOut bytes.Buffer
	setupEnvTestCmd.Stdout = &envtestOut
	err := setupEnvTestCmd.Run()
	suite.Require().NoError(err, "Did you remember to `go install sigs.k8s.io/controller-runtime/tools/setup-envtest@latest`?")
	return strings.TrimSpace(envtestOut.String())
}

func (suite *ApplianceTestSuite) createConfigMap(fixtureFileName string) string {
	// Create a random namespace for each test
	namespace := "test-appliance-" + suite.randomSlug()
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespace,
		},
	}
	_, err := suite.k8sClient.CoreV1().Namespaces().Create(suite.ctx, ns, metav1.CreateOptions{})
	suite.Require().NoError(err)

	cfgMap := suite.newConfigMap(namespace, fixtureFileName)
	_, err = suite.k8sClient.CoreV1().ConfigMaps(namespace).Create(suite.ctx, cfgMap, metav1.CreateOptions{})
	suite.Require().NoError(err)
	return namespace
}

func (suite *ApplianceTestSuite) updateConfigMap(namespace, fixtureFileName string) {
	cfgMap := suite.newConfigMap(namespace, fixtureFileName)
	_, err := suite.k8sClient.CoreV1().ConfigMaps(namespace).Update(suite.ctx, cfgMap, metav1.UpdateOptions{})
	suite.Require().NoError(err)
}

// Synchronize test and controller code by counting ReconcileFinished events.
// Some tests might want to wait for more than 1 to appear.
func (suite *ApplianceTestSuite) getConfigMapReconcileEventCount(namespace string) int32 {
	t := suite.T()
	events, err := suite.k8sClient.CoreV1().Events(namespace).List(suite.ctx, metav1.ListOptions{
		FieldSelector: "involvedObject.name=sg,reason=ReconcileFinished",
		TypeMeta:      metav1.TypeMeta{Kind: "ConfigMap"},
	})
	require.NoError(t, err)
	if len(events.Items) == 0 {
		return 0
	}
	event := events.Items[0]
	return event.Count
}

func (suite *ApplianceTestSuite) newConfigMap(namespace, fixtureFileName string) *corev1.ConfigMap {
	t := suite.T()
	cfgBytes, err := os.ReadFile(filepath.Join("testdata", "sg", fixtureFileName+".yaml"))
	require.NoError(t, err)

	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "sg",
			Namespace: namespace,
			Annotations: map[string]string{
				annotationKeyManaged: "true",
			},
		},
		Data: map[string]string{"spec": string(cfgBytes)},
	}
}

func (suite *ApplianceTestSuite) randomSlug() string {
	buf := make([]byte, 3)
	_, err := rand.Read(buf)
	suite.Require().NoError(err)
	return hex.EncodeToString(buf)
}
