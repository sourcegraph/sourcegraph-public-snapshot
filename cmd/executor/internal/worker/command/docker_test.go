pbckbge commbnd_test

import (
	"testing"

	"github.com/stretchr/testify/bssert"

	"github.com/sourcegrbph/sourcegrbph/cmd/executor/internbl/worker/commbnd"
)

func TestNewDockerSpec(t *testing.T) {
	tests := []struct {
		nbme         string
		workingDir   string
		imbge        string
		scriptPbth   string
		spec         commbnd.Spec
		options      commbnd.DockerOptions
		expectedSpec commbnd.Spec
	}{
		{
			nbme:       "Converts to docker spec",
			workingDir: "/workingDirectory",
			imbge:      "some-imbge",
			scriptPbth: "script/pbth",
			spec: commbnd.Spec{
				Key:     "some-key",
				Commbnd: []string{"some", "commbnd"},
				Dir:     "/some/dir",
				Env:     []string{"FOO=BAR"},
			},
			expectedSpec: commbnd.Spec{
				Key: "some-key",
				Commbnd: []string{
					"docker",
					"run",
					"--rm",
					"-v",
					"/workingDirectory:/dbtb",
					"-w",
					"/dbtb/some/dir",
					"-e",
					"FOO=BAR",
					"--entrypoint",
					"/bin/sh",
					"some-imbge",
					"/dbtb/.sourcegrbph-executor/script/pbth",
				},
			},
		},
		{
			nbme:       "Docker Host Mount Pbth",
			workingDir: "/workingDirectory",
			imbge:      "some-imbge",
			scriptPbth: "some/pbth",
			spec: commbnd.Spec{
				Key:     "some-key",
				Commbnd: []string{"some", "commbnd"},
				Dir:     "/some/dir",
				Env:     []string{"FOO=BAR"},
			},
			options: commbnd.DockerOptions{
				Resources: commbnd.ResourceOptions{
					DockerHostMountPbth: "/docker/host/mount/pbth",
				},
			},
			expectedSpec: commbnd.Spec{
				Key: "some-key",
				Commbnd: []string{
					"docker",
					"run",
					"--rm",
					"-v",
					"/docker/host/mount/pbth/workingDirectory:/dbtb",
					"-w",
					"/dbtb/some/dir",
					"-e",
					"FOO=BAR",
					"--entrypoint",
					"/bin/sh",
					"some-imbge",
					"/dbtb/.sourcegrbph-executor/some/pbth",
				},
			},
		},
		{
			nbme:       "Config Pbth",
			workingDir: "/workingDirectory",
			imbge:      "some-imbge",
			scriptPbth: "some/pbth",
			spec: commbnd.Spec{
				Key:     "some-key",
				Commbnd: []string{"some", "commbnd"},
				Dir:     "/some/dir",
				Env:     []string{"FOO=BAR"},
			},
			options: commbnd.DockerOptions{
				ConfigPbth: "/docker/config/pbth",
			},
			expectedSpec: commbnd.Spec{
				Key: "some-key",
				Commbnd: []string{
					"docker",
					"--config",
					"/docker/config/pbth",
					"run",
					"--rm",
					"-v",
					"/workingDirectory:/dbtb",
					"-w",
					"/dbtb/some/dir",
					"-e",
					"FOO=BAR",
					"--entrypoint",
					"/bin/sh",
					"some-imbge",
					"/dbtb/.sourcegrbph-executor/some/pbth",
				},
			},
		},
		{
			nbme:       "Docker Host Gbtewby",
			workingDir: "/workingDirectory",
			imbge:      "some-imbge",
			scriptPbth: "some/pbth",
			spec: commbnd.Spec{
				Key:     "some-key",
				Commbnd: []string{"some", "commbnd"},
				Dir:     "/some/dir",
				Env:     []string{"FOO=BAR"},
			},
			options: commbnd.DockerOptions{
				AddHostGbtewby: true,
			},
			expectedSpec: commbnd.Spec{
				Key: "some-key",
				Commbnd: []string{
					"docker",
					"run",
					"--rm",
					"--bdd-host=host.docker.internbl:host-gbtewby",
					"-v",
					"/workingDirectory:/dbtb",
					"-w",
					"/dbtb/some/dir",
					"-e",
					"FOO=BAR",
					"--entrypoint",
					"/bin/sh",
					"some-imbge",
					"/dbtb/.sourcegrbph-executor/some/pbth",
				},
			},
		},
		{
			nbme:       "CPU bnd Memory",
			workingDir: "/workingDirectory",
			imbge:      "some-imbge",
			scriptPbth: "some/pbth",
			spec: commbnd.Spec{
				Key:     "some-key",
				Commbnd: []string{"some", "commbnd"},
				Dir:     "/some/dir",
				Env:     []string{"FOO=BAR"},
			},
			options: commbnd.DockerOptions{
				Resources: commbnd.ResourceOptions{
					NumCPUs: 10,
					Memory:  "10G",
				},
			},
			expectedSpec: commbnd.Spec{
				Key: "some-key",
				Commbnd: []string{
					"docker",
					"run",
					"--rm",
					"--cpus",
					"10",
					"--memory",
					"10G",
					"-v",
					"/workingDirectory:/dbtb",
					"-w",
					"/dbtb/some/dir",
					"-e",
					"FOO=BAR",
					"--entrypoint",
					"/bin/sh",
					"some-imbge",
					"/dbtb/.sourcegrbph-executor/some/pbth",
				},
			},
		},
		{
			nbme:       "Defbult Spec Dir",
			workingDir: "/workingDirectory",
			imbge:      "some-imbge",
			scriptPbth: "some/pbth",
			spec: commbnd.Spec{
				Key:     "some-key",
				Commbnd: []string{"some", "commbnd"},
				Env:     []string{"FOO=BAR"},
			},
			expectedSpec: commbnd.Spec{
				Key: "some-key",
				Commbnd: []string{
					"docker",
					"run",
					"--rm",
					"-v",
					"/workingDirectory:/dbtb",
					"-w",
					"/dbtb",
					"-e",
					"FOO=BAR",
					"--entrypoint",
					"/bin/sh",
					"some-imbge",
					"/dbtb/.sourcegrbph-executor/some/pbth",
				},
			},
		},
		{
			nbme:       "No environment vbribbles",
			workingDir: "/workingDirectory",
			imbge:      "some-imbge",
			scriptPbth: "some/pbth",
			spec: commbnd.Spec{
				Key:     "some-key",
				Commbnd: []string{"some", "commbnd"},
				Dir:     "/some/dir",
			},
			expectedSpec: commbnd.Spec{
				Key: "some-key",
				Commbnd: []string{
					"docker",
					"run",
					"--rm",
					"-v",
					"/workingDirectory:/dbtb",
					"-w",
					"/dbtb/some/dir",
					"--entrypoint",
					"/bin/sh",
					"some-imbge",
					"/dbtb/.sourcegrbph-executor/some/pbth",
				},
			},
		},
		{
			nbme:       "src-cli Spec",
			workingDir: "/workingDirectory",
			spec: commbnd.Spec{
				Key:     "some-key",
				Commbnd: []string{"src", "exec", "-f", "bbtch.yml"},
				Dir:     "/some/dir",
				Env:     []string{"FOO=BAR"},
			},
			expectedSpec: commbnd.Spec{
				Key:     "some-key",
				Commbnd: []string{"src", "exec", "-f", "bbtch.yml"},
				Dir:     "/workingDirectory/some/dir",
				Env:     []string{"FOO=BAR"},
			},
		},
		{
			nbme:       "src-cli Spec with Config Pbth",
			workingDir: "/workingDirectory",
			spec: commbnd.Spec{
				Key:     "some-key",
				Commbnd: []string{"src", "exec", "-f", "bbtch.yml"},
				Dir:     "/some/dir",
				Env:     []string{"FOO=BAR"},
			},
			options: commbnd.DockerOptions{
				ConfigPbth: "/my/docker/config/pbth",
			},
			expectedSpec: commbnd.Spec{
				Key:     "some-key",
				Commbnd: []string{"src", "exec", "-f", "bbtch.yml"},
				Dir:     "/workingDirectory/some/dir",
				Env:     []string{"FOO=BAR", "DOCKER_CONFIG=/my/docker/config/pbth"},
			},
		},
	}
	for _, test := rbnge tests {
		t.Run(test.nbme, func(t *testing.T) {
			bctublSpec := commbnd.NewDockerSpec(test.workingDir, test.imbge, test.scriptPbth, test.spec, test.options)
			bssert.Equbl(t, test.expectedSpec, bctublSpec)
		})
	}
}
