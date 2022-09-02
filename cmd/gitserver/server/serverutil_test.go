package server

import (
	"net/http/httptest"
	"os"
	"os/exec"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/schema"
)

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
			conf := test.tlsConfig
			if conf == nil {
				conf = &tlsConfig{}
			}
			configureRemoteGitCommand(test.input, conf)
			if !reflect.DeepEqual(test.input.Env, test.expectedEnv) {
				t.Errorf("\ngot:  %s\nwant: %s\n", test.input.Env, test.expectedEnv)
			}
			if !reflect.DeepEqual(test.input.Args, test.expectedArgs) {
				t.Errorf("\ngot:  %s\nwant: %s\n", test.input.Args, test.expectedArgs)
			}
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
		if !reflect.DeepEqual(cmd.Env, want) {
			t.Errorf("mismatch for %#+v (-want +got):\n%s", tc.conf, cmp.Diff(want, cmd.Env))
		}
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
			if actual := w.String(); testCase.text != actual {
				t.Fatalf("\ngot:\n%s\nwant:\n%s\n", actual, testCase.text)
			}
		})
	}
}

func TestFlushingResponseWriter(t *testing.T) {
	flush := make(chan struct{})
	fw := &flushingResponseWriter{
		w: httptest.NewRecorder(),
		flusher: flushFunc(func() {
			flush <- struct{}{}
		}),
	}
	done := make(chan struct{})
	go func() {
		fw.periodicFlush()
		close(done)
	}()

	_, _ = fw.Write([]byte("hi"))

	select {
	case <-flush:
		close(flush)
	case <-time.After(5 * time.Second):
		t.Fatal("periodic flush did not happen")
	}

	fw.Close()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("periodic flush goroutine did not close")
	}
}

type flushFunc func()

func (f flushFunc) Flush() {
	f()
}
