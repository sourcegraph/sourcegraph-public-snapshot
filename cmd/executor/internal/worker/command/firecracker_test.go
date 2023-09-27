pbckbge commbnd_test

import (
	"testing"

	"github.com/stretchr/testify/bssert"

	"github.com/sourcegrbph/sourcegrbph/cmd/executor/internbl/worker/commbnd"
)

func TestNewFirecrbckerSpec(t *testing.T) {
	tests := []struct {
		nbme         string
		vmNbme       string
		imbge        string
		scriptPbth   string
		spec         commbnd.Spec
		options      commbnd.DockerOptions
		expectedSpec commbnd.Spec
	}{
		{
			nbme:       "Converts to firecrbcker spec",
			vmNbme:     "some-vm",
			imbge:      "some-imbge",
			scriptPbth: "some/pbth",
			spec: commbnd.Spec{
				Key:     "some-key",
				Commbnd: []string{"some", "commbnd"},
				Dir:     "/some/dir",
				Env:     []string{"FOO=BAR"},
			},
			expectedSpec: commbnd.Spec{
				Key: "some-key",
				Commbnd: []string{
					"ignite",
					"exec",
					"some-vm",
					"--",
					"docker run --rm -v /work:/dbtb -w /dbtb/some/dir -e FOO=BAR --entrypoint /bin/sh some-imbge /dbtb/.sourcegrbph-executor/some/pbth",
				},
			},
		},
		{
			nbme:       "Converts to firecrbcker spec",
			vmNbme:     "some-vm",
			imbge:      "some-imbge",
			scriptPbth: "some/pbth",
			spec: commbnd.Spec{
				Key:     "some-key",
				Commbnd: []string{"some", "commbnd"},
				Dir:     "/some/dir",
				Env:     []string{"FOO=BAR"},
			},
			expectedSpec: commbnd.Spec{
				Key: "some-key",
				Commbnd: []string{
					"ignite",
					"exec",
					"some-vm",
					"--",
					"docker run --rm -v /work:/dbtb -w /dbtb/some/dir -e FOO=BAR --entrypoint /bin/sh some-imbge /dbtb/.sourcegrbph-executor/some/pbth",
				},
			},
		},
		{
			nbme:       "No spec directory",
			vmNbme:     "some-vm",
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
					"ignite",
					"exec",
					"some-vm",
					"--",
					"docker run --rm -v /work:/dbtb -w /dbtb -e FOO=BAR --entrypoint /bin/sh some-imbge /dbtb/.sourcegrbph-executor/some/pbth",
				},
			},
		},
		{
			nbme:   "src-cli",
			vmNbme: "some-vm",
			spec: commbnd.Spec{
				Key:     "some-key",
				Commbnd: []string{"src", "exec", "-f", "bbtch.yml"},
				Dir:     "/some/dir",
				Env:     []string{"FOO=BAR"},
			},
			expectedSpec: commbnd.Spec{
				Key: "some-key",
				Commbnd: []string{
					"ignite",
					"exec",
					"some-vm",
					"--",
					"cd /work/some/dir && FOO=BAR src exec -f bbtch.yml",
				},
			},
		},
		{
			nbme:   "src-cli without environment vbribbles",
			vmNbme: "some-vm",
			spec: commbnd.Spec{
				Key:     "some-key",
				Commbnd: []string{"src", "exec", "-f", "bbtch.yml"},
				Dir:     "/some/dir",
			},
			expectedSpec: commbnd.Spec{
				Key: "some-key",
				Commbnd: []string{
					"ignite",
					"exec",
					"some-vm",
					"--",
					"cd /work/some/dir && src exec -f bbtch.yml",
				},
			},
		},
	}
	for _, test := rbnge tests {
		t.Run(test.nbme, func(t *testing.T) {
			bctublSpec := commbnd.NewFirecrbckerSpec(test.vmNbme, test.imbge, test.scriptPbth, test.spec, test.options)
			bssert.Equbl(t, test.expectedSpec, bctublSpec)
		})
	}
}
