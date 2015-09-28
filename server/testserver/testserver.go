// +build exectest

package testserver

import (
	"flag"
	"fmt"
	"go/build"
	"io/ioutil"
	"log"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/context"

	"golang.org/x/oauth2"
	"sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/auth"
	"sourcegraph.com/sourcegraph/sourcegraph/auth/accesstoken"
	"sourcegraph.com/sourcegraph/sourcegraph/auth/idkey"
	"sourcegraph.com/sourcegraph/sourcegraph/auth/sharedsecret"
	"sourcegraph.com/sourcegraph/sourcegraph/conf"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/wellknown"
	"sourcegraph.com/sourcegraph/sourcegraph/server/internal/store/fs"
	"sourcegraph.com/sourcegraph/sourcegraph/sgx"
	"sourcegraph.com/sourcegraph/sourcegraph/sgx/sgxcmd"
	storecli "sourcegraph.com/sourcegraph/sourcegraph/store/cli"
	appdashcli "sourcegraph.com/sourcegraph/sourcegraph/util/traceutil/cli"
	"sourcegraph.com/sourcegraph/srclib/flagutil"
)

var (
	verbose  = flag.Bool("testapp.v", false, "verbose output for test app init")
	keepTemp = flag.Bool("testapp.keeptemp", false, "keep temp dirs (do not remove) after exiting")

	// Store must be provided via env vars because otherwise packages
	// that don't use testserver.Server would fail because they lack the
	// CLI flag.
	Store = getTestStore()

	waitServerStart = flag.Duration("testapp.wait", 2*time.Second, "max time to wait for server to start")
)

func getTestStore() string {
	v := os.Getenv("TEST_STORE")
	if v == "" {
		v = "fs"
	}
	return v
}

// Server is a testing helper for integration tests. It lets you spawn an
// external sourcegraph server process optionally letting you
// configure command line flags for the server. It also has helper
// methods for creating CLI commands to perform client operations on
// the corresponding server and HTTP/gRPC clients to interact with the
// server. The server it spawns is self contained, using randomly
// assigned ports and temporary directories that it cleans up after
// Close() is called.
type Server struct {
	Config Config

	SGPATH string

	// Client is an anonymous API client that hits the test Server's API.
	Client *sourcegraph.Client

	// ServerCmd is the exec'd child process subprocess.
	ServerCmd *exec.Cmd

	dbConfig

	Ctx context.Context

	// basePortListener is used to reserve ports. The N ports (where N
	// is the number of args to selectUnusedPorts) after the port that
	// basePortListener listens on are considered reserved for
	// listeners that sgx spawns.
	basePortListener net.Listener
}

type Config struct {
	Flags      []interface{} // flags to `src`
	Endpoints  conf.EndpointOpts
	Serve      sgx.ServeCmd
	ServeFlags []interface{} // flags to `src serve`
}

func (c *Config) args() ([]string, error) {
	flags := c.Flags
	flags = append(flags, &c.Endpoints, "serve", &c.Serve)
	flags = append(flags, c.ServeFlags...)
	return makeCommandLineArgs(flags...)
}

func (s *Server) allEnvConfig() []string {
	v := s.serverEnvConfig()
	if Store == "pgsql" {
		v = append(v, s.dbEnvConfig()...)
	}
	v = append(v, s.srclibEnvConfig()...)
	return v
}

func (s *Server) srclibEnvConfig() []string {
	return []string{
		"SRCLIBPATH=" + getSRCLIBPATHWithSampleToolchain(),

		// Force no Docker so the test doesn't have to worry about
		// building the Docker image for the srclib-sample
		// toolchain.
		"SG_SRCLIB_ENABLE_DOCKER=f",

		"SG_BUILD_LOG_DIR=" + filepath.Join(s.SGPATH, "build-logs"),
	}
}

