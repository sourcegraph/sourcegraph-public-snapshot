package selfupdate

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/sourcegraph/log/logr"
	"github.com/sourcegraph/log/logtest"
	"github.com/sourcegraph/sourcegraph/internal/appliance/config"
	"github.com/sourcegraph/sourcegraph/internal/appliance/k8senvtest"
	"github.com/sourcegraph/sourcegraph/internal/k8s/resource/container"
	"github.com/sourcegraph/sourcegraph/internal/k8s/resource/deployment"
	"github.com/sourcegraph/sourcegraph/internal/k8s/resource/pod"
	"github.com/sourcegraph/sourcegraph/internal/releaseregistry"
)

func TestReplaceTag(t *testing.T) {
	img := "index.docker.io/sourcegraph/appliance:1.2.3"
	updated := replaceTag(img, "4.5.6")
	require.Equal(t, "index.docker.io/sourcegraph/appliance:4.5.6", updated)
}

func TestReplaceTagNeverPanics(t *testing.T) {
	img := "badImageNameFormat"
	updated := replaceTag(img, "4.5.6")
	require.Equal(t, ":4.5.6", updated)
}

func TestGetLatestTag_ReturnsLatestSupportedPublicVersion(t *testing.T) {
	relregClient := NewMockReleaseRegistryClient()
	selfUpdater := &SelfUpdate{
		Logger:       logtest.Scoped(t),
		RelregClient: relregClient,
	}
	relregClient.ListVersionsFunc.PushReturn([]releaseregistry.ReleaseInfo{
		{Version: "v4.5.6", Public: false},
		{Version: "v4.5.5", Public: true},
		{Version: "v4.5.4", Public: true},
		{Version: "v4.5.3", Public: false},
		{Version: "v3.17.1", Public: true},
	}, nil)

	latest, err := selfUpdater.getLatestTag(context.Background(), "4.3.0")
	require.NoError(t, err)
	require.Equal(t, "4.5.5", latest)
}

func TestSelfUpdateLoop(t *testing.T) {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	logger := logtest.Scoped(t)
	k8sConfig, cleanup, err := k8senvtest.SetupEnvtest(ctx, logr.New(logger), k8senvtest.NewNoopReconciler)
	require.NoError(t, err)
	defer func() {
		require.NoError(t, cleanup())
	}()

	k8sClient, err := client.New(k8sConfig, client.Options{})
	require.NoError(t, err)

	// provision example appliance deployment
	dep := buildTestDeployment()
	err = k8sClient.Create(ctx, dep)
	require.NoError(t, err)

	cfgMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      config.ConfigmapName,
			Namespace: "default",
			Annotations: map[string]string{
				config.AnnotationKeyCurrentVersion: "4.3.0",
			},
		},
	}
	err = k8sClient.Create(ctx, cfgMap)
	require.NoError(t, err)

	relregClient := NewMockReleaseRegistryClient()
	relregClient.ListVersionsFunc.SetDefaultReturn([]releaseregistry.ReleaseInfo{
		{Version: "4.5.7", Public: true},
	}, nil)
	selfUpdater := &SelfUpdate{
		Interval:       time.Second,
		Logger:         logger,
		RelregClient:   relregClient,
		K8sClient:      k8sClient,
		DeploymentName: "appliance",
		Namespace:      "default",
	}

	loopDone := make(chan struct{})
	go func() {
		_ = selfUpdater.Loop(ctx)
		close(loopDone)
	}()

	require.Eventually(t, func() bool {
		var dep appsv1.Deployment
		name := types.NamespacedName{Name: "appliance", Namespace: "default"}
		require.NoError(t, k8sClient.Get(ctx, name, &dep))
		return strings.HasSuffix(dep.Spec.Template.Spec.Containers[0].Image, "4.5.7")
	}, time.Second*5, time.Second)

	cancel()
	<-loopDone
}

func buildTestDeployment() *appsv1.Deployment {
	name := "appliance"
	defaultContainer := container.NewContainer(name, nil, config.ContainerConfig{
		Image:     "index.docker.io/sourcegraph/appliance:5.4.7765",
		Resources: &corev1.ResourceRequirements{},
	})
	podTemplate := pod.NewPodTemplate(name, nil)
	podTemplate.Template.Spec.Containers = []corev1.Container{defaultContainer}
	defaultDeployment := deployment.NewDeployment(name, "default", "5.4.7765")
	defaultDeployment.Spec.Template = podTemplate.Template
	return &defaultDeployment
}
