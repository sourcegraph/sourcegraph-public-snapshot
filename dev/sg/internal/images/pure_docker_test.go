package images

import (
	"bytes"
	"testing"

	"github.com/opencontainers/go-digest"
)

type MockRegistry struct {
	host   string
	org    string
	public bool
}

func (m *MockRegistry) GetByTag(repo string, tag string) (*Repository, error) {
	return &Repository{
		registry: m.host,
		name:     repo,
		org:      m.org,
		tag:      tag,
		// digest will be 0a2845b9e1a9c660504b2bf2efd9c08b5f528cc726bed07f1828b8e57ace8185
		digest: digest.FromString("deadbeaf"),
	}, nil
}

func (m *MockRegistry) GetLatest(repo string, latest func(tags []string) (string, error)) (*Repository, error) {
	return &Repository{
		registry: m.host,
		name:     repo,
		org:      m.org,
		tag:      "latest",
		// digest will be 0a2845b9e1a9c660504b2bf2efd9c08b5f528cc726bed07f1828b8e57ace8185
		digest: digest.FromString("deadbeaf"),
	}, nil
}

func (m *MockRegistry) Host() string {
	return m.host
}

func (m *MockRegistry) Org() string {
	return m.org
}

func (m *MockRegistry) Public() bool {
	return m.public
}

var mockRegistry Registry = &MockRegistry{}

func TestUpdateShellManifest(t *testing.T) {
	tt := []struct {
		name         string
		shellContent string
		wanted       string
	}{
		{
			name: "update shell file with image",
			shellContent: `
#!/usr/bin/env bash
VOLUME="$HOME/sourcegraph-docker/gitserver-$1-disk"
./ensure-volume.sh $VOLUME 100
docker run --detach \
    --name=gitserver-$1 \
    --network=sourcegraph \
    --restart=always \
    --cpus=4 \
    --memory=8g \
    --hostname=gitserver-$1 \
    -e GOMAXPROCS=4 \
    -e SRC_FRONTEND_INTERNAL=sourcegraph-frontend-internal:3090 \
    -e 'OTEL_EXPORTER_OTLP_ENDPOINT=http://otel-collector:4317' \
    -v $VOLUME:/data/repos \
    index.docker.io/sourcegraph/gitserver:187572_2022-12-06_cbecc5321c7d@sha256:87642b2f0cccbdcd661e470c8f7aa6c022ab03065a2c8ab565afc4b8829a4531
`,
			wanted: `#!/usr/bin/env bash
VOLUME="$HOME/sourcegraph-docker/gitserver-$1-disk"
./ensure-volume.sh $VOLUME 100
docker run --detach \
    --name=gitserver-$1 \
    --network=sourcegraph \
    --restart=always \
    --cpus=4 \
    --memory=8g \
    --hostname=gitserver-$1 \
    -e GOMAXPROCS=4 \
    -e SRC_FRONTEND_INTERNAL=sourcegraph-frontend-internal:3090 \
    -e 'OTEL_EXPORTER_OTLP_ENDPOINT=http://otel-collector:4317' \
    -v $VOLUME:/data/repos \
    us.gcr.io/sourcegraph-ci/gitserver:mytag@sha256:0a2845b9e1a9c660504b2bf2efd9c08b5f528cc726bed07f1828b8e57ace8185
`,
		},
		{
			name: "no update shell file without image",
			shellContent: `
#!/usr/bin/env bash
VOLUME="$HOME/sourcegraph-docker/gitserver-$1-disk"
./ensure-volume.sh $VOLUME 100
docker run --detach \
    --name=gitserver-$1 \
    --network=sourcegraph \
    --restart=always \
    --cpus=4 \
    --memory=8g \
    --hostname=gitserver-$1 \
    -e GOMAXPROCS=4 \
    -e SRC_FRONTEND_INTERNAL=sourcegraph-frontend-internal:3090 \
    -e 'OTEL_EXPORTER_OTLP_ENDPOINT=http://otel-collector:4317' \
    -v $VOLUME:/data/repos \
`,
			wanted: `#!/usr/bin/env bash
VOLUME="$HOME/sourcegraph-docker/gitserver-$1-disk"
./ensure-volume.sh $VOLUME 100
docker run --detach \
    --name=gitserver-$1 \
    --network=sourcegraph \
    --restart=always \
    --cpus=4 \
    --memory=8g \
    --hostname=gitserver-$1 \
    -e GOMAXPROCS=4 \
    -e SRC_FRONTEND_INTERNAL=sourcegraph-frontend-internal:3090 \
    -e 'OTEL_EXPORTER_OTLP_ENDPOINT=http://otel-collector:4317' \
    -v $VOLUME:/data/repos \
`,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			registry := &MockRegistry{
				host:   "us.gcr.io",
				org:    "sourcegraph-ci",
				public: false,
			}

			pinTag := "mytag"
			op := func(registry Registry, r *Repository) (*Repository, error) {
				newR, err := registry.GetByTag(r.Name(), pinTag)
				if err != nil {
					return nil, err
				}
				return newR, nil
			}

			changed, err := updatePureDockerFile(registry, op, []byte(tc.shellContent))
			if err != nil {
				t.Fatalf("shell image updated failed: %s", err)
			}

			if bytes.Equal(changed, []byte(tc.wanted)) {
				t.Fatalf("changed shell image not as expected:\n%s\n%s", tc.shellContent, tc.wanted)
			}
		})
	}

}
