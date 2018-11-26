package server_test

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/pkg/errors"
	"github.com/sourcegraph/ctxvfs"
	"github.com/sourcegraph/go-langserver/pkg/lsp"
	lsext "github.com/sourcegraph/go-langserver/pkg/lspext"
	"github.com/sourcegraph/go-lsp/lspext"
	gobuildserver "github.com/sourcegraph/sourcegraph/enterprise/cmd/xlang-go/internal/server"
	"github.com/sourcegraph/sourcegraph/pkg/gituri"
	"github.com/sourcegraph/sourcegraph/pkg/gosrc"
	"github.com/sourcegraph/sourcegraph/pkg/vfsutil"
)

func init() {
	gosrc.RuntimeVersion = "go1.7.1"
}

func TestIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skip long integration test")
	}

	tests := map[lsp.DocumentURI]struct { // map key is rootURI
		mode              string
		ciBlacklist       bool
		pinDepReposToRev  map[string]string // so that file:line:col expectations are stable
		wantHover         map[string]string
		wantDefinition    map[string]string
		wantXDefinition   map[string]string
		wantReferences    map[string][]string
		wantSymbols       map[string][]string
		wantXDependencies string
		wantXReferences   map[*lsext.WorkspaceReferencesParams][]string
		wantXPackages     []string
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
			wantXDefinition: map[string]string{
				"mux.go:61:38": "git://github.com/golang/go?go1.7.1#src/net/http/request.go:76:6 id:net/http/-/Request name:Request package:net/http packageName:http recv: vendor:false",
			},
			wantXDependencies: "gorilla-mux.json",
			wantXPackages:     []string{"github.com/gorilla/mux"},
		},
		"git://github.com/coreos/fuze?7df4f06041d9daba45e4c68221b9b04203dff1d8": {
			mode: "go",
			pinDepReposToRev: map[string]string{
				"https://github.com/stretchr/testify": "976c720a22c8eb4eb6a0b4348ad85ad12491a506",
				"https://github.com/go-check/check":   "4f90aeace3a26ad7021961c297b22c42160c7b25",
				"https://github.com/go-yaml/yaml":     "a5b47d31c556af34a302ce5d659e6fea44d90de0",
			},
			wantHover: map[string]string{
				"config/convert.go:262:26": "func ParseBase2Bytes(s string) (Base2Bytes, error)", // vendored
				"config/vendor/github.com/coreos/ignition/config/vendor/github.com/coreos/go-semver/semver/semver_test.go:287:27": "func Marshal(in interface{}) (out []byte, err error)",
			},
			wantDefinition: map[string]string{
				"config/convert.go:262:26": "git://github.com/coreos/fuze?7df4f06041d9daba45e4c68221b9b04203dff1d8#config/vendor/github.com/alecthomas/units/bytes.go:30:6", // vendored TODO(sqs): really want the below result which has the non-vendored path as well, need to implement that
				//"config/convert.go:262:26": "git://github.com/coreos/fuze?7df4f06041d9daba45e4c68221b9b04203dff1d8#config/vendor/github.com/alecthomas/units/bytes.go:30:6 git://github.com/alecthomas/units#bytes.go:30:6", // vendored

				"config/vendor/github.com/coreos/ignition/config/vendor/github.com/coreos/go-semver/semver/semver_test.go:287:27": "git://github.com/go-yaml/yaml?v2.2.1#yaml.go:138:6", // diff repo
			},
			wantXDefinition: map[string]string{
				"config/convert.go:262:26": "git://github.com/coreos/fuze?7df4f06041d9daba45e4c68221b9b04203dff1d8#config/vendor/github.com/alecthomas/units/bytes.go:30:6 id:github.com/coreos/fuze/config/vendor/github.com/alecthomas/units/-/ParseBase2Bytes name:ParseBase2Bytes package:github.com/coreos/fuze/config/vendor/github.com/alecthomas/units packageName:units recv: vendor:true",
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
			wantXDefinition: map[string]string{
				"golint/golint.go:91:18": "git://github.com/golang/lint?c7bacac2b21ca01afa1dee0acf64df3ce047c28f#lint.go:31:6 id:github.com/golang/lint/-/Linter name:Linter package:github.com/golang/lint packageName:lint recv: vendor:false",
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
			wantXDefinition: map[string]string{
				"csrf.go:57:28": "git://github.com/gorilla/securecookie?HEAD#securecookie.go:154:6 id:github.com/gorilla/securecookie/-/SecureCookie name:SecureCookie package:github.com/gorilla/securecookie packageName:securecookie recv: vendor:false",
			},
			wantXDependencies: "gorilla-csrf.json",
		},
		"git://github.com/golang/go?f75aafdf56dd90eab75cfeac8cf69358f73ba171": {
			// SHA is equivalent to go1.7.1 tag, but make sure we
			// retain the original rev spec in definition results.
			mode:        "go",
			ciBlacklist: true, // skip on CI since the repo is large
			wantHover: map[string]string{
				"src/encoding/hex/hex.go:70:12":  "func fromHexChar(c byte) (byte, bool)", // func decl
				"src/encoding/hex/hex.go:104:18": "type Buffer struct",                    // bytes.Buffer
				"src/net/http/server.go:78:32":   "type Request struct",
			},
			wantDefinition: map[string]string{
				"src/encoding/hex/hex.go:70:12":  "git://github.com/golang/go?f75aafdf56dd90eab75cfeac8cf69358f73ba171#src/encoding/hex/hex.go:70:6", // func decl
				"src/encoding/hex/hex.go:104:18": "git://github.com/golang/go?f75aafdf56dd90eab75cfeac8cf69358f73ba171#src/bytes/buffer.go:17:6",     // stdlib type
			},
			wantXDefinition: map[string]string{
				"src/encoding/hex/hex.go:70:12":  "git://github.com/golang/go?f75aafdf56dd90eab75cfeac8cf69358f73ba171#src/encoding/hex/hex.go:70:6 id:encoding/hex/-/fromHexChar name:fromHexChar package:encoding/hex packageName:hex recv: vendor:false",
				"src/encoding/hex/hex.go:104:18": "git://github.com/golang/go?f75aafdf56dd90eab75cfeac8cf69358f73ba171#src/bytes/buffer.go:17:6 id:bytes/-/Buffer name:Buffer package:bytes packageName:bytes recv: vendor:false",
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
				"Sum256":                       []string{"git://github.com/golang/go?f75aafdf56dd90eab75cfeac8cf69358f73ba171#src/crypto/sha256/sha256.go:function:sha256.Sum256:176:5"},
				"dir:src/crypto/sha256 Sum256": []string{"git://github.com/golang/go?f75aafdf56dd90eab75cfeac8cf69358f73ba171#src/crypto/sha256/sha256.go:function:sha256.Sum256:176:5"},
				"dir:crypto/sha256 Sum256":     []string{}, // invalid dir
				"dir:foo Sum256":               []string{}, // invalid dir
			},
			wantXDependencies: "golang-go.json",
		},
		"git://github.com/docker/machine?e1a03348ad83d8e8adb19d696bc7bcfb18ccd770": {
			mode:        "go",
			ciBlacklist: true, // skip on CI due to large repo size
			wantHover: map[string]string{
				"libmachine/provision/provisioner.go:107:50": "func RunSSHCommandFromDriver(...",
			},
			wantDefinition: map[string]string{
				"libmachine/provision/provisioner.go:107:50": "git://github.com/docker/machine?e1a03348ad83d8e8adb19d696bc7bcfb18ccd770#libmachine/drivers/utils.go:36:6",
			},
			wantXDefinition: map[string]string{
				"libmachine/provision/provisioner.go:107:50": "git://github.com/docker/machine?e1a03348ad83d8e8adb19d696bc7bcfb18ccd770#libmachine/drivers/utils.go:36:6 id:github.com/docker/machine/libmachine/drivers/-/RunSSHCommandFromDriver name:RunSSHCommandFromDriver package:github.com/docker/machine/libmachine/drivers packageName:drivers recv: vendor:false",
			},
		},
		"git://github.com/kubernetes/kubernetes?c41c24fbf300cd7ba504ea1ac2e052c4a1bbed33": {
			mode:        "go",
			ciBlacklist: true, // skip on CI due to large repo size
			pinDepReposToRev: map[string]string{
				"https://github.com/kubernetes/client-go": "5fe6fc56cb38d04ef4af601a03599c984229dea2",
			},
			wantHover: map[string]string{
				"pkg/ssh/ssh.go:49:38":               "func NewCounter(...",
				"pkg/util/workqueue/queue.go:113:15": "struct field L sync.Locker",
			},
			wantSymbols: map[string][]string{
				"kubectlAnn": []string{"git://github.com/kubernetes/kubernetes?c41c24fbf300cd7ba504ea1ac2e052c4a1bbed33#pkg/kubectl/kubectl.go:constant:kubectlAnnotationPrefix:31:1"},
			},
			wantXDependencies: "kubernetes-kubernetes.json",
		},
		"git://github.com/uber-go/atomic?3b8db5e93c4c02efbc313e17b2e796b0914a01fb": {
			mode: "go",
			wantDefinition: map[string]string{
				// glide.lock specifies testify to something other than HEAD
				"atomic_test.go:32:12": "git://github.com/stretchr/testify?d77da356e56a7428ad25149ca77381849a6a5232#require/require.go:58:6",
			},
		},
		"git://github.com/sgtest/godep-include?d92076664c875c0134dbd475b81f88d97df2bc41": {
			mode: "go",
			wantDefinition: map[string]string{
				// Godeps.json specifies testify to something other than HEAD
				"foo.go:12:12": "git://github.com/stretchr/testify?d77da356e56a7428ad25149ca77381849a6a5232#require/require.go:58:6",
			},
		},
	}
	for rootURI, test := range tests {
		root, err := gituri.Parse(string(rootURI))
		if err != nil {
			t.Fatal(err)
		}
		label := strings.Replace(strings.TrimPrefix(root.Path, "/"), "/", "-", -1)
		t.Run(label, func(t *testing.T) {
			if os.Getenv("CI") != "" && test.ciBlacklist {
				t.Skipf("Skipping the %s integration test in CI", rootURI)
			}

			cleanup := useGithubForVFS()
			defer cleanup()

			{
				// If integration tests depend on external repos, we
				// need to use a pinned, hardcoded revision instead of
				// "HEAD", or else any file:line:col expectations we
				// have will break if the dep repo's files change.
				orig := gobuildserver.NewDepRepoVFS
				gobuildserver.NewDepRepoVFS = func(ctx context.Context, cloneURL *url.URL, rev string) (ctxvfs.FileSystem, error) {
					if pinRev, ok := test.pinDepReposToRev[cloneURL.String()]; ok {
						rev = pinRev
					} else if len(rev) != 40 && rev != gosrc.RuntimeVersion {
						// It's OK to hardcode allowable Git tags
						// (such as "goN.N.N") here, since we know
						// those to be stable. Branches like "master"
						// are not stable and are not OK to hardcode
						// here.
						// We panic since t.Fatal does not interact nicely with subtests
						panic(fmt.Sprintf("TestIntegration/%s: must specify pinDepReposToRev in integration test definition so that test analysis is deterministic/stable (and not dependent on the mutable git rev spec %q for repo %q)", label, rev, cloneURL))
					}
					return orig(ctx, cloneURL, rev)
				}
				defer func() {
					gobuildserver.NewDepRepoVFS = orig
				}()
			}

			ctx := context.Background()

			c, done := connectionToNewBuildServer(string(rootURI), t)
			defer done()

			// Prepare the connection.
			if err := c.Call(ctx, "initialize", lspext.InitializeParams{
				InitializeParams: lsp.InitializeParams{RootURI: "file:///"},
				OriginalRootURI:  rootURI,
			}, nil); err != nil {
				t.Fatal("initialize:", err, rootURI)
			}

			root, err := gituri.Parse(string(rootURI))
			if err != nil {
				t.Fatal(err)
			}

			lspTests(t, ctx, c, root, test.wantHover, test.wantDefinition, test.wantXDefinition, test.wantReferences, test.wantSymbols, test.wantXDependencies, test.wantXReferences, test.wantXPackages)

			if err := c.Close(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

// useGitHubForVFS allows us to serve repository data from codeload.github.com
// for test performance instead of from gitserver. This technically means we
// aren't testing gitserver, but that is well tested separately, and the
// benefit of fast tests here outweighs the benefits of a coarser integration
// test.
func useGithubForVFS() func() {
	origRemoteFS := gobuildserver.RemoteFS
	gobuildserver.RemoteFS = func(ctx context.Context, initializeParams lspext.InitializeParams) (ctxvfs.FileSystem, error) {
		u, err := gituri.Parse(string(initializeParams.OriginalRootURI))
		if err != nil {
			return nil, errors.Wrap(err, "could not parse workspace URI for remotefs")
		}
		if u.Rev() == "" {
			return nil, errors.Errorf("rev is required in uri: %s", initializeParams.OriginalRootURI)
		}
		return vfsutil.NewGitHubRepoVFS(string(u.Repo()), u.Rev())
	}

	return func() {
		gobuildserver.RemoteFS = origRemoteFS
	}
}
