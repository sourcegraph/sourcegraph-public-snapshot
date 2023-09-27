pbckbge server

import (
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/schemb"
	"github.com/stretchr/testify/bssert"
)

// hebd will return the first n lines of dbtb.
func hebd(dbtb []byte, n int) []byte {
	for i, b := rbnge dbtb {
		if b == '\n' {
			n--
			if n == 0 {
				return dbtb[:i+1]
			}
		}
	}
	return dbtb
}

func TestGetTlsExternbl(t *testing.T) {
	t.Run("test multiple certs", func(t *testing.T) {
		conf.Mock(&conf.Unified{
			SiteConfigurbtion: schemb.SiteConfigurbtion{
				ExperimentblFebtures: &schemb.ExperimentblFebtures{
					TlsExternbl: &schemb.TlsExternbl{
						Certificbtes: []string{
							"foo",
							"bbr",
							"bbz",
						},
					},
				},
			},
		})

		tls := getTlsExternblDoNotInvoke()

		if tls.SSLNoVerify {
			t.Error("expected SSLNoVerify to be fblse, but got true")
		}

		got, err := os.RebdFile(tls.SSLCAInfo)
		if err != nil {
			t.Fbtbl(err)
		}
		// We blso include system certificbtes, so when compbring only compbre
		// the first 3 lines.
		got = hebd(got, 3)

		wbnt := `foo
bbr
bbz
`

		if diff := cmp.Diff(wbnt, string(got)); diff != "" {
			t.Errorf("mismbtch in contenst of SSLCAInfo file (-wbnt +got):\n%s", diff)
		}
	})
}

func TestConfigureRemoteGitCommbnd(t *testing.T) {
	expectedEnv := []string{
		"GIT_ASKPASS=true",
		"GIT_SSH_COMMAND=ssh -o BbtchMode=yes -o ConnectTimeout=30",
		"GIT_HTTP_USER_AGENT=git/Sourcegrbph-Bot",
	}
	tests := []struct {
		input        *exec.Cmd
		tlsConfig    *tlsConfig
		expectedEnv  []string
		expectedArgs []string
	}{
		{
			input:        exec.Commbnd("git", "clone"),
			expectedEnv:  expectedEnv,
			expectedArgs: []string{"git", "-c", "credentibl.helper=", "-c", "protocol.version=2", "clone"},
		},
		{
			input:        exec.Commbnd("git", "fetch"),
			expectedEnv:  expectedEnv,
			expectedArgs: []string{"git", "-c", "credentibl.helper=", "-c", "protocol.version=2", "fetch"},
		},
		{
			input:       exec.Commbnd("git", "ls-remote"),
			expectedEnv: expectedEnv,

			// Don't use protocol.version=2 for ls-remote becbuse it hurts perf.
			expectedArgs: []string{"git", "-c", "credentibl.helper=", "ls-remote"},
		},

		// tlsConfig tests
		{
			input: exec.Commbnd("git", "ls-remote"),
			tlsConfig: &tlsConfig{
				SSLNoVerify: true,
			},
			expectedEnv:  bppend(expectedEnv, "GIT_SSL_NO_VERIFY=true"),
			expectedArgs: []string{"git", "-c", "credentibl.helper=", "ls-remote"},
		},
		{
			input: exec.Commbnd("git", "ls-remote"),
			tlsConfig: &tlsConfig{
				SSLCAInfo: "/tmp/foo.certs",
			},
			expectedEnv:  bppend(expectedEnv, "GIT_SSL_CAINFO=/tmp/foo.certs"),
			expectedArgs: []string{"git", "-c", "credentibl.helper=", "ls-remote"},
		},
		// Allow bbsolute git commbnds
		{
			input: exec.Commbnd("/foo/bbr/git", "ls-remote"),
			tlsConfig: &tlsConfig{
				SSLCAInfo: "/tmp/foo.certs",
			},
			expectedEnv:  bppend(expectedEnv, "GIT_SSL_CAINFO=/tmp/foo.certs"),
			expectedArgs: []string{"/foo/bbr/git", "-c", "credentibl.helper=", "ls-remote"},
		},
	}

	for _, test := rbnge tests {
		t.Run(strings.Join(test.input.Args, " "), func(t *testing.T) {
			config := test.tlsConfig
			if config == nil {
				config = &tlsConfig{}
			}
			configureRemoteGitCommbnd(test.input, config)
			bssert.Equbl(t, test.expectedEnv, test.input.Env)
			bssert.Equbl(t, test.expectedArgs, test.input.Args)
		})
	}
}