func (s *Server) serverEnvConfig() []string {
	return []string{
		"SGPATH=" + s.SGPATH,
		"HTTP_DISCOVERY_PORT=" + os.Getenv("HTTP_DISCOVERY_PORT"),
		"HTTP_DISCOVERY_INSECURE=t",
		"DEBUG=t",
	}
}

func (s *Server) CmdAs(login string, args []string) (*exec.Cmd, error) {
	u, err := s.Client.Users.Get(sharedsecret.NewContext(s.Ctx, "tmp"), &sourcegraph.UserSpec{Login: login})
	if err != nil {
		return nil, err
	}
	k := idkey.FromContext(s.Ctx)
	token, err := accesstoken.New(
		k,
		auth.Actor{UID: int(u.UID), ClientID: k.ID},
		map[string]string{"GrantType": "CmdAs"},
		0,
	)
	if err != nil {
		return nil, err
	}
	return s.Cmd([]string{"SRC_TOKEN=" + string(token.AccessToken)}, args), nil
}

func (s *Server) CmdAsSystem(args []string) (*exec.Cmd, error) {
	src := sharedsecret.TokenSource(idkey.FromContext(s.Ctx), "tmp")
	tok, err := src.Token()
	if err != nil {
		return nil, err
	}
	return s.Cmd([]string{"SRC_TOKEN=" + string(tok.AccessToken)}, args), nil
}

// Cmd returns a command that can be executed to perform client
// operations against the server spawned by s.
func (s *Server) Cmd(env []string, args []string) *exec.Cmd {
	configArgs, err := makeCommandLineArgs(
		&s.Config.Endpoints,
	)
	if err != nil {
		log.Fatal(err)
	}

	cmd := exec.Command(sgxcmd.Path)
	cmd.Args = append(cmd.Args, configArgs...)
	cmd.Args = append(cmd.Args, args...)
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	cmd.Env = []string{"USER=" + os.Getenv("USER"), "PATH=" + os.Getenv("PATH"), "HOME=" + os.Getenv("HOME"), "SRCLIBPATH=" + os.Getenv("SRCLIBPATH"), "SRCLIBCACHE=" + os.Getenv("SRCLIBCACHE")}
	cmd.Env = append(cmd.Env, env...)
	cmd.Env = append(cmd.Env, "GOPATH="+os.Getenv("GOPATH")) // for app templates (defaultBase func)
	return cmd
}

func (s *Server) SrclibCmd(env []string, args []string) *exec.Cmd {
	cmd := exec.Command(sgxcmd.Path, "srclib")
	cmd.Args = append(cmd.Args, args...)
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	cmd.Env = []string{"USER=" + os.Getenv("USER"), "PATH=" + os.Getenv("PATH")}
	cmd.Env = append(cmd.Env, s.srclibEnvConfig()...)
	cmd.Env = append(cmd.Env, env...)
	cmd.Env = append(cmd.Env, "GOPATH="+os.Getenv("GOPATH")) // for app templates (defaultBase func)
	return cmd
}

func (s *Server) AsUser(ctx context.Context, login string) (context.Context, error) {
	u, err := s.Client.Users.Get(ctx, &sourcegraph.UserSpec{Login: login})
	if err != nil {
		return nil, err
	}
	return s.AsUID(ctx, int(u.UID)), nil
}

func (s *Server) AsUID(ctx context.Context, uid int) context.Context {
	k := idkey.FromContext(s.Ctx)
	token, err := accesstoken.New(
		k,
		auth.Actor{UID: uid, ClientID: k.ID},
		map[string]string{"GrantType": "AsUID"},
		0,
	)
	if err != nil {
		panic(err)
	}
	return sourcegraph.WithCredentials(ctx, oauth2.StaticTokenSource(token))
}

