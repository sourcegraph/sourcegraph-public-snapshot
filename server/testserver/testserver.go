// +build exectest

package testserver

import (
	"flag"
	"fmt"
	"go/build"
	"io/ioutil"
	"log"
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
	"sourcegraph.com/sourcegraph/sourcegraph/cli"
	"sourcegraph.com/sourcegraph/sourcegraph/cli/srccmd"
	"sourcegraph.com/sourcegraph/sourcegraph/go-sourcegraph/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/auth"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/auth/accesstoken"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/auth/idkey"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/auth/sharedsecret"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/conf"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/wellknown"
	appdashcli "sourcegraph.com/sourcegraph/sourcegraph/util/traceutil/cli"
	"sourcegraph.com/sourcegraph/srclib/flagutil"
)

var (
	verbose  = flag.Bool("testapp.v", false, "verbose output for test app init")
	keepTemp = flag.Bool("testapp.keeptemp", false, "keep temp dirs (do not remove) after exiting")

	waitServerStart = flag.Duration("testapp.wait", 10*time.Second, "max time to wait for server to start")
)

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
	// listeners that src spawns.
	basePortListener net.Listener
}

type Config struct {
	Flags      []interface{} // flags to `src`
	Endpoint   cli.EndpointOpts
	Serve      cli.ServeCmd
	ServeFlags []interface{} // flags to `src serve`
}

func (c *Config) args() ([]string, error) {
	flags := c.Flags
	flags = append(flags, &c.Endpoint, "serve", &c.Serve)
	flags = append(flags, c.ServeFlags...)
	return makeCommandLineArgs(flags...)
}

func (s *Server) allEnvConfig() []string {
	v := s.serverEnvConfig()
	v = append(v, s.dbEnvConfig()...)
	v = append(v, s.srclibEnvConfig()...)
	return v
}

func (s *Server) srclibEnvConfig() []string {
	return []string{
		"SG_BUILD_LOG_DIR=" + filepath.Join(s.SGPATH, "build-logs"),
	}
}

func (s *Server) serverEnvConfig() []string {
	return []string{
		"SGPATH=" + s.SGPATH,
		"DEBUG=t",
	}
}

func (s *Server) CmdAs(login string, args []string) (*exec.Cmd, error) {
	u, err := s.Client.Users.Get(sharedsecret.NewContext(s.Ctx, "tmp"), &sourcegraph.UserSpec{Login: login})
	if err != nil {
		return nil, err
	}
	k := idkey.FromContext(s.Ctx)
	token, err := accesstoken.New(k, &auth.Actor{UID: int(u.UID), Write: true}, nil, 10*time.Minute, true)
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
		&s.Config.Endpoint,
	)
	if err != nil {
		log.Fatal(err)
	}

	cmd := exec.Command(srccmd.Path)
	cmd.Args = append(cmd.Args, configArgs...)
	cmd.Args = append(cmd.Args, args...)
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	cmd.Env = []string{"USER=" + os.Getenv("USER"), "PATH=" + os.Getenv("PATH"), "HOME=" + os.Getenv("HOME"), "SRCLIBPATH=" + os.Getenv("SRCLIBPATH"), "SRCLIBCACHE=" + os.Getenv("SRCLIBCACHE"),
		"SRC_AUTH_FILE=/dev/null", // don't heed the local dev user's ~/.src-auth file
	}
	cmd.Env = append(cmd.Env, env...)
	cmd.Env = append(cmd.Env, "GOPATH="+os.Getenv("GOPATH")) // for app templates (defaultBase func)

	if *verbose {
		log.Printf("# test server cmd: %v", cmd.Args)
	}

	return cmd
}

func (s *Server) AsUser(ctx context.Context, login string) (context.Context, error) {
	u, err := s.Client.Users.Get(ctx, &sourcegraph.UserSpec{Login: login})
	if err != nil {
		return nil, err
	}
	return s.AsUIDWithAccess(ctx, int(u.UID), true, false), nil
}

func (s *Server) AsUIDWithAccess(ctx context.Context, uid int, write, admin bool) context.Context {
	k := idkey.FromContext(s.Ctx)
	token, err := accesstoken.New(k, &auth.Actor{UID: uid, Write: write, Admin: admin}, nil, 10*time.Minute, true)
	if err != nil {
		panic(err)
	}
	return sourcegraph.WithCredentials(ctx, oauth2.StaticTokenSource(token))
}

