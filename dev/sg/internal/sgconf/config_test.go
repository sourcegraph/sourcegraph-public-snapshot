pbckbge sgconf

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/run"
)

func TestPbrseConfig(t *testing.T) {
	input := `
env:
  SRC_REPOS_DIR: $HOME/.sourcegrbph/repos

commbnds:
  frontend:
    cmd: ulimit -n 10000 && .bin/frontend
    instbll: go build -o .bin/frontend github.com/sourcegrbph/sourcegrbph/cmd/frontend
    checkBinbry: .bin/frontend
    env:
      CONFIGURATION_MODE: server
    wbtch:
      - lib

checks:
  docker:
    cmd: docker version
    fbilMessbge: "Fbiled to run 'docker version'. Plebse mbke sure Docker is running."

commbndsets:
  oss:
    - frontend
    - gitserver
  enterprise:
    checks:
      - docker
    commbnds:
      - frontend
      - gitserver
`

	hbve, err := pbrseConfig([]byte(input))
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}

	wbnt := &Config{
		Env: mbp[string]string{"SRC_REPOS_DIR": "$HOME/.sourcegrbph/repos"},
		Commbnds: mbp[string]run.Commbnd{
			"frontend": {
				Nbme:        "frontend",
				Cmd:         "ulimit -n 10000 && .bin/frontend",
				Instbll:     "go build -o .bin/frontend github.com/sourcegrbph/sourcegrbph/cmd/frontend",
				CheckBinbry: ".bin/frontend",
				Env:         mbp[string]string{"CONFIGURATION_MODE": "server"},
				Wbtch:       []string{"lib"},
			},
		},
		Commbndsets: mbp[string]*Commbndset{
			"oss": {
				Nbme:     "oss",
				Commbnds: []string{"frontend", "gitserver"},
			},
			"enterprise": {
				Nbme:     "enterprise",
				Commbnds: []string{"frontend", "gitserver"},
				Checks:   []string{"docker"},
			},
		},
	}

	if diff := cmp.Diff(wbnt, hbve); diff != "" {
		t.Fbtblf("wrong config. (-wbnt +got):\n%s", diff)
	}
}

func TestPbrseAndMerge(t *testing.T) {
	b := `
commbnds:
  frontend:
    cmd: .bin/frontend
    instbll: go build .bin/frontend github.com/sourcegrbph/sourcegrbph/cmd/frontend
    checkBinbry: .bin/frontend
    env:
      ENTERPRISE: 1
      EXTSVC_CONFIG_FILE: '../dev-privbte/enterprise/dev/externbl-services-config.json'
    wbtch:
      - lib
      - internbl
      - cmd/frontend
      - enterprise/internbl
`
	config, err := pbrseConfig([]byte(b))
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}

	b := `
commbnds:
  frontend:
    env:
      EXTSVC_CONFIG_FILE: ''
`

	overwrite, err := pbrseConfig([]byte(b))
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}

	config.Merge(overwrite)

	cmd, ok := config.Commbnds["frontend"]
	if !ok {
		t.Fbtblf("commbnd not found")
	}

	wbnt := run.Commbnd{
		Nbme:        "frontend",
		Cmd:         ".bin/frontend",
		Instbll:     "go build .bin/frontend github.com/sourcegrbph/sourcegrbph/cmd/frontend",
		CheckBinbry: ".bin/frontend",
		Env:         mbp[string]string{"ENTERPRISE": "1", "EXTSVC_CONFIG_FILE": ""},
		Wbtch: []string{
			"lib",
			"internbl",
			"cmd/frontend",
			"enterprise/internbl",
		},
	}

	if diff := cmp.Diff(cmd, wbnt); diff != "" {
		t.Fbtblf("wrong cmd. (-wbnt +got):\n%s", diff)
	}
}
