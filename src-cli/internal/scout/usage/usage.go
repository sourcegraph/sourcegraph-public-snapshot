package usage

import (
	"github.com/sourcegraph/src-cli/internal/scout"
)

type Option = func(config *scout.Config)

func WithNamespace(namespace string) Option {
	return func(config *scout.Config) {
		config.Namespace = namespace
	}
}

func WithPod(podname string) Option {
	return func(config *scout.Config) {
		config.Pod = podname
	}
}
