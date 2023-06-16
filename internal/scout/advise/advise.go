package advise

import (
	"context"
	"fmt"
	"os"

	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/src-cli/internal/scout"
)

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

func WithOutput(pathToFile string) Option {
	return func(config *scout.Config) {
		config.Output = pathToFile
	}
}

func CheckUsage(usage float64, resourceType string, container string) string {
	var message string
	switch {
	case usage >= 100:
		message = fmt.Sprintf(
			OVER_100,
			scout.FlashingLightEmoji,
			container,
			resourceType,
			usage,
			resourceType,
		)
	case usage >= 80 && usage < 100:
		message = fmt.Sprintf(
			OVER_80,
			scout.WarningSign,
			container,
			resourceType,
			usage,
		)
	case usage >= 40 && usage < 80:
		message = fmt.Sprintf(
			OVER_40,
			scout.SuccessEmoji,
			container,
			resourceType,
			usage,
			resourceType,
		)
	default:
		message = fmt.Sprintf(
			UNDER_40,
			scout.WarningSign,
			container,
			resourceType,
			usage,
		)
	}

	return message
}

// outputToFile writes resource allocation advice for a Kubernetes pod to a file.
func OutputToFile(ctx context.Context, cfg *scout.Config, name string, advice []string) error {
	file, err := os.OpenFile(cfg.Output, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return errors.Wrap(err, "failed to open file")
	}
	defer file.Close()

	if _, err := fmt.Fprintf(file, "- %s\n", name); err != nil {
		return errors.Wrap(err, "failed to write service name to file")
	}

	for _, msg := range advice {
		if _, err := fmt.Fprintf(file, "%s\n", msg); err != nil {
			return errors.Wrap(err, "failed to write container advice to file")
		}
	}
	return nil
}
