package internal

import (
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"

	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/schema"
)

// head will return the first n lines of data.
func head(data []byte, n int) []byte {
	for i, b := range data {
		if b == '\n' {
			n--
			if n == 0 {
				return data[:i+1]
			}
		}
	}
	return data
}

func TestGetTlsExternal(t *testing.T) {
	t.Run("test multiple certs", func(t *testing.T) {
		conf.Mock(&conf.Unified{
			SiteConfiguration: schema.SiteConfiguration{
				ExperimentalFeatures: &schema.ExperimentalFeatures{
					TlsExternal: &schema.TlsExternal{
						Certificates: []string{
							"foo",
							"bar",
							"baz",
						},
					},
				},
			},
		})

		tls := getTlsExternalDoNotInvoke()

		if tls.SSLNoVerify {
			t.Error("expected SSLNoVerify to be false, but got true")
		}

		got, err := os.ReadFile(tls.SSLCAInfo)
		if err != nil {
			t.Fatal(err)
		}
		// We also include system certificates, so when comparing only compare
		// the first 3 lines.
		got = head(got, 3)

		want := `foo
bar
baz
`

		if diff := cmp.Diff(want, string(got)); diff != "" {
			t.Errorf("mismatch in contenst of SSLCAInfo file (-want +got):\n%s", diff)
		}
	})
}

func TestConfigureRemoteGitCommand(t *testing.T) {
	expectedEnv := []string{
		"GIT_ASKPASS=true",
		"GIT_SSH_COMMAND=ssh -o BatchMode=yes -o ConnectTimeout=30",
		"GIT_HTTP_USER_AGENT=git/Sourcegraph-Bot",
	}
	tests := []struct {
		input        *exec.Cmd
		tlsConfig    *tlsConfig
		expectedEnv  []string
		expectedArgs []string
	}{
		{
			input:        exec.Command("git", "clone"),
			expectedEnv:  expectedEnv,
			expectedArgs: []string{"git", "-c", "credential.helper=", "-c", "protocol.version=2", "clone"},
		},
		{
			input:        exec.Command("git", "fetch"),
			expectedEnv:  expectedEnv,
			expectedArgs: []string{"git", "-c", "credential.helper=", "-c", "protocol.version=2", "fetch"},
		},
		{
			input:       exec.Command("git", "ls-remote"),
			expectedEnv: expectedEnv,

			// Don't use protocol.version=2 for ls-remote because it hurts perf.
			expectedArgs: []string{"git", "-c", "credential.helper=", "ls-remote"},
		},

		// tlsConfig tests
		{
			input: exec.Command("git", "ls-remote"),
			tlsConfig: &tlsConfig{
				SSLNoVerify: true,
			},
			expectedEnv:  append(expectedEnv, "GIT_SSL_NO_VERIFY=true"),
			expectedArgs: []string{"git", "-c", "credential.helper=", "ls-remote"},
		},
		{
			input: exec.Command("git", "ls-remote"),
			tlsConfig: &tlsConfig{
				SSLCAInfo: "/tmp/foo.certs",
			},
			expectedEnv:  append(expectedEnv, "GIT_SSL_CAINFO=/tmp/foo.certs"),
			expectedArgs: []string{"git", "-c", "credential.helper=", "ls-remote"},
		},
		// Allow absolute git commands
		{
			input: exec.Command("/foo/bar/git", "ls-remote"),
			tlsConfig: &tlsConfig{
				SSLCAInfo: "/tmp/foo.certs",
			},
			expectedEnv:  append(expectedEnv, "GIT_SSL_CAINFO=/tmp/foo.certs"),
			expectedArgs: []string{"/foo/bar/git", "-c", "credential.helper=", "ls-remote"},
		},
	}

	for _, test := range tests {
		t.Run(strings.Join(test.input.Args, " "), func(t *testing.T) {
			config := test.tlsConfig
			if config == nil {
				config = &tlsConfig{}
			}
			configureRemoteGitCommand(test.input, config)
			assert.Equal(t, test.expectedEnv, test.input.Env)
			assert.Equal(t, test.expectedArgs, test.input.Args)
		})
	}
}

func TestConfigureRemoteP4FusionCommandWithoutCArgs(t *testing.T) {
	expectedEnv := []string{
		"GIT_ASKPASS=true",
		"GIT_SSH_COMMAND=ssh -o BatchMode=yes -o ConnectTimeout=30",
		"GIT_HTTP_USER_AGENT=git/Sourcegraph-Bot",
	}
	input := exec.Command("p4-fusion", "--path", "some_path", "--client", "some_client", "--user", "some_user")
	expectedArgs := []string{"p4-fusion", "--path", "some_path", "--client", "some_client", "--user", "some_user"}

	configureRemoteGitCommand(input, &tlsConfig{})
	assert.Equal(t, expectedEnv, input.Env)
	assert.Equal(t, expectedArgs, input.Args)
}

