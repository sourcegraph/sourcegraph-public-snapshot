package conf

import (
	"strings"
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/conf/confdefaults"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
)

func TestMustValidateDefaults(t *testing.T) {
	t.Run("DevAndTesting", func(t *testing.T) {
		mustValidate(t, confdefaults.DevAndTesting)
	})
	t.Run("DockerContainer", func(t *testing.T) {
		mustValidate(t, confdefaults.DockerContainer)
	})
	t.Run("KubernetesOrDockerComposeOrPureDocker", func(t *testing.T) {
		mustValidate(t, confdefaults.KubernetesOrDockerComposeOrPureDocker)
	})
	t.Run("App", func(t *testing.T) {
		mustValidate(t, confdefaults.App)
	})
}

// mustValidate makes sure the given configuration passes validation.
func mustValidate(t *testing.T, cfg conftypes.RawUnified) {
	t.Helper()

	problems, err := Validate(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if len(problems) > 0 {
		t.Fatalf("conf: problems with default configuration for:\n  %s", strings.Join(problems.Messages(), "\n  "))
	}
}