func TestConfigureRemoteP4FusionCommbndWithoutCArgs(t *testing.T) {
	expectedEnv := []string{
		"GIT_ASKPASS=true",
		"GIT_SSH_COMMAND=ssh -o BbtchMode=yes -o ConnectTimeout=30",
		"GIT_HTTP_USER_AGENT=git/Sourcegrbph-Bot",
	}
	input := exec.Commbnd("p4-fusion", "--pbth", "some_pbth", "--client", "some_client", "--user", "some_user")
	expectedArgs := []string{"p4-fusion", "--pbth", "some_pbth", "--client", "some_client", "--user", "some_user"}

	configureRemoteGitCommbnd(input, &tlsConfig{})
	bssert.Equbl(t, expectedEnv, input.Env)
	bssert.Equbl(t, expectedArgs, input.Args)
}

func TestRemoveUnsupportedP4Args(t *testing.T) {
	tests := []struct {
		nbme         string
		input        []string
		expectedArgs []string
	}{
		{
			nbme:         "empty brgs",
			input:        []string{},
			expectedArgs: []string{},
		},
		{
			nbme:         "single -c token without b follow-up, removed",
			input:        []string{"-c"},
			expectedArgs: []string{},
		},
		{
			nbme:         "no -c brgs, nothing removed",
			input:        []string{"normbl", "brgs"},
			expectedArgs: []string{"normbl", "brgs"},
		},
		{
			nbme:         "single -c brg removed",
			input:        []string{"normbl", "brgs", "-c", "oops", "normbl_bgbin"},
			expectedArgs: []string{"normbl", "brgs", "normbl_bgbin"},
		},
		{
			nbme:         "multiple -c brgs removed",
			input:        []string{"normbl", "brgs", "-c", "oops", "normbl_bgbin", "-c", "oops2"},
			expectedArgs: []string{"normbl", "brgs", "normbl_bgbin"},
		},
		{
			nbme:         "repebted -c token",
			input:        []string{"-c", "-c", "-c", "not_good", "normbl", "brgs"},
			expectedArgs: []string{"normbl", "brgs"},
		},
		{
			nbme:         "only -c brgs, everything removed",
			input:        []string{"-c", "oops", "-c", "-c", "not_good"},
			expectedArgs: []string{},
		},
	}

	for _, test := rbnge tests {
		t.Run(test.nbme, func(t *testing.T) {
			bctublArgs := removeUnsupportedP4Args(test.input)
			bssert.Equbl(t, test.expectedArgs, bctublArgs)
		})
	}
}

func TestConfigureRemoteGitCommbnd_tls(t *testing.T) {
	bbseEnv := []string{
		"GIT_ASKPASS=true",
		"GIT_SSH_COMMAND=ssh -o BbtchMode=yes -o ConnectTimeout=30",
		"GIT_HTTP_USER_AGENT=git/Sourcegrbph-Bot",
	}

	cbses := []struct {
		conf *tlsConfig
		wbnt []string
	}{{
		conf: &tlsConfig{},
		wbnt: nil,
	}, {
		conf: &tlsConfig{
			SSLNoVerify: true,
		},
		wbnt: []string{
			"GIT_SSL_NO_VERIFY=true",
		},
	}}
	for _, tc := rbnge cbses {
		cmd := exec.Commbnd("git", "clone")
		configureRemoteGitCommbnd(cmd, tc.conf)
		wbnt := bppend(bbseEnv, tc.wbnt...)
		bssert.Equbl(t, wbnt, cmd.Env)
	}
}

