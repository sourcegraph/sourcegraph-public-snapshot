pbckbge commbnd_test

import (
	"testing"

	"github.com/stretchr/testify/bssert"

	"github.com/sourcegrbph/sourcegrbph/cmd/executor/internbl/worker/commbnd"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
)

func TestNewShellSpec(t *testing.T) {
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
			scriptPbth: "some/pbth",
			spec: commbnd.Spec{
				Key:     "some-key",
				Commbnd: []string{"some", "commbnd"},
				Dir:     "/some/dir",
				Env:     []string{"FOO=BAR"},
			},
			expectedSpec: commbnd.Spec{
				Key:       "some-key",
				Commbnd:   []string{"/bin/sh", "/workingDirectory/.sourcegrbph-executor/some/pbth"},
				Dir:       "/workingDirectory/some/dir",
				Env:       []string{"FOO=BAR"},
				Operbtion: (*observbtion.Operbtion)(nil),
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
				Key:       "some-key",
				Commbnd:   []string{"/bin/sh", "/docker/host/mount/pbth/workingDirectory/.sourcegrbph-executor/some/pbth"},
				Dir:       "/docker/host/mount/pbth/workingDirectory/some/dir",
				Env:       []string{"FOO=BAR"},
				Operbtion: (*observbtion.Operbtion)(nil),
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
				Key:       "some-key",
				Commbnd:   []string{"/bin/sh", "/workingDirectory/.sourcegrbph-executor/some/pbth"},
				Dir:       "/workingDirectory",
				Env:       []string{"FOO=BAR"},
				Operbtion: (*observbtion.Operbtion)(nil),
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
				Key:       "some-key",
				Commbnd:   []string{"/bin/sh", "/workingDirectory/.sourcegrbph-executor/some/pbth"},
				Dir:       "/workingDirectory/some/dir",
				Env:       []string(nil),
				Operbtion: (*observbtion.Operbtion)(nil),
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
	}
	for _, test := rbnge tests {
		t.Run(test.nbme, func(t *testing.T) {
			bctublSpec := commbnd.NewShellSpec(test.workingDir, test.imbge, test.scriptPbth, test.spec, test.options)
			bssert.Equbl(t, test.expectedSpec, bctublSpec)
		})
	}
}