func (s *Server) Close() {
	if err := s.basePortListener.Close(); err != nil {
		log.Fatal(err)
	}
	if ps := s.ServerCmd.ProcessState; ps != nil && ps.Exited() {
		log.Fatalf("Test server %d exited: %s.", s.ServerCmd.Process.Pid, s.ServerCmd.ProcessState.Sys())
	}
	if err := s.ServerCmd.Process.Signal(os.Interrupt); err != nil {
		log.Fatal(err)
	}
	go func() {
		time.Sleep(1000 * time.Millisecond)
		if err := s.ServerCmd.Process.Kill(); err != nil && !strings.Contains(err.Error(), "process already finished") {
			log.Fatal(err)
		}
	}()
	if _, err := s.ServerCmd.Process.Wait(); err != nil && !strings.Contains(err.Error(), "no child processes") {
		log.Fatal(err)
	}
	if !*keepTemp {
		if err := os.RemoveAll(s.SGPATH); err != nil {
			log.Fatal(err)
		}
	}

	if Store == "pgsql" {
		s.dbConfig.close()
	}
}

func (s *Server) AbsURL(rest string) string {
	return strings.TrimSuffix(s.Config.Serve.AppURL, "/") + rest
}

var portRand = rand.New(rand.NewSource(time.Now().UTC().UnixNano()))

func randomPort() int {
	for {
		port := 10000 + portRand.Intn(40000)
		c, err := net.DialTimeout("tcp4", fmt.Sprintf(":%d", port), time.Millisecond*50)
		if e, ok := err.(net.Error); ok && !e.Temporary() {
			return port
		}
		if err == nil {
			if err := c.Close(); err != nil {
				log.Fatal(err)
			}
		}
	}
}

var (
	selectUnusedPortsMu sync.Mutex
	loPort              = 10000 // never reuse ports, always start at the next highest port after the last used port
)

func (s *Server) selectUnusedPorts(ports ...*int) {
	selectUnusedPortsMu.Lock()
	defer selectUnusedPortsMu.Unlock()

	portRangeIsUnused := func(lo, hi int) bool {
		for p := lo; p <= hi; p++ {
			c, err := net.DialTimeout("tcp", fmt.Sprintf(":%d", p), time.Millisecond*50)
			if e, ok := err.(net.Error); ok && !e.Temporary() {
				continue
			}
			if c != nil {
				if err := c.Close(); err != nil {
					log.Fatal(err)
				}
			}
			return false
		}
		return true
	}

	// To avoid conflicting with other ports that may be chosen by
	// other processes running this same test routine, treat the base
	// port as the owner of the other ports assigned.
	findPortRangeModN := func(n int) int {
		for port := loPort; port < 65535; port++ {
			if port%n != 0 {
				continue
			}
			l, err := net.Listen("tcp", ":"+strconv.Itoa(port))
			if err != nil {
				if strings.Contains(err.Error(), "address already in use") {
					continue
				}
				log.Fatal(err)
			}

			if !portRangeIsUnused(port+1, port+1+len(ports)) {
				if err := l.Close(); err != nil {
					log.Fatal(err)
				}
				continue
			}

			s.basePortListener = l
			return port
		}
		log.Fatalf("Failed to find an unused port to bind on whose number mod %d == 0.", n)
		panic("unreachable")
	}

	basePort := findPortRangeModN(len(ports) + 1)
	loPort = basePort + len(ports) + 2
	for i, port := range ports {
		*port = basePort + i + 1
	}
}

func parseURL(urlStr string) *url.URL {
	url, err := url.Parse(urlStr)
	if err != nil {
		log.Fatal(err)
	}
	return url
}

// NewServer creates a new test application for running integration
// tests. It also has several useful helper methods and fields for
// tests.
func NewServer() (*Server, context.Context) {
	a, ctx := NewUnstartedServer()

	if err := a.Start(); err != nil {
		log.Fatal(err)
	}
	return a, ctx
}

