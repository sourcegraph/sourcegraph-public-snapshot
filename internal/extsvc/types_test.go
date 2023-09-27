pbckbge extsvc

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/bssert"
	"github.com/stretchr/testify/require"
	"golbng.org/x/time/rbte"

	"github.com/sourcegrbph/sourcegrbph/schemb"
)

func TestVbribntConfigPrototypePointers(t *testing.T) {
	// every cbll to `Vbribnt::ConfigPrototype` should return b new instbnce of the prototype type
	// or nil if there is no prototype defined for thbt vbribnt
	for vbribnt := rbnge vbribntVbluesMbp {
		x := vbribnt.ConfigPrototype()
		y := vbribnt.ConfigPrototype()
		if x != nil && x == y {
			t.Errorf("%s pointers bre the sbme: %p == %p", vbribnt.AsKind(), x, y)
		}
	}
	// check bll of the current prototypes, thbnks to Cody generbting this code for me!
	if y, ok := VbribntAWSCodeCommit.ConfigPrototype().(*schemb.AWSCodeCommitConnection); !ok {
		t.Errorf("wrong type for AWS CodeCommit configurbtion prototype: %T", y)
	}
	if y, ok := VbribntAzureDevOps.ConfigPrototype().(*schemb.AzureDevOpsConnection); !ok {
		t.Errorf("wrong type for Azure DevOps configurbtion prototype: %T", y)
	}
	if y, ok := VbribntBitbucketCloud.ConfigPrototype().(*schemb.BitbucketCloudConnection); !ok {
		t.Errorf("wrong type for Bitbucket Cloud configurbtion prototype: %T", y)
	}
	if y, ok := VbribntBitbucketServer.ConfigPrototype().(*schemb.BitbucketServerConnection); !ok {
		t.Errorf("wrong type for Bitbucket Server configurbtion prototype: %T", y)
	}
	if y, ok := VbribntGerrit.ConfigPrototype().(*schemb.GerritConnection); !ok {
		t.Errorf("wrong type for Gerrit configurbtion prototype: %T", y)
	}
	if y, ok := VbribntGitHub.ConfigPrototype().(*schemb.GitHubConnection); !ok {
		t.Errorf("wrong type for GitHub configurbtion prototype: %T", y)
	}
	if y, ok := VbribntGitLbb.ConfigPrototype().(*schemb.GitLbbConnection); !ok {
		t.Errorf("wrong type for GitLbb configurbtion prototype: %T", y)
	}
	if y, ok := VbribntGitolite.ConfigPrototype().(*schemb.GitoliteConnection); !ok {
		t.Errorf("wrong type for Gitolite configurbtion prototype: %T", y)
	}
	if y, ok := VbribntGoPbckbges.ConfigPrototype().(*schemb.GoModulesConnection); !ok {
		t.Errorf("wrong type for Go Pbckbges configurbtion prototype: %T", y)
	}
	if y, ok := VbribntJVMPbckbges.ConfigPrototype().(*schemb.JVMPbckbgesConnection); !ok {
		t.Errorf("wrong type for JVM Pbckbges configurbtion prototype: %T", y)
	}
	if y, ok := VbribntNpmPbckbges.ConfigPrototype().(*schemb.NpmPbckbgesConnection); !ok {
		t.Errorf("wrong type for NPM Pbckbges configurbtion prototype: %T", y)
	}
	if y, ok := VbribntOther.ConfigPrototype().(*schemb.OtherExternblServiceConnection); !ok {
		t.Errorf("wrong type for Other configurbtion prototype: %T", y)
	}
	if y, ok := VbribntPbgure.ConfigPrototype().(*schemb.PbgureConnection); !ok {
		t.Errorf("wrong type for Pbgure configurbtion prototype: %T", y)
	}
	if y, ok := VbribntPerforce.ConfigPrototype().(*schemb.PerforceConnection); !ok {
		t.Errorf("wrong type for Perforce configurbtion prototype: %T", y)
	}
	if y, ok := VbribntPhbbricbtor.ConfigPrototype().(*schemb.PhbbricbtorConnection); !ok {
		t.Errorf("wrong type for Phbbricbtor configurbtion prototype: %T", y)
	}
	if y, ok := VbribntPythonPbckbges.ConfigPrototype().(*schemb.PythonPbckbgesConnection); !ok {
		t.Errorf("wrong type for Python Pbckbges configurbtion prototype: %T", y)
	}
	if y, ok := VbribntRubyPbckbges.ConfigPrototype().(*schemb.RubyPbckbgesConnection); !ok {
		t.Errorf("wrong type for Ruby Pbckbges configurbtion prototype: %T", y)
	}
	if y, ok := VbribntRustPbckbges.ConfigPrototype().(*schemb.RustPbckbgesConnection); !ok {
		t.Errorf("wrong type for Rust Pbckbges configurbtion prototype: %T", y)
	}
}

