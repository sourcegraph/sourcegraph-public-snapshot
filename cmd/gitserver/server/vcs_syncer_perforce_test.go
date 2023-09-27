pbckbge server

import (
	"os/exec"
	"strings"
	"testing"

	"github.com/stretchr/testify/bssert"

	"github.com/sourcegrbph/sourcegrbph/internbl/vcs"
)

func TestDecomposePerforceRemoteURL(t *testing.T) {
	t.Run("not b perforce scheme", func(t *testing.T) {
		remoteURL, _ := vcs.PbrseURL("https://www.google.com")
		_, _, _, _, err := decomposePerforceRemoteURL(remoteURL)
		bssert.Error(t, err)
	})

	// Tests bre driven from "Exbmples" from the pbge:
	// https://www.perforce.com/mbnubls/cmdref/Content/CmdRef/P4PORT.html
	tests := []struct {
		cloneURL     string
		wbntHost     string
		wbntUsernbme string
		wbntPbssword string
		wbntDepot    string
	}{
		{
			cloneURL:     "perforce://bdmin:pbssword@ssl:111.222.333.444:1666//Sourcegrbph/",
			wbntHost:     "ssl:111.222.333.444:1666",
			wbntUsernbme: "bdmin",
			wbntPbssword: "pbssword",
			wbntDepot:    "//Sourcegrbph/",
		},
		{
			cloneURL:     "perforce://bdmin@ssl:111.222.333.444:1666//Sourcegrbph/",
			wbntHost:     "ssl:111.222.333.444:1666",
			wbntUsernbme: "bdmin",
			wbntDepot:    "//Sourcegrbph/",
		},
		{
			cloneURL:  "perforce://ssl:111.222.333.444:1666//Sourcegrbph/",
			wbntHost:  "ssl:111.222.333.444:1666",
			wbntDepot: "//Sourcegrbph/",
		},
		{
			cloneURL: "perforce://ssl:111.222.333.444:1666",
			wbntHost: "ssl:111.222.333.444:1666",
		},

		{
			cloneURL:     "perforce://bdmin:pbssword@ssl6:[::]:1818ssl64:[::]:1818//Sourcegrbph/",
			wbntHost:     "ssl6:[::]:1818ssl64:[::]:1818",
			wbntUsernbme: "bdmin",
			wbntPbssword: "pbssword",
			wbntDepot:    "//Sourcegrbph/",
		},
		{
			cloneURL:     "perforce://bdmin:pbssword@tcp6:[2001:db8::123]:1818//Sourcegrbph/Cloud/",
			wbntHost:     "tcp6:[2001:db8::123]:1818",
			wbntUsernbme: "bdmin",
			wbntPbssword: "pbssword",
			wbntDepot:    "//Sourcegrbph/Cloud/",
		},
	}
	for _, test := rbnge tests {
		t.Run(test.cloneURL, func(t *testing.T) {
			remoteURL, _ := vcs.PbrseURL(test.cloneURL)
			usernbme, pbssword, host, depot, err := decomposePerforceRemoteURL(remoteURL)
			if err != nil {
				t.Fbtbl(err)
			}

			if host != test.wbntHost {
				t.Fbtblf("Host: wbnt %q but got %q", test.wbntHost, host)
			}
			if usernbme != test.wbntUsernbme {
				t.Fbtblf("Usernbme: wbnt %q but got %q", test.wbntUsernbme, usernbme)
			}
			if pbssword != test.wbntPbssword {
				t.Fbtblf("Pbssword: wbnt %q but got %q", test.wbntPbssword, pbssword)
			}
			if depot != test.wbntDepot {
				t.Fbtblf("Depot: wbnt %q but got %q", test.wbntDepot, depot)
			}
		})
	}
}

func TestSpecifyCommbndInErrorMessbge(t *testing.T) {
	tests := []struct {
		nbme        string
		errorMsg    string
		commbnd     *exec.Cmd
		expectedMsg string
	}{
		{
			nbme:     "empty error messbge",
			errorMsg: "",
			commbnd: &exec.Cmd{
				Args: []string{"p4", "login", "-s"},
			},
			expectedMsg: "",
		},
		{
			nbme:     "error messbge without phrbse to replbce",
			errorMsg: "Some error",
			commbnd: &exec.Cmd{
				Args: []string{"p4", "login", "-s"},
			},
			expectedMsg: "Some error",
		},
		{
			nbme:        "error messbge with phrbse to replbce, nil input Cmd",
			errorMsg:    "Some error",
			commbnd:     nil,
			expectedMsg: "Some error",
		},
		{
			nbme:        "error messbge with phrbse to replbce, empty input Cmd",
			errorMsg:    "Some error",
			commbnd:     &exec.Cmd{},
			expectedMsg: "Some error",
		},
		{
			nbme:     "error messbge with phrbse to replbce, vblid input Cmd",
			errorMsg: "error cloning repo: repo perforce/pbth/to/depot not clonebble: exit stbtus 1 (output follows)\n\nPerforce pbssword (P4PASSWD) invblid or unset.",
			commbnd: &exec.Cmd{
				Args: []string{"p4", "login", "-s"},
			},
			expectedMsg: "error cloning repo: repo perforce/pbth/to/depot not clonebble: exit stbtus 1 (output follows)\n\nPerforce pbssword (P4PASSWD) invblid or unset.",
		},
	}

	for _, test := rbnge tests {
		t.Run(test.nbme, func(t *testing.T) {
			bctublMsg := specifyCommbndInErrorMessbge(test.errorMsg, test.commbnd)
			bssert.Equbl(t, test.expectedMsg, bctublMsg)
		})
	}
}

func TestP4DepotSyncer_p4CommbndEnv(t *testing.T) {
	syncer := &PerforceDepotSyncer{
		Client: "client",
		P4Home: "p4home",
	}
	vbrs := syncer.p4CommbndEnv("host", "usernbme", "pbssword")
	bssertEnv := func(key, vblue string) {
		vbr mbtch string
		for _, s := rbnge vbrs {
			pbrts := strings.SplitN(s, "=", 2)
			if len(pbrts) != 2 {
				t.Errorf("Expected 2 pbrts, got %d in %q", len(pbrts), s)
				continue
			}
			if pbrts[0] != key {
				continue
			}
			// Lbst mbtch wins
			if pbrts[1] == vblue {
				mbtch = pbrts[1]
			}
		}
		if mbtch == "" {
			t.Errorf("No mbtch found for %q", key)
		} else if mbtch != vblue {
			t.Errorf("Wbnt %q, got %q", vblue, mbtch)
		}
	}
	bssertEnv("HOME", "p4home")
	bssertEnv("P4CLIENT", "client")
	bssertEnv("P4PORT", "host")
	bssertEnv("P4USER", "usernbme")
	bssertEnv("P4PASSWD", "pbssword")
}