func NewUnstartedServer() (*Server, context.Context) {
	var s Server

	// SGPATH
	sgpath, err := ioutil.TempDir("", "sgtest-sgpath")
	if err != nil {
		log.Fatal(err)
	}
	s.SGPATH = sgpath

	// Find unused ports
	var httpPort, grpcPort, appdashHTTPPort int
	s.selectUnusedPorts(&httpPort, &grpcPort, &appdashHTTPPort)

	// HTTP
	s.Config.Serve.HTTPAddr = fmt.Sprintf(":%d", httpPort)

	// gRPC
	s.Config.Serve.GRPCAddr = fmt.Sprintf(":%d", grpcPort)
	s.Config.Endpoints.GRPCEndpoint = fmt.Sprintf("http://localhost:%d", grpcPort)

	// App
	s.Config.Serve.AppURL = fmt.Sprintf("http://localhost:%d/", httpPort)

	// Store type
	s.Config.ServeFlags = append(s.Config.ServeFlags, &storecli.Flags{
		Store: Store,
	})

	reposDir := filepath.Join(sgpath, "repos")
	if err := os.MkdirAll(reposDir, 0700); err != nil {
		log.Fatal(err)
	}
	buildStoreDir := filepath.Join(sgpath, "buildstore")
	dbDir := filepath.Join(sgpath, "db")
	statusDir := filepath.Join(sgpath, "statuses")

	// FS
	s.Config.ServeFlags = append(s.Config.ServeFlags, &fs.Flags{
		ReposDir:      reposDir,
		BuildStoreDir: buildStoreDir,
		DBDir:         dbDir,
		RepoStatusDir: statusDir,
	})

	// Appdash
	s.Config.ServeFlags = append(s.Config.ServeFlags, &appdashcli.ServerFlags{
		HTTPAddr: fmt.Sprintf(":%d", appdashHTTPPort),
	})
	s.Config.ServeFlags = append(s.Config.ServeFlags, &appdashcli.ClientFlags{
		URL: fmt.Sprintf("http://localhost:%d", appdashHTTPPort),
	})

	// Graphstore
	s.Config.Serve.GraphStoreOpts.Root = reposDir

	// Worker
	s.Config.Serve.WorkCmd = sgx.WorkCmd{
		NumWorkers:  1,
		DequeueMsec: 100,
		BuildRoot:   filepath.Join(sgpath, "builds"),
	}

	s.Ctx = context.Background()
	s.Ctx, err = s.Config.Endpoints.WithEndpoints(s.Ctx)
	if err != nil {
		log.Fatal(err)
	}
	s.Ctx = conf.WithAppURL(s.Ctx, parseURL(s.Config.Serve.AppURL))

	// ID key
	idkey.Bits = 512 // small for testing
	idKey, err := idkey.Generate()
	if err != nil {
		log.Fatal(err)
	}
	s.Config.Serve.IDKeyFile = filepath.Join(sgpath, "id.pem")
	idKeyBytes, err := idKey.MarshalText()
	if err != nil {
		log.Fatal(err)
	}
	if err := ioutil.WriteFile(s.Config.Serve.IDKeyFile, idKeyBytes, 0700); err != nil {
		log.Fatal(err)
	}
	s.Ctx = idkey.NewContext(s.Ctx, idKey)

	s.Ctx = s.AsUID(s.Ctx, 1)

	if Store == "pgsql" {
		if err := s.configDB(); err != nil {
			log.Fatal(err)
		}
	}

	// Server command.
	s.ServerCmd = exec.Command(sgxcmd.Path)
	s.ServerCmd.Stdout = os.Stderr
	s.ServerCmd.Stderr = os.Stderr
	s.ServerCmd.Env = append(bareEnvConfig(), s.allEnvConfig()...)
	//cmd.SysProcA ttr = &syscall.SysProcAttr{Pdeathsig: syscall.SIGINT} // kill child when parent dies

	return &s, s.Ctx
}

