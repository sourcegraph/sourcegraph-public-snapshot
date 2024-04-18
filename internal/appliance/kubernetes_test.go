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
	ctrlMgrDone chan (struct{})

	k8sClient *kubernetes.Clientset
}

func TestApplianceTestSuite(t *testing.T) {
	suite.Run(t, new(ApplianceTestSuite))
}

func (s *ApplianceTestSuite) SetupSuite() {
	s.ctx, s.cancelCtx = context.WithCancel(context.Background())
	s.setupEnvtest()
}

func (s *ApplianceTestSuite) TearDownSuite() {
	t := s.T()
	s.cancelCtx()
	require.NoError(t, s.testEnv.Stop())
	<-s.ctrlMgrDone
}

func (s *ApplianceTestSuite) setupEnvtest() {
	t := s.T()
	logger := stdr.New(stdlog.New(os.Stderr, "", stdlog.LstdFlags))

	// TODO figure out a way to make this work with bazel. Either download the
	// etcd/k8s executables that setup-envtest downloads and pass that directory
	// into BinaryAssetsDirectory below, or download setup-envtest without
	// assuming it's already on the PATH.
	// Ideally these assets could be cached in buildkite too.
	setupEnvTestCmd := exec.Command("setup-envtest", "use", "1.28.0", "--bin-dir", "/tmp/envtest", "-p", "path")
	var envtestOut bytes.Buffer
	setupEnvTestCmd.Stdout = &envtestOut
	err := setupEnvTestCmd.Run()
	require.NoError(t, err)

	s.testEnv = &envtest.Environment{
		BinaryAssetsDirectory: strings.TrimSpace(envtestOut.String()),
	}
	cfg, err := s.testEnv.Start()
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
	s.k8sClient, err = kubernetes.NewForConfig(cfg)
	require.NoError(t, err)

	reconciler := &Reconciler{
		Client:   ctrlMgr.GetClient(),
		Scheme:   ctrlMgr.GetScheme(),
		Recorder: ctrlMgr.GetEventRecorderFor("appliance"),
	}
	require.NoError(t, reconciler.SetupWithManager(ctrlMgr))

	// Start controller manager async. We'll stop it with context cancellation
	// later.
	s.ctrlMgrDone = make(chan struct{})
	go func() {
		require.NoError(t, ctrlMgr.Start(s.ctx))
		close(s.ctrlMgrDone)
	}()
}

func (s *ApplianceTestSuite) createConfigMap(fixtureFileName string) string {
	// Create a random namespace for each test
	namespace := "test-appliance-" + s.randomSlug()
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespace,
		},
	}
	_, err := s.k8sClient.CoreV1().Namespaces().Create(s.ctx, ns, metav1.CreateOptions{})
	s.Require().NoError(err)

	cfgMap := s.newConfigMap(namespace, fixtureFileName)
	_, err = s.k8sClient.CoreV1().ConfigMaps(namespace).Create(s.ctx, cfgMap, metav1.CreateOptions{})
	s.Require().NoError(err)
	return namespace
}

func (s *ApplianceTestSuite) updateConfigMap(namespace, fixtureFileName string) {
	cfgMap := s.newConfigMap(namespace, fixtureFileName)
	_, err := s.k8sClient.CoreV1().ConfigMaps(namespace).Update(s.ctx, cfgMap, metav1.UpdateOptions{})
	s.Require().NoError(err)
}

// Synchronize test and controller code by counting ReconcileFInished events.
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

func (s *ApplianceTestSuite) newConfigMap(namespace, fixtureFileName string) *corev1.ConfigMap {
	t := s.T()
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

func (s *ApplianceTestSuite) randomSlug() string {
	buf := make([]byte, 3)
	_, err := rand.Read(buf)
	s.Require().NoError(err)
	return hex.EncodeToString(buf)
}
