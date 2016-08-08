package localstore

import (
	"errors"
	"os"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/conf/feature"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/langp"
)

func langpClient() (*langp.Client, error) {
	if !feature.Features.Universe {
		return nil, errors.New("Universe feature is not enabled")
	}
	return langp.NewClient(os.Getenv("SG_LANGUAGE_PROCESSOR"))
}
