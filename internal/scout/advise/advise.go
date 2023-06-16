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
	UNDER_PROVISIONED = "%s %s: %s is under-provisioned (%.2f%% usage). Add resources."
	WELL_PROVISIONED  = "%s %s: %s is well-provisioned (%.2f%% usage). No action needed."
	OVER_PROVISIONED  = "%s %s: %s is over-provisioned (%.2f%% usage). Trim resources."
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

func WithWarnings(includeWarnings bool) Option {
	return func(config *scout.Config) {
		config.Warnings = includeWarnings
	}
}

func CheckUsage(usage float64, resourceType string, container string) scout.Advice {
	var advice scout.Advice
	switch {
	case usage >= 80:
		advice.Kind = scout.DANGER
		advice.Msg = fmt.Sprintf(
			UNDER_PROVISIONED,
			scout.FlashingLightEmoji,
			container,
			resourceType,
			usage,
		)
	case usage >= 20 && usage < 80:
		advice.Kind = scout.HEALTHY
		advice.Msg = fmt.Sprintf(
			WELL_PROVISIONED,
			scout.SuccessEmoji,
			container,
			resourceType,
			usage,
		)
	default:
		advice.Kind = scout.WARNING
		advice.Msg = fmt.Sprintf(
			OVER_PROVISIONED,
			scout.WarningSign,
			container,
			resourceType,
			usage,
		)
	}

	return advice
}

// outputToFile writes resource allocation advice for a Kubernetes pod to a file.
func OutputToFile(ctx context.Context, cfg *scout.Config, name string, advice []scout.Advice) error {
	file, err := os.OpenFile(cfg.Output, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return errors.Wrap(err, "failed to open file")
	}
	defer file.Close()

	if _, err := fmt.Fprintf(file, "- %s\n", name); err != nil {
		return errors.Wrap(err, "failed to write service name to file")
	}

	for _, adv := range advice {
		if _, err := fmt.Fprintf(file, "%s\n", adv.Msg); err != nil {
			return errors.Wrap(err, "failed to write container advice to file")
		}
	}
	return nil
}
