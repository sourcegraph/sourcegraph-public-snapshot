package reconciler

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
	"time"

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

	"github.com/sourcegraph/sourcegraph/internal/appliance/config"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// Test helpers

type ApplianceTestSuite struct {
	suite.Suite

	ctx            context.Context
	k8sClient      *kubernetes.Clientset
	envtestCleanup func() error
}

func TestApplianceTestSuite(t *testing.T) {
	suite.Run(t, new(ApplianceTestSuite))
}

func (suite *ApplianceTestSuite) SetupSuite() {
	suite.ctx = context.Background()
	var err error
	suite.k8sClient, suite.envtestCleanup, err = setupEnvtest(suite.ctx)
	suite.Require().NoError(err)
}

func (suite *ApplianceTestSuite) TearDownSuite() {
	suite.Require().NoError(suite.envtestCleanup())
}

func setupEnvtest(ctx context.Context) (*kubernetes.Clientset, func() error, error) {
	ctx, cancel := context.WithCancel(ctx)

	logger := stdr.New(stdlog.New(os.Stderr, "", stdlog.LstdFlags))

	kubeBuilderAssets, err := kubebuilderAssetPath()
	if err != nil {
		cancel()
		return nil, nil, err
	}
	testEnv := &envtest.Environment{
		AttachControlPlaneOutput: true,
		BinaryAssetsDirectory:    kubeBuilderAssets,
	}
	apiServerCfg := testEnv.ControlPlane.GetAPIServer()
	apiServerCfg.Configure().Set("bind-address", "127.0.0.1")
	apiServerCfg.Configure().Set("advertise-address", "127.0.0.1")
	cfg, err := testEnv.Start()
	if err != nil {
		cancel()
		return nil, nil, err
	}

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
	if err != nil {
		cancel()
		return nil, nil, err
	}
	k8sClient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		cancel()
		return nil, nil, err
	}

	reconciler := &Reconciler{
		Client:   ctrlMgr.GetClient(),
		Scheme:   ctrlMgr.GetScheme(),
		Recorder: ctrlMgr.GetEventRecorderFor("appliance"),
	}

	// TODO invert this dependency
	if err := reconciler.SetupWithManager(ctrlMgr); err != nil {
		cancel()
		return nil, nil, err
	}

	// Start controller manager async. We'll stop it with context cancellation
	// later.
	ctrlMgrDone := make(chan error, 1)
	go func() {
		if err := ctrlMgr.Start(ctx); err != nil {
			ctrlMgrDone <- err
			return
		}
		close(ctrlMgrDone)
	}()

	cleanup := func() error {
		cancel()
		if err := testEnv.Stop(); err != nil {
			return err
		}
		return <-ctrlMgrDone
	}

	return k8sClient, cleanup, nil
}

func kubebuilderAssetPath() (string, error) {
	if os.Getenv("BAZEL_TEST") == "" {
		return kubebuilderAssetPathLocalDev()
	}

	assetPaths := strings.Split(os.Getenv("KUBEBUILDER_ASSET_PATHS"), " ")
	if len(assetPaths) == 0 {
		return "", errors.New("expected KUBEBUILDER_ASSET_PATHS to not be empty")
	}
	arbAssetPath, err := runfiles.Rlocation(assetPaths[0])
	if err != nil {
		return "", err
	}
	return filepath.Dir(arbAssetPath), nil
}

// In the hermetic bazel environment, we skip setup-envtest, handling the assets
// directly in a hermetic and cachable way.
// If we're using `go test`, which can be convenient for local dev, we fall back
// on expecting setup-envtest to be present on the developer machine.
func kubebuilderAssetPathLocalDev() (string, error) {
	setupEnvTestCmd := exec.Command("setup-envtest", "use", "1.28.0", "--bin-dir", "/tmp/envtest", "-p", "path")
	var envtestOut bytes.Buffer
	setupEnvTestCmd.Stdout = &envtestOut
	if err := setupEnvTestCmd.Run(); err != nil {
		return "", errors.Wrap(err, "error running setup-envtest - did you remember to `go install sigs.k8s.io/controller-runtime/tools/setup-envtest@latest`?")
	}
	return strings.TrimSpace(envtestOut.String()), nil
}

func (suite *ApplianceTestSuite) createConfigMapAndAwaitReconciliation(fixtureFileName string) string {
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
	suite.awaitReconciliation(namespace, func() {
		_, err := suite.k8sClient.CoreV1().ConfigMaps(namespace).Create(suite.ctx, cfgMap, metav1.CreateOptions{})
		suite.Require().NoError(err)
	})
	return namespace
}

func (suite *ApplianceTestSuite) updateConfigMapAndAwaitReconciliation(namespace, fixtureFileName string) {
	cfgMap := suite.newConfigMap(namespace, fixtureFileName)
	suite.awaitReconciliation(namespace, func() {
		_, err := suite.k8sClient.CoreV1().ConfigMaps(namespace).Update(suite.ctx, cfgMap, metav1.UpdateOptions{})
		suite.Require().NoError(err)
	})
}

// Synchronize test and controller code by counting ReconcileFinished events. We
// expect exactly 2 from one initial creation or update of an SG ConfigMap. This
// is because we update the ConfigMap at the end of the reconcile loop with
// annotations. This triggers another reconcile loop. This all ends when the
// changes are no-ops.
func (suite *ApplianceTestSuite) awaitReconciliation(namespace string, op func()) {
	events := suite.getConfigMapReconcileEventCount(namespace)
	op()
	suite.Require().Eventually(func() bool {
		return suite.getConfigMapReconcileEventCount(namespace) >= events+2
	}, time.Second*10, time.Millisecond*200)
}

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
				config.AnnotationKeyManaged: "true",
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
