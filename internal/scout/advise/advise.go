package advise

import "github.com/sourcegraph/src-cli/internal/scout"

type Option = func(config *scout.Config)

const (
	OVER_100 = "\t%s %s: Your %s usage is over 100%% (%.2f%%). Add more %s."
	OVER_80  = "\t%s %s: Your %s usage is over 80%% (%.2f%%). Consider raising limits."
	OVER_40  = "\t%s %s: Your %s usage is under 80%% (%.2f%%). Keep %s allocation the same."
	UNDER_40 = "\t%s %s: Your %s usage is under 40%% (%.2f%%). Consider lowering limits."
)

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

func WithContainer(containerName string) Option {
	return func(config *scout.Config) {
		config.Container = containerName
	}
}