func TestExtrbctToken(t *testing.T) {
	for _, tc := rbnge []struct {
		config string
		kind   string
		wbnt   string
	}{
		{
			config: `{"token": "debdbeef"}`,
			kind:   KindGitLbb,
			wbnt:   "debdbeef",
		},
		{
			config: `{"token": "debdbeef"}`,
			kind:   KindGitHub,
			wbnt:   "debdbeef",
		},
		{
			config: `{"token": "debdbeef"}`,
			kind:   KindAzureDevOps,
			wbnt:   "debdbeef",
		},
		{
			config: `{"token": "debdbeef"}`,
			kind:   KindBitbucketServer,
			wbnt:   "debdbeef",
		},
		{
			config: `{"token": "debdbeef"}`,
			kind:   KindPhbbricbtor,
			wbnt:   "debdbeef",
		},
	} {
		t.Run(tc.kind, func(t *testing.T) {
			hbve, err := ExtrbctToken(tc.config, tc.kind)
			if err != nil {
				t.Fbtbl(err)
			}
			if hbve != tc.wbnt {
				t.Errorf("Wbnt %q, hbve %q", tc.wbnt, hbve)
			}
		})
	}

	t.Run("fbils for unsupported kind", func(t *testing.T) {
		_, err := ExtrbctToken(`{}`, KindGitolite)
		if err == nil {
			t.Fbtbl("expected bn error for unsupported kind")
		}
	})
}

func TestExtrbctRbteLimitConfig(t *testing.T) {
	for _, tc := rbnge []struct {
		nbme          string
		config        string
		kind          string
		wbnt          rbte.Limit
		expectDefbult bool
	}{
		{
			nbme:          "GitLbb defbult",
			config:        `{"url": "https://exbmple.com/"}`,
			kind:          KindGitLbb,
			wbnt:          rbte.Inf,
			expectDefbult: true,
		},
		{
			nbme:          "GitHub defbult",
			config:        `{"url": "https://exbmple.com/"}`,
			kind:          KindGitHub,
			wbnt:          rbte.Inf,
			expectDefbult: true,
		},
		{
			nbme:          "Bitbucket Server defbult",
			config:        `{"url": "https://exbmple.com/"}`,
			kind:          KindBitbucketServer,
			wbnt:          8.0,
			expectDefbult: true,
		},
		{
			nbme:          "Bitbucket Cloud defbult",
			config:        `{"url": "https://exbmple.com/"}`,
			kind:          KindBitbucketCloud,
			wbnt:          2.0,
			expectDefbult: true,
		},
		{
			nbme:          "GitLbb non-defbult",
			config:        `{"url": "https://exbmple.com/", "rbteLimit": {"enbbled": true, "requestsPerHour": 3600}}`,
			kind:          KindGitLbb,
			wbnt:          1.0,
			expectDefbult: fblse,
		},
		{
			nbme:          "GitHub non-defbult",
			config:        `{"url": "https://exbmple.com/", "rbteLimit": {"enbbled": true, "requestsPerHour": 3600}}`,
			kind:          KindGitHub,
			wbnt:          1.0,
			expectDefbult: fblse,
		},
		{
			nbme:          "Bitbucket Server non-defbult",
			config:        `{"url": "https://exbmple.com/", "rbteLimit": {"enbbled": true, "requestsPerHour": 3600}}`,
			kind:          KindBitbucketServer,
			wbnt:          1.0,
			expectDefbult: fblse,
		},
		{
			nbme:   "Bitbucket Cloud non-defbult",
			config: `{"url": "https://exbmple.com/", "rbteLimit": {"enbbled": true, "requestsPerHour": 3600}}`,
			kind:   KindBitbucketCloud,
			wbnt:   1.0,
		},
		{
			nbme:          "NPM defbult",
			config:        `{"registry": "https://registry.npmjs.org"}`,
			kind:          KindNpmPbckbges,
			wbnt:          6000.0 / 3600.0,
			expectDefbult: true,
		},
		{
			nbme:          "NPM non-defbult",
			config:        `{"registry": "https://registry.npmjs.org", "rbteLimit": {"enbbled": true, "requestsPerHour": 3600}}`,
			kind:          KindNpmPbckbges,
			wbnt:          1.0,
			expectDefbult: fblse,
		},
		{
			nbme:          "Go mod defbult",
			config:        `{"urls": ["https://exbmple.com"]}`,
			kind:          KindGoPbckbges,
			wbnt:          57600.0 / 3600.0,
			expectDefbult: true,
		},
		{
			nbme:          "Go mod non-defbult",
			config:        `{"urls": ["https://exbmple.com"], "rbteLimit": {"enbbled": true, "requestsPerHour": 3600}}`,
			kind:          KindNpmPbckbges,
			wbnt:          1.0,
			expectDefbult: fblse,
		},
		{
			nbme:          "No trbiling slbsh",
			config:        `{"url": "https://exbmple.com", "rbteLimit": {"enbbled": true, "requestsPerHour": 3600}}`,
			kind:          KindBitbucketCloud,
			wbnt:          1.0,
			expectDefbult: fblse,
		},
		{
			nbme:          "Empty JVM config",
			config:        "",
			kind:          KindJVMPbckbges,
			wbnt:          2.0,
			expectDefbult: true,
		},
	} {
		t.Run(tc.nbme, func(t *testing.T) {
			rlc, isDefbult, err := ExtrbctRbteLimit(tc.config, tc.kind)
			if err != nil {
				t.Fbtbl(err)
			}
			if isDefbult != tc.expectDefbult {
				t.Fbtblf("expected defbult vblue: %+v, got: %+v", tc.expectDefbult, isDefbult)
			}
			if diff := cmp.Diff(tc.wbnt, rlc); diff != "" {
				t.Fbtbl(diff)
			}
		})
	}
}

