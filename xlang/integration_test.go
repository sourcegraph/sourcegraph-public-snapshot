package xlang_test

import (
	"context"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/sourcegraph/ctxvfs"
	"github.com/sourcegraph/go-langserver/pkg/lsp"
	"sourcegraph.com/sourcegraph/sourcegraph/xlang"
	gobuildserver "sourcegraph.com/sourcegraph/sourcegraph/xlang/golang/buildserver"
	"sourcegraph.com/sourcegraph/sourcegraph/xlang/uri"
	"sourcegraph.com/sourcegraph/sourcegraph/xlang/vfsutil"
)

func TestIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skip long integration test")
	}

	tests := map[string]struct { // map key is rootPath
		mode             string
		pinDepReposToRev map[string]string // so that file:line:col expectations are stable
		wantHover        map[string]string
		wantDefinition   map[string]string
		wantReferences   map[string][]string
		wantSymbols      map[string][]string
	}{
		"git://github.com/gorilla/mux?0a192a193177452756c362c20087ddafcf6829c4": {
			mode: "go",
			pinDepReposToRev: map[string]string{
				"https://github.com/gorilla/context": "08b5f424b9271eedf6f9f0ce86cb9396ed337a42",
			},
			wantHover: map[string]string{
				"mux.go:61:38": "type Request struct", // stdlib
			},
			wantDefinition: map[string]string{
				"mux.go:61:38": "git://github.com/golang/go?go1.7.1#src/net/http/request.go:76:6", // stdlib
			},
		},
		"git://github.com/coreos/fuze?7df4f06041d9daba45e4c68221b9b04203dff1d8": {
			mode: "go",
			pinDepReposToRev: map[string]string{
				"https://github.com/stretchr/testify": "976c720a22c8eb4eb6a0b4348ad85ad12491a506",
				"https://gopkg.in/check.v1.git":       "4f90aeace3a26ad7021961c297b22c42160c7b25",
				"https://gopkg.in/yaml.v2.git":        "a5b47d31c556af34a302ce5d659e6fea44d90de0",
			},
			wantHover: map[string]string{
				"config/convert.go:262:26": "func ParseBase2Bytes(s string) (Base2Bytes, error)", // vendored
			},
			wantDefinition: map[string]string{
				"config/convert.go:262:26": "git://github.com/coreos/fuze?7df4f06041d9daba45e4c68221b9b04203dff1d8#config/vendor/github.com/alecthomas/units/bytes.go:30:6", // vendored TODO(sqs): really want the below result which has the non-vendored path as well, need to implement that
				//"config/convert.go:262:26": "git://github.com/coreos/fuze?7df4f06041d9daba45e4c68221b9b04203dff1d8#config/vendor/github.com/alecthomas/units/bytes.go:30:6 git://github.com/alecthomas/units#bytes.go:30:6", // vendored
			},
		},
		"git://github.com/golang/lint?c7bacac2b21ca01afa1dee0acf64df3ce047c28f": {
			mode: "go",
			pinDepReposToRev: map[string]string{
				"https://github.com/golang/tools": "73d2e795b859a48cba2d70040c384dd1cea7e113",
			},
			wantHover: map[string]string{
				"golint/golint.go:91:18": "type Linter struct", // diff pkg, same repo
			},
			wantDefinition: map[string]string{
				"golint/golint.go:91:18": "git://github.com/golang/lint?c7bacac2b21ca01afa1dee0acf64df3ce047c28f#lint.go:31:6", // diff pkg, same repo
			},
		},
		"git://github.com/gorilla/csrf?a8abe8abf66db8f4a9750d76ba95b4021a354757": {
			mode: "go",
			pinDepReposToRev: map[string]string{
				"https://github.com/gorilla/securecookie": "c13558c2b1c44da35e0eb043053609a5ba3a1f19",
				"https://github.com/gorilla/context":      "08b5f424b9271eedf6f9f0ce86cb9396ed337a42",
				"https://github.com/pkg/errors":           "839d9e913e063e28dfd0e6c7b7512793e0a48be9",
			},
			wantHover: map[string]string{
				"csrf.go:57:28": "type SecureCookie struct", // diff repo
			},
			wantDefinition: map[string]string{
				"csrf.go:57:28": "git://github.com/gorilla/securecookie?HEAD#securecookie.go:154:6", // diff repo
			},
		},
		"git://github.com/golang/go?f75aafdf56dd90eab75cfeac8cf69358f73ba171": {
			// SHA is equivalent to go1.7.1 tag, but make sure we
			// retain the original rev spec in definition results.
			mode: "go",
			wantHover: map[string]string{
				"src/encoding/hex/hex.go:70:12":  "func fromHexChar(c byte) (byte, bool)", // func decl
				"src/encoding/hex/hex.go:104:18": "type Buffer struct",                    // bytes.Buffer
				"src/net/http/server.go:78:32":   "type Request struct",
			},
			wantDefinition: map[string]string{
				"src/encoding/hex/hex.go:70:12":  "git://github.com/golang/go?f75aafdf56dd90eab75cfeac8cf69358f73ba171#src/encoding/hex/hex.go:70:6", // func decl
				"src/encoding/hex/hex.go:104:18": "git://github.com/golang/go?f75aafdf56dd90eab75cfeac8cf69358f73ba171#src/bytes/buffer.go:17:6",     // stdlib type
			},
			wantReferences: map[string][]string{
				"src/net/http/httptest/server.go:204:25": []string{ // httptest.Server
					"git://github.com/golang/go?f75aafdf56dd90eab75cfeac8cf69358f73ba171#src/net/http/httptest/server.go:204:18",
					"git://github.com/golang/go?f75aafdf56dd90eab75cfeac8cf69358f73ba171#src/net/http/httptest/server_test.go:92:5",
					"git://github.com/golang/go?f75aafdf56dd90eab75cfeac8cf69358f73ba171#src/net/http/serve_test.go:2625:7",
					"git://github.com/golang/go?f75aafdf56dd90eab75cfeac8cf69358f73ba171#src/net/http/transport_test.go:2553:6",
					"git://github.com/golang/go?f75aafdf56dd90eab75cfeac8cf69358f73ba171#src/net/http/transport_test.go:478:5",
					"git://github.com/golang/go?f75aafdf56dd90eab75cfeac8cf69358f73ba171#src/net/http/transport_test.go:532:5",
				},
			},
			wantSymbols: map[string][]string{
				"Sum256": []string{"git://github.com/golang/go?f75aafdf56dd90eab75cfeac8cf69358f73ba171#src/crypto/sha256/sha256.go:function:sha256.Sum256:176:5"},
			},
		},
		"git://github.com/docker/machine?e1a03348ad83d8e8adb19d696bc7bcfb18ccd770": {
			mode: "go",
			wantHover: map[string]string{
				"libmachine/provision/provisioner.go:107:50": "func RunSSHCommandFromDriver(...",
			},
			wantDefinition: map[string]string{
				"libmachine/provision/provisioner.go:107:50": "git://github.com/docker/machine?e1a03348ad83d8e8adb19d696bc7bcfb18ccd770#libmachine/drivers/utils.go:36:6",
			},
		},
	}
	for rootPath, test := range tests {
		label := strings.TrimPrefix(strings.Replace(strings.Replace(rootPath, "//", "", 1), "/", "-", -1), "git:") // abbreviated label
		t.Run(label, func(t *testing.T) {
			if os.Getenv("CI") != "" && strings.Contains(rootPath, "github.com/golang/go") {
				t.Skip("Skipping the Go stdlib integration test in CI; it exceeds the available memory (4 GB) and fails the whole build.")
			}

			{
				// Serve repository data from codeload.github.com for
				// test performance instead of from gitserver. This
				// technically means we aren't testing gitserver, but
				// that is well tested separately, and the benefit of
				// fast tests here outweighs the benefits of a coarser
				// integration test.
				orig := xlang.NewRemoteRepoVFS
				xlang.NewRemoteRepoVFS = func(cloneURL *url.URL, rev string) (ctxvfs.FileSystem, error) {
					fullName := cloneURL.Host + strings.TrimSuffix(cloneURL.Path, ".git") // of the form "github.com/foo/bar"
					return vfsutil.NewGitHubRepoVFS(fullName, rev, "", true)
				}
				defer func() {
					xlang.NewRemoteRepoVFS = orig
				}()
			}
			{
				// If integration tests depend on external repos, we
				// need to use a pinned, hardcoded revision instead of
				// "HEAD", or else any file:line:col expectations we
				// have will break if the dep repo's files change.
				orig := gobuildserver.NewDepRepoVFS
				gobuildserver.NewDepRepoVFS = func(cloneURL *url.URL, rev string) (ctxvfs.FileSystem, error) {
					if pinRev, ok := test.pinDepReposToRev[cloneURL.String()]; ok {
						rev = pinRev
					} else if len(rev) != 40 && rev != "go1.7.1" {
						// It's OK to hardcode allowable Git tags
						// (such as "goN.N.N") here, since we know
						// those to be stable. Branches like "master"
						// are not stable and are not OK to hardcode
						// here.
						t.Fatalf("must specify pinDepReposToRev in integration test definition so that test analysis is deterministic/stable (and not dependent on the mutable git rev spec %q for repo %q)", rev, cloneURL)
					}
					return orig(cloneURL, rev)
				}
				defer func() {
					gobuildserver.NewDepRepoVFS = orig
				}()
			}

			ctx := context.Background()
			proxy := xlang.NewProxy()
			addr, done := startProxy(t, proxy)
			defer done()
			c := dialProxy(t, addr, nil)

			// Prepare the connection.
			if err := c.Call(ctx, "initialize", xlang.ClientProxyInitializeParams{
				InitializeParams: lsp.InitializeParams{RootPath: rootPath},
				Mode:             test.mode,
			}, nil); err != nil {
				t.Fatal("initialize:", err)
			}

			root, err := uri.Parse(rootPath)
			if err != nil {
				t.Fatal(err)
			}

			lspTests(t, ctx, c, root, test.wantHover, test.wantDefinition, test.wantReferences, test.wantSymbols)

			if err := c.Close(); err != nil {
				t.Fatal(err)
			}
		})
	}
}