func (s *Server) Close() {
	if err := s.basePortListener.Close(); err != nil {
		log.Fatal(err)
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
		// Because the build workers may still be running, and hence still writing
		// to SGPATH, sometimes RemoveAll can fail, so we just try to delete the
		// directory a few times, waiting for the workers to exit.
		var err error
		for i := 0; i < 10; i++ {
			err = os.RemoveAll(s.SGPATH)
			if err == nil {
				break
			}
			log.Println(err)
			time.Sleep(1 * time.Second)
		}
		if err != nil {
			log.Fatal(err)
		}
	}

	s.dbConfig.close()
}

func (s *Server) AbsURL(rest string) string {
	return strings.TrimSuffix(s.Config.Serve.AppURL, "/") + rest
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

func NewUnstartedServerTLS() (*Server, context.Context) {
	s, ctx := newUnstartedServer("https")

	s.Config.Serve.KeyFile = filepath.Join(s.SGPATH, "localhost.key")
	s.Config.Serve.CertFile = filepath.Join(s.SGPATH, "localhost.crt")
	if err := ioutil.WriteFile(s.Config.Serve.KeyFile, localhostKey, 0600); err != nil {
		log.Fatal(err)
	}
	if err := ioutil.WriteFile(s.Config.Serve.CertFile, localhostCert, 0600); err != nil {
		log.Fatal(err)
	}

	return s, ctx
}

func NewUnstartedServer() (*Server, context.Context) {
	return newUnstartedServer("http")
}

func newUnstartedServer(scheme string) (*Server, context.Context) {
	var s Server

	s.Config.Flags = append(s.Config.Flags, "-v")

	// SGPATH
	sgpath, err := ioutil.TempDir("", "sgtest-sgpath")
	if err != nil {
		log.Fatal(err)
	}
	s.SGPATH = sgpath

	// Find unused ports
	var httpPort, httpsPort int
	s.selectUnusedPorts(&httpPort, &httpsPort)

	var mainHTTPPort int
	switch scheme {
	case "http":
		mainHTTPPort = httpPort
	case "https":
		mainHTTPPort = httpsPort
	default:
		panic("bad scheme: " + scheme)
	}

	// HTTP
	s.Config.Serve.HTTPAddr = fmt.Sprintf(":%d", httpPort)
	s.Config.Serve.HTTPSAddr = fmt.Sprintf(":%d", httpsPort)
	s.Config.Endpoint.URL = fmt.Sprintf("%s://localhost:%d", scheme, mainHTTPPort)

	// Other config
	s.Config.Serve.NoInitialOnboarding = true
	s.Config.Serve.RegisterURL = ""

	// App
	s.Config.Serve.AppURL = fmt.Sprintf("%s://localhost:%d/", scheme, mainHTTPPort)

	reposDir := filepath.Join(sgpath, "repos")
	if err := os.MkdirAll(reposDir, 0700); err != nil {
		log.Fatal(err)
	}

	// FS
	s.Config.Serve.ReposDir = reposDir

	// Disable Appdash, because it depends on InfluxDB we would have to use
	// random ports for all InfluxDB services (influxd, admin, collectd,
	// graphite, httpd, opentsdb, and udp services) and also use random influx
	// directories for both influxdb and it's meta service, otherwise we could
	// not avoid race conditions between parralel tests. Not worth it.
	s.Config.ServeFlags = append(s.Config.ServeFlags, &appdashcli.ServerConfig{
		Disable: true,
	})

	// Graphstore
	s.Config.Serve.GraphStoreOpts.Root = reposDir

	// Worker
	s.Config.Serve.WorkCmd = cli.WorkCmd{
		DequeueMsec: 100,
		Parallel:    2,
	}

	s.Ctx = context.Background()
	s.Ctx = sourcegraph.WithGRPCEndpoint(s.Ctx, s.Config.Endpoint.URLOrDefault())
	s.Ctx = conf.WithURL(s.Ctx, parseURL(s.Config.Serve.AppURL))

	// ID key
	idkey.SetTestEnvironment(1024) // Minimum RSA size for SSH is 1024
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

	s.Ctx = s.AsUIDWithAccess(s.Ctx, 1, false, true)

	if err := s.configDB(); err != nil {
		log.Fatal(err)
	}

	// Server command.
	s.ServerCmd = exec.Command(srccmd.Path)
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

	cmdFinished := make(chan bool, 1)
	go func() {
		s.ServerCmd.Wait()
		cmdFinished <- true
	}()

	// Wait for server to be ready.
	for start, maxWait := time.Now(), *waitServerStart; ; {
		select {
		case <-cmdFinished:
			if ps := s.ServerCmd.ProcessState; ps != nil && ps.Exited() {
				return fmt.Errorf("server PID %d (%s) exited unexpectedly: %v", s.ServerCmd.Process.Pid, s.Config.Serve.AppURL, ps.Sys())
			}
		default:
			// busy wait
		}
		if time.Since(start) > maxWait {
			s.Close()
			return fmt.Errorf("timeout waiting for test server at %s to start (%s)", s.Config.Serve.AppURL, maxWait)
		}
		if resp, err := http.Get(s.Config.Serve.AppURL + wellknown.ConfigPath); err == nil {
			resp.Body.Close()
			break
		}
		time.Sleep(25 * time.Millisecond)
	}
	time.Sleep(75 * time.Millisecond)

	s.Client, _ = sourcegraph.NewClientFromContext(s.Ctx)

	return nil
}

func bareEnvConfig() []string {
	var env []string
	for _, v := range os.Environ() {
		if strings.HasPrefix(v, "SGPATH") || strings.HasPrefix(v, "PG") ||
			strings.HasPrefix(v, "GITHUB_CLIENT_") || strings.HasPrefix(v, "SRCLIBPATH=") ||
			strings.HasPrefix(v, "SG_SRCLIB_") || strings.HasPrefix(v, "SG_URL") ||
			strings.HasPrefix(v, "SRC_CLIENT_") {
			continue
		}
		env = append(env, v)
	}
	return env
}

// SrclibSampleToolchain returns the dir and the SRCLIBPATH for the
// included srclib-sample toolchain.
//
// If buildDocker is true, its corresponding Docker image is also
// built. The dockerImage is the Docker image tag to use, and it is
// only returned if buildDocker is true.
//
// This func assumes that the sample toolchain has been vendored in
// the sourcegraph repo at
// testutil/testdata/srclibpath/sourcegraph.com/sourcegraph/srclib-sample. The
// vendored srclib-sample toolchain can be updated by manually copying
// files from the srclib-sample repo (there's no automatic way to
// update this yet because it's probably an infrequent
// occurrence). Note that you should delete srclib-sample's .git and
// testdata dirs and NOT check them into the sourcegraph repo.
func SrclibSampleToolchain(buildDocker bool) (dir, srclibpath, dockerImage string) {
	p, err := build.Default.Import("sourcegraph.com/sourcegraph/sourcegraph/util/testutil", "", build.FindOnly)
	if err != nil {
		log.Fatal("Couldn't find testutil package (while getting SRCLIBPATH for test app):", err)
	}
	srclibpath = filepath.Join(p.Dir, "testdata/srclibpath")
	dir = filepath.Join(p.Dir, "testdata/srclibpath/sourcegraph.com/sourcegraph/srclib-sample")
	if fi, err := os.Stat(dir); err != nil || !fi.Mode().IsDir() {
		log.Fatalf("Failed to locate srclib-sample dir in SRCLIBPATH for test app: error %v, IsDir=%v", err, fi.Mode().IsDir())
	}

	if buildDocker {
		dockerImage = "sourcegraph-test/srclib-sample"
		cmd := exec.Command("docker", "build", "-t", dockerImage, ".")
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			log.Fatalf("Error building srclib sample toolchain in %s.\n\n%s", dir, out)
		}
	}

	return
}