func TestRemoveUnsupportedP4Args(t *testing.T) {
	tests := []struct {
		name         string
		input        []string
		expectedArgs []string
	}{
		{
			name:         "empty args",
			input:        []string{},
			expectedArgs: []string{},
		},
		{
			name:         "single -c token without a follow-up, removed",
			input:        []string{"-c"},
			expectedArgs: []string{},
		},
		{
			name:         "no -c args, nothing removed",
			input:        []string{"normal", "args"},
			expectedArgs: []string{"normal", "args"},
		},
		{
			name:         "single -c arg removed",
			input:        []string{"normal", "args", "-c", "oops", "normal_again"},
			expectedArgs: []string{"normal", "args", "normal_again"},
		},
		{
			name:         "multiple -c args removed",
			input:        []string{"normal", "args", "-c", "oops", "normal_again", "-c", "oops2"},
			expectedArgs: []string{"normal", "args", "normal_again"},
		},
		{
			name:         "repeated -c token",
			input:        []string{"-c", "-c", "-c", "not_good", "normal", "args"},
			expectedArgs: []string{"normal", "args"},
		},
		{
			name:         "only -c args, everything removed",
			input:        []string{"-c", "oops", "-c", "-c", "not_good"},
			expectedArgs: []string{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actualArgs := removeUnsupportedP4Args(test.input)
			assert.Equal(t, test.expectedArgs, actualArgs)
		})
	}
}

func TestConfigureRemoteGitCommand_tls(t *testing.T) {
	baseEnv := []string{
		"GIT_ASKPASS=true",
		"GIT_SSH_COMMAND=ssh -o BatchMode=yes -o ConnectTimeout=30",
		"GIT_HTTP_USER_AGENT=git/Sourcegraph-Bot",
	}

	cases := []struct {
		conf *tlsConfig
		want []string
	}{{
		conf: &tlsConfig{},
		want: nil,
	}, {
		conf: &tlsConfig{
			SSLNoVerify: true,
		},
		want: []string{
			"GIT_SSL_NO_VERIFY=true",
		},
	}}
	for _, tc := range cases {
		cmd := exec.Command("git", "clone")
		configureRemoteGitCommand(cmd, tc.conf)
		want := append(baseEnv, tc.want...)
		assert.Equal(t, want, cmd.Env)
	}
}

func TestProgressWriter(t *testing.T) {
	testCases := []struct {
		name   string
		writes []string
		text   string
	}{
		{
			name:   "identity",
			writes: []string{"hello"},
			text:   "hello",
		},
		{
			name:   "single write begin newline",
			writes: []string{"\nhelloworld"},
			text:   "\nhelloworld",
		},
		{
			name:   "single write contains newline",
			writes: []string{"hello\nworld"},
			text:   "hello\nworld",
		},
		{
			name:   "single write end newline",
			writes: []string{"helloworld\n"},
			text:   "helloworld\n",
		},
		{
			name:   "first write end newline",
			writes: []string{"hello\n", "world"},
			text:   "hello\nworld",
		},
		{
			name:   "second write begin newline",
			writes: []string{"hello", "\nworld"},
			text:   "hello\nworld",
		},
		{
			name:   "single write begin return",
			writes: []string{"\rhelloworld"},
			text:   "helloworld",
		},
		{
			name:   "single write contains return",
			writes: []string{"hello\rworld"},
			text:   "world",
		},
		{
			name:   "single write end return",
			writes: []string{"helloworld\r"},
			text:   "helloworld\r",
		},
		{
			name:   "first write contains return",
			writes: []string{"hel\rlo", "world"},
			text:   "loworld",
		},
		{
			name:   "first write end return",
			writes: []string{"hello\r", "world"},
			text:   "world",
		},
		{
			name:   "second write begin return",
			writes: []string{"hello", "\rworld"},
			text:   "world",
		},
		{
			name:   "second write contains return",
			writes: []string{"hello", "wor\rld"},
			text:   "ld",
		},
		{
			name:   "second write ends return",
			writes: []string{"hello", "world\r"},
			text:   "helloworld\r",
		},
		{
			name:   "third write",
			writes: []string{"hello", "world\r", "hola"},
			text:   "hola",
		},
		{
			name:   "progress one write",
			writes: []string{"progress\n1%\r20%\r100%\n"},
			text:   "progress\n100%\n",
		},
		{
			name:   "progress multiple writes",
			writes: []string{"progress\n", "1%\r", "2%\r", "100%"},
			text:   "progress\n100%",
		},
		{
			name:   "one two three four",
			writes: []string{"one\ntwotwo\nthreethreethree\rfourfourfourfour\n"},
			text:   "one\ntwotwo\nfourfourfourfour\n",
		},
		{
			name:   "real git",
			writes: []string{"Cloning into bare repository '/Users/nick/.sourcegraph/repos/github.com/nicksnyder/go-i18n/.git'...\nremote: Counting objects: 2148, done.        \nReceiving objects:   0% (1/2148)   \rReceiving objects: 100% (2148/2148), 473.65 KiB | 366.00 KiB/s, done.\nResolving deltas:   0% (0/1263)   \rResolving deltas: 100% (1263/1263), done.\n"},
			text:   "Cloning into bare repository '/Users/nick/.sourcegraph/repos/github.com/nicksnyder/go-i18n/.git'...\nremote: Counting objects: 2148, done.        \nReceiving objects: 100% (2148/2148), 473.65 KiB | 366.00 KiB/s, done.\nResolving deltas: 100% (1263/1263), done.\n",
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			var w progressWriter
			for _, write := range testCase.writes {
				_, _ = w.Write([]byte(write))
			}
			assert.Equal(t, testCase.text, w.String())
		})
	}
}
