package shared

import (
	"path/filepath"

	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"sigs.k8s.io/controller-runtime/pkg/cache"

	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type Config struct {
	env.BaseConfig

	k8sConfig *rest.Config
	metrics   metricsConfig
	grpc      grpcConfig
	namespace string
}

func (c *Config) Load() {
	var kubeConfig string
	k8sConfig, err := rest.InClusterConfig()
	if err != nil {
		if home := homedir.HomeDir(); home != "" {
			kubeConfig = c.Get("KUBECONFIG", filepath.Join(home, ".kube", "config"), "Absolute path to the kubeconfig file.")
		}

		k8sConfig, err = clientcmd.BuildConfigFromFlags("", kubeConfig)
		if err != nil {
			c.AddError(errors.New("could not create kubernetes client config"))
		}
	}

	c.k8sConfig = k8sConfig
	c.metrics.addr = c.Get("APPLIANCE_METRICS_ADDR", ":8080", "Appliance metrics server address.")
	c.metrics.secure = c.GetBool("APPLIANCE_METRICS_SECURE", "false", "Appliance metrics server uses https.")
	c.grpc.addr = c.Get("APPLIANCE_GRPC_ADDR", ":9000", "Appliance gRPC address.")
	c.namespace = c.Get("APPLIANCE_NAMESPACE", cache.AllNamespaces, "Namespace to monitor. Defaults to all.")
}

func (c *Config) Validate() error {
	var errs error
	return errs
}

type metricsConfig struct {
	addr   string
	secure bool
}

type grpcConfig struct {
	addr string
}