// makeCommandLineArgs takes a list of EITHER (1) flag structs (like
// client.CredentialsOpts) or (2) strings (which denote subcommands) and
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

// localhostCert is a PEM-encoded TLS cert with SAN IPs
// "127.0.0.1" and "[::1]", expiring at the last second of 2049 (the end
// of ASN.1 time).
// generated from src/crypto/tls:
// go run generate_cert.go  --rsa-bits 1024 --host 127.0.0.1,::1,example.com --ca --start-date "Jan 1 00:00:00 1970" --duration=1000000h
var localhostCert = []byte(`-----BEGIN CERTIFICATE-----
MIICEzCCAXygAwIBAgIQMIMChMLGrR+QvmQvpwAU6zANBgkqhkiG9w0BAQsFADAS
MRAwDgYDVQQKEwdBY21lIENvMCAXDTcwMDEwMTAwMDAwMFoYDzIwODQwMTI5MTYw
MDAwWjASMRAwDgYDVQQKEwdBY21lIENvMIGfMA0GCSqGSIb3DQEBAQUAA4GNADCB
iQKBgQDuLnQAI3mDgey3VBzWnB2L39JUU4txjeVE6myuDqkM/uGlfjb9SjY1bIw4
iA5sBBZzHi3z0h1YV8QPuxEbi4nW91IJm2gsvvZhIrCHS3l6afab4pZBl2+XsDul
rKBxKKtD1rGxlG4LjncdabFn9gvLZad2bSysqz/qTAUStTvqJQIDAQABo2gwZjAO
BgNVHQ8BAf8EBAMCAqQwEwYDVR0lBAwwCgYIKwYBBQUHAwEwDwYDVR0TAQH/BAUw
AwEB/zAuBgNVHREEJzAlggtleGFtcGxlLmNvbYcEfwAAAYcQAAAAAAAAAAAAAAAA
AAAAATANBgkqhkiG9w0BAQsFAAOBgQCEcetwO59EWk7WiJsG4x8SY+UIAA+flUI9
tyC4lNhbcF2Idq9greZwbYCqTTTr2XiRNSMLCOjKyI7ukPoPjo16ocHj+P3vZGfs
h1fIw3cSS2OolhloGw/XM6RWPWtPAlGykKLciQrBru5NAPvCMsb/I1DAceTiotQM
fblo6RBxUQ==
-----END CERTIFICATE-----`)

