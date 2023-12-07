package example

import (
	"testing"

	"github.com/hexops/autogold/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"

	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/spec"
)

func TestNewService(t *testing.T) {
	f, err := NewService(Template{
		ID:    "msp-example",
		Dev:   true,
		Owner: "core-services",
	})
	require.NoError(t, err)

	t.Run("is valid", func(t *testing.T) {
		var s spec.Spec
		require.NoError(t, yaml.Unmarshal(f, &s))
		assert.Empty(t, s.Validate())
	})

	autogold.Expect(`service:
  id: msp-example
  name: Msp Example
  owners:
  - core-services

build:
  # TODO: Configure the correct image for your service here. If you use a private
  # registry like us.gcr.io or Artifact Registry, access will automatically be
  # granted for your service to pull the correct image.
  image: us.gcr.io/sourcegraph-dev/msp-example
  # TODO: Configure where the source code for your service lives here.
  source:
    repo: github.com/sourcegraph/sourcegraph
    dir: cmd/msp-example

environments:
- id: dev
  # TODO: We initially provision in 'test' to make it easy to access the project
  # during setup. Once done, you should change this to 'external' or 'internal'.
  category: test
  # Specify a deployment strategy for upgrades.
  deploy:
    type: manual
    manual:
      tag: insiders
  # Specify an externally facing domain.
  domain:
    type: cloudflare
    cloudflare:
      subdomain: msp-example
      zone: sgdev.org
  # Specify environment configuration your service needs to operate.
  env:
    SRC_LOG_LEVEL: info
    SRC_LOG_FORMAT: json_gcp
  # Specify how your service should scale.
  instances:
    resources:
      cpu: 1
      memory: 1Gi
    scaling:
      maxCount: 3
      minCount: 1
  startupProbe:
    # Only enable if your service implements MSP /-/healthz conventions.
    disabled: true
`).Equal(t, string(f))
}

func TestNewJob(t *testing.T) {
	f, err := NewJob(Template{
		ID:    "msp-example",
		Dev:   true,
		Owner: "core-services",
	})
	require.NoError(t, err)

	t.Run("is valid", func(t *testing.T) {
		var s spec.Spec
		require.NoError(t, yaml.Unmarshal(f, &s))
		assert.Empty(t, s.Validate())
	})

	autogold.Expect(`service:
  kind: job
  id: msp-example
  name: Msp Example
  owners:
  - core-services

build:
  # TODO: Configure the correct image for your job here. If you use a private
  # registry like us.gcr.io or Artifact Registry, access will automatically be
  # granted for your job to pull the correct image.
  image: us.gcr.io/sourcegraph-dev/msp-example
  # TODO: Configure where the source code for your job lives here.
  source:
    repo: github.com/sourcegraph/sourcegraph
    dir: cmd/msp-example

environments:
- id: dev
  # TODO: We initially provision in 'test' to make it easy to access the project
  # during setup. Once done, you should change this to 'external' or 'internal'.
  category: test
  # Specify a strategy for updating the image.
  deploy:
    type: manual
    manual:
      tag: insiders
  # Specify the schedule at which to run your job.
  schedule:
    cron: 0 * * * *
    deadline: 600 # 10 minutes
  # Specify environment configuration your service needs to operate.
  env:
    SRC_LOG_LEVEL: info
    SRC_LOG_FORMAT: json_gcp
  # Specify the resources your job gets.
  instances:
    resources:
      cpu: 1
      memory: 1Gi
`).Equal(t, string(f))
}