func (s *Server) Start() error {
	// Set flags on server cmd.
	sgxArgs, err := s.Config.args()
	if err != nil {
		return err
	}
	s.ServerCmd.Args = append(s.ServerCmd.Args, sgxArgs...)

	if *verbose {
		log.Printf("testapp cmd     = %v", s.ServerCmd.Args)
		log.Printf("testapp cmd.Env = %v", s.allEnvConfig())
	}

	if err := s.ServerCmd.Start(); err != nil {
		return fmt.Errorf("starting server: %s", err)
	}

	go s.ServerCmd.Wait()

	// Wait for server to be ready.
	for start, maxWait := time.Now(), *waitServerStart; ; {
		if ps := s.ServerCmd.ProcessState; ps != nil && ps.Exited() {
			return fmt.Errorf("server PID %d (%s) exited unexpectedly: %v", s.ServerCmd.Process.Pid, s.Config.Serve.AppURL, ps.Sys())
		}
		if time.Since(start) > maxWait {
			return fmt.Errorf("timeout waiting for test server at %s to start (%s)", s.Config.Serve.AppURL, maxWait)
		}
		if resp, err := http.Get(s.Config.Serve.AppURL + wellknown.ConfigPath); err == nil {
			resp.Body.Close()
			break
		}
		time.Sleep(25 * time.Millisecond)
	}
	time.Sleep(75 * time.Millisecond)

	s.Client = sourcegraph.NewClientFromContext(s.Ctx)

	return nil
}

func bareEnvConfig() []string {
	var env []string
	for _, v := range os.Environ() {
		if strings.HasPrefix(v, "PG") || strings.HasPrefix(v, "GITHUB_CLIENT_") || strings.HasPrefix(v, "SRCLIBPATH=") || strings.HasPrefix(v, "SG_SRCLIB_") || strings.HasPrefix(v, "SG_URL") || strings.HasPrefix(v, "SRC_CLIENT_") {
			continue
		}
		env = append(env, v)
	}
	return env
}

// getSRCLIBPATHWithSampleToolchain returns a SRCLIBPATH env var value
// that has only the srclib-sample toolchain installed. Note that it
// doesn't build the toolchain's Docker image, so it must be run as a
// program (not via Docker).
//
// This func assumes that the sample toolchain has been vendored in
// the sourcegraph repo at
// testutil/testdata/srclibpath/sourcegraph.com/sourcegraph/srclib-sample. The
// vendored srclib-sample toolchain can be updated by manually copying
// files from the srclib-sample repo (there's no automatic way to
// update this yet because it's probably an infrequent
// occurrence). Note that you should delete srclib-sample's .git and
// testdata dirs and NOT check them into the sourcegraph repo.
func getSRCLIBPATHWithSampleToolchain() (srclibpath string) {
	p, err := build.Default.Import("sourcegraph.com/sourcegraph/sourcegraph/util/testutil", "", build.FindOnly)
	if err != nil {
		log.Fatal("Couldn't find testutil package (while getting SRCLIBPATH for test app):", err)
	}
	srclibpath = filepath.Join(p.Dir, "testdata/srclibpath")
	sampleDir := filepath.Join(p.Dir, "testdata/srclibpath/sourcegraph.com/sourcegraph/srclib-sample")
	if fi, err := os.Stat(sampleDir); err != nil || !fi.Mode().IsDir() {
		log.Fatalf("Failed to locate srclib-sample dir in SRCLIBPATH for test app: error %v, IsDir=%v", err, fi.Mode().IsDir())
	}
	return srclibpath
}

// makeCommandLineArgs takes a list of EITHER (1) flag structs (like
// sgx.CredentialsOpts) or (2) strings (which denote subcommands) and
// converts them into command-line arguments.
func makeCommandLineArgs(flagsAndSubcommands ...interface{}) ([]string, error) {
	var args []string
	for _, v := range flagsAndSubcommands {
		switch v := v.(type) {
		case string:
			args = append(args, v)
		default:
			optArgs, err := flagutil.MarshalArgs(v)
			if err != nil {
				return nil, err
			}
			args = append(args, optArgs...)
		}
	}
	return args, nil
}