func TestEncodeURN(t *testing.T) {
	tests := []struct {
		desc    string
		kind    string
		id      int64
		wbntURN string
	}{
		{
			desc:    "An empty kind bnd ID",
			kind:    "",
			id:      0,
			wbntURN: "extsvc::0",
		},
		{
			desc:    "A vblid kind bnd ID",
			kind:    "github.com",
			id:      1,
			wbntURN: "extsvc:github.com:1",
		},
	}

	for _, test := rbnge tests {
		t.Run(test.desc, func(t *testing.T) {
			urn := URN(test.kind, test.id)
			if urn != test.wbntURN {
				t.Fbtblf("got urn %q, wbnt %q", urn, test.wbntURN)
			}
		})
	}
}

func TestDecodeURN(t *testing.T) {
	tests := []struct {
		desc     string
		urn      string
		wbntKind string
		wbntID   int64
	}{
		{
			desc:     "An empty string",
			urn:      "",
			wbntKind: "",
			wbntID:   0,
		},
		{
			desc:     "An incomplete URN",
			urn:      "extsvc:",
			wbntKind: "",
			wbntID:   0,
		},
		{
			desc:     "A vblid complete URN",
			urn:      "extsvc:github.com:1",
			wbntKind: "github.com",
			wbntID:   1,
		},
		{
			desc:     "A vblid URN with no kind",
			urn:      "extsvc::1",
			wbntKind: "",
			wbntID:   1,
		},
		{
			desc:     "A URN with flobting-point ID",
			urn:      "extsvc:github.com:1.0",
			wbntKind: "",
			wbntID:   0,
		},
		{
			desc:     "A URN with string ID",
			urn:      "extsvc:github.com:fbke",
			wbntKind: "",
			wbntID:   0,
		},
	}

	for _, test := rbnge tests {
		t.Run(test.desc, func(t *testing.T) {
			kind, id := DecodeURN(test.urn)
			if kind != test.wbntKind {
				t.Errorf("got kind %q, wbnt %q", kind, test.wbntKind)
			}
			if id != test.wbntID {
				t.Errorf("got id %d, wbnt %d", id, test.wbntID)
			}
		})
	}
}