// localhostKey is the private key for localhostCert.
var localhostKey = []byte(`-----BEGIN RSA PRIVATE KEY-----
MIICXgIBAAKBgQDuLnQAI3mDgey3VBzWnB2L39JUU4txjeVE6myuDqkM/uGlfjb9
SjY1bIw4iA5sBBZzHi3z0h1YV8QPuxEbi4nW91IJm2gsvvZhIrCHS3l6afab4pZB
l2+XsDulrKBxKKtD1rGxlG4LjncdabFn9gvLZad2bSysqz/qTAUStTvqJQIDAQAB
AoGAGRzwwir7XvBOAy5tM/uV6e+Zf6anZzus1s1Y1ClbjbE6HXbnWWF/wbZGOpet
3Zm4vD6MXc7jpTLryzTQIvVdfQbRc6+MUVeLKwZatTXtdZrhu+Jk7hx0nTPy8Jcb
uJqFk541aEw+mMogY/xEcfbWd6IOkp+4xqjlFLBEDytgbIECQQDvH/E6nk+hgN4H
qzzVtxxr397vWrjrIgPbJpQvBsafG7b0dA4AFjwVbFLmQcj2PprIMmPcQrooz8vp
jy4SHEg1AkEA/v13/5M47K9vCxmb8QeD/asydfsgS5TeuNi8DoUBEmiSJwma7FXY
fFUtxuvL7XvjwjN5B30pNEbc6Iuyt7y4MQJBAIt21su4b3sjXNueLKH85Q+phy2U
fQtuUE9txblTu14q3N7gHRZB4ZMhFYyDy8CKrN2cPg/Fvyt0Xlp/DoCzjA0CQQDU
y2ptGsuSmgUtWj3NM9xuwYPm+Z/F84K6+ARYiZ6PYj013sovGKUFfYAqVXVlxtIX
qyUBnu3X9ps8ZfjLZO7BAkEAlT4R5Yl6cGhaJQYZHOde3JEMhNRcVFMO8dJDaFeo
f9Oeos0UUothgiDktdQHxdNEwLjQf7lJJBzV+5OtwswCWA==
-----END RSA PRIVATE KEY-----`)