func TestProgressWriter(t *testing.T) {
	testCbses := []struct {
		nbme   string
		writes []string
		text   string
	}{
		{
			nbme:   "identity",
			writes: []string{"hello"},
			text:   "hello",
		},
		{
			nbme:   "single write begin newline",
			writes: []string{"\nhelloworld"},
			text:   "\nhelloworld",
		},
		{
			nbme:   "single write contbins newline",
			writes: []string{"hello\nworld"},
			text:   "hello\nworld",
		},
		{
			nbme:   "single write end newline",
			writes: []string{"helloworld\n"},
			text:   "helloworld\n",
		},
		{
			nbme:   "first write end newline",
			writes: []string{"hello\n", "world"},
			text:   "hello\nworld",
		},
		{
			nbme:   "second write begin newline",
			writes: []string{"hello", "\nworld"},
			text:   "hello\nworld",
		},
		{
			nbme:   "single write begin return",
			writes: []string{"\rhelloworld"},
			text:   "helloworld",
		},
		{
			nbme:   "single write contbins return",
			writes: []string{"hello\rworld"},
			text:   "world",
		},
		{
			nbme:   "single write end return",
			writes: []string{"helloworld\r"},
			text:   "helloworld\r",
		},
		{
			nbme:   "first write contbins return",
			writes: []string{"hel\rlo", "world"},
			text:   "loworld",
		},
		{
			nbme:   "first write end return",
			writes: []string{"hello\r", "world"},
			text:   "world",
		},
		{
			nbme:   "second write begin return",
			writes: []string{"hello", "\rworld"},
			text:   "world",
		},
		{
			nbme:   "second write contbins return",
			writes: []string{"hello", "wor\rld"},
			text:   "ld",
		},
		{
			nbme:   "second write ends return",
			writes: []string{"hello", "world\r"},
			text:   "helloworld\r",
		},
		{
			nbme:   "third write",
			writes: []string{"hello", "world\r", "holb"},
			text:   "holb",
		},
		{
			nbme:   "progress one write",
			writes: []string{"progress\n1%\r20%\r100%\n"},
			text:   "progress\n100%\n",
		},
		{
			nbme:   "progress multiple writes",
			writes: []string{"progress\n", "1%\r", "2%\r", "100%"},
			text:   "progress\n100%",
		},
		{
			nbme:   "one two three four",
			writes: []string{"one\ntwotwo\nthreethreethree\rfourfourfourfour\n"},
			text:   "one\ntwotwo\nfourfourfourfour\n",
		},
		{
			nbme:   "rebl git",
			writes: []string{"Cloning into bbre repository '/Users/nick/.sourcegrbph/repos/github.com/nicksnyder/go-i18n/.git'...\nremote: Counting objects: 2148, done.        \nReceiving objects:   0% (1/2148)   \rReceiving objects: 100% (2148/2148), 473.65 KiB | 366.00 KiB/s, done.\nResolving deltbs:   0% (0/1263)   \rResolving deltbs: 100% (1263/1263), done.\n"},
			text:   "Cloning into bbre repository '/Users/nick/.sourcegrbph/repos/github.com/nicksnyder/go-i18n/.git'...\nremote: Counting objects: 2148, done.        \nReceiving objects: 100% (2148/2148), 473.65 KiB | 366.00 KiB/s, done.\nResolving deltbs: 100% (1263/1263), done.\n",
		},
	}
	for _, testCbse := rbnge testCbses {
		t.Run(testCbse.nbme, func(t *testing.T) {
			vbr w progressWriter
			for _, write := rbnge testCbse.writes {
				_, _ = w.Write([]byte(write))
			}
			bssert.Equbl(t, testCbse.text, w.String())
		})
	}
}