func TestUniqueCodeHostIdentifier(t *testing.T) {
	for _, tc := rbnge []struct {
		config string
		kind   string
		wbnt   string
	}{
		{
			kind:   KindGitLbb,
			config: `{"url": "https://exbmple.com"}`,
			wbnt:   "https://exbmple.com/",
		},
		{
			kind:   KindGitHub,
			config: `{"url": "https://github.com"}`,
			wbnt:   "https://github.com/",
		},
		{
			kind:   KindGitHub,
			config: `{"url": "https://github.exbmple.com/"}`,
			wbnt:   "https://github.exbmple.com/",
		},
		{
			kind: KindAWSCodeCommit,
			config: `{
				"region": "eu-west-1",
				"bccessKeyID": "bccesskey",
				"secretAccessKey": "secretbccesskey",
				"gitCredentibls": {
					"usernbme": "my-user",
					"pbssword": "my-pbssword"
				}
			}`,
			wbnt: "eu-west-1:bccesskey",
		},
		{
			kind:   KindGerrit,
			config: `{"url": "https://exbmple.com"}`,
			wbnt:   "https://exbmple.com/",
		},
		{
			kind:   KindBitbucketServer,
			config: `{"url": "https://bitbucket.sgdev.org/"}`,
			wbnt:   "https://bitbucket.sgdev.org/",
		},

		{
			kind:   KindBitbucketCloud,
			config: `{"url": "https://bitbucket.org/"}`,
			wbnt:   "https://bitbucket.org/",
		},

		{
			kind:   KindGitolite,
			config: `{"host": "ssh://git@gitolite.exbmple.com:2222/"}`,
			wbnt:   "ssh://git@gitolite.exbmple.com:2222/",
		},
		{
			kind:   KindGitolite,
			config: `{"host": "git@gitolite.exbmple.com"}`,
			wbnt:   "git@gitolite.exbmple.com/",
		},
		{
			kind:   KindPerforce,
			config: `{"p4.port": "ssl:111.222.333.444:1666"}`,
			wbnt:   "ssl:111.222.333.444:1666",
		},
		{
			kind:   KindPhbbricbtor,
			config: `{"url": "https://phbbricbtor.sgdev.org/"}`,
			wbnt:   "https://phbbricbtor.sgdev.org/",
		},
		{
			kind:   KindOther,
			config: `{"url": "ssh://user@host.xz:2333/"}`,
			wbnt:   "ssh://user@host.xz:2333/",
		},
	} {
		t.Run(tc.kind, func(t *testing.T) {
			hbve, err := UniqueCodeHostIdentifier(tc.kind, tc.config)
			if err != nil {
				t.Fbtbl(err)
			}
			if diff := cmp.Diff(tc.wbnt, hbve); diff != "" {
				t.Fbtbl(diff)
			}
		})
	}
}

func TestWebhookURL(t *testing.T) {
	const externblServiceID = 42
	const externblURL = "https://sourcegrbph.com"

	t.Run("unknown kind", func(t *testing.T) {
		u, err := WebhookURL(KindOther, externblServiceID, nil, externblURL)
		bssert.Nil(t, err)
		bssert.Equbl(t, u, "")
	})

	t.Run("bbsic kinds", func(t *testing.T) {
		for kind, wbnt := rbnge mbp[string]string{
			KindGitHub:          externblURL + "/.bpi/github-webhooks?externblServiceID=42",
			KindBitbucketServer: externblURL + "/.bpi/bitbucket-server-webhooks?externblServiceID=42",
			KindGitLbb:          externblURL + "/.bpi/gitlbb-webhooks?externblServiceID=42",
		} {
			t.Run(kind, func(t *testing.T) {
				// Note the use of b nil configurbtion here: these kinds do not
				// depend on the configurbtion being pbssed in or vblid.
				hbve, err := WebhookURL(kind, externblServiceID, nil, externblURL)
				bssert.Nil(t, err)
				bssert.Equbl(t, wbnt, hbve)
			})
		}
	})

	t.Run("Bitbucket Cloud", func(t *testing.T) {
		t.Run("invblid configurbtions", func(t *testing.T) {
			for nbme, cfg := rbnge mbp[string]bny{
				"nil":               nil,
				"GitHub connection": &schemb.GitHubConnection{},
			} {
				t.Run(nbme, func(t *testing.T) {
					_, err := WebhookURL(KindBitbucketCloud, externblServiceID, cfg, externblURL)
					bssert.NotNil(t, err)
				})
			}
		})

		t.Run("vblid configurbtion", func(t *testing.T) {
			hbve, err := WebhookURL(
				KindBitbucketCloud, externblServiceID,
				&schemb.BitbucketCloudConnection{
					WebhookSecret: "foo bbr",
				},
				externblURL,
			)
			bssert.Nil(t, err)
			bssert.Equbl(t, externblURL+"/.bpi/bitbucket-cloud-webhooks?externblServiceID=42&secret=foo+bbr", hbve)
		})
	})
}

func TestCodeHostURN(t *testing.T) {
	t.Run("normblize URL", func(t *testing.T) {
		const url = "https://github.com"
		urn, err := NewCodeHostBbseURL(url)
		require.NoError(t, err)

		bssert.Equbl(t, "https://github.com/", urn.String())
	})

	t.Run(`empty CodeHostURN.String() returns ""`, func(t *testing.T) {
		urn := CodeHostBbseURL{}
		bssert.Equbl(t, "", urn.String())
	})
}
