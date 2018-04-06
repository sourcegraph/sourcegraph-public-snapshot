package langservers

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"sort"
	"strings"
	"sync"
	"syscall"

	log15 "gopkg.in/inconshreveable/log15.v2"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/app/envvar"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/goroutine"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/conf"
	"sourcegraph.com/sourcegraph/sourcegraph/schema"
)

// Languages is the list of languages that have a supported language server.
var Languages []string

// StaticInfoT represents static information about a language server.
type StaticInfoT struct {
	// DisplayName is the display name of the language. e.g. "PHP" and
	// "TypeScript", rather than language keys which are always lowercase "php"
	// and "typescript".
	DisplayName string

	// HomepageURL is the URL to the language server's homepage, or an empty
	// string if there is none.
	HomepageURL string

	// IssuesURL is the URL to the language server's open/known issues, or an
	// empty string if there is none.
	IssuesURL string

	// DocsURL is the URL to the language server's documentation, or an empty
	// string if there is none.
	DocsURL string

	siteConfig schema.Langservers
}

var StaticInfo = map[string]*StaticInfoT{
	"go": {
		DisplayName: "Go",
		HomepageURL: "https://github.com/sourcegraph/go-langserver",
		IssuesURL:   "https://github.com/sourcegraph/go-langserver/issues",
		DocsURL:     "https://github.com/sourcegraph/go-langserver/blob/master/README.md",
		siteConfig:  schema.Langservers{Language: "go", Address: "tcp://go:4389"},
	},
	"typescript": {
		DisplayName: "TypeScript",
		HomepageURL: "https://github.com/sourcegraph/javascript-typescript-langserver",
		IssuesURL:   "https://github.com/sourcegraph/javascript-typescript-langserver/issues",
		DocsURL:     "https://github.com/sourcegraph/javascript-typescript-langserver/blob/master/README.md",
		siteConfig:  schema.Langservers{Language: "typescript", Address: "tcp://typescript:2088"},
	},
	"javascript": {
		DisplayName: "JavaScript",
		HomepageURL: "https://github.com/sourcegraph/javascript-typescript-langserver",
		IssuesURL:   "https://github.com/sourcegraph/javascript-typescript-langserver/issues",
		DocsURL:     "https://github.com/sourcegraph/javascript-typescript-langserver/blob/master/README.md",
		siteConfig:  schema.Langservers{Language: "javascript", Address: "tcp://typescript:2088"},
	},
	"python": {
		DisplayName: "Python",
		HomepageURL: "https://github.com/sourcegraph/python-langserver",
		IssuesURL:   "https://github.com/sourcegraph/python-langserver/issues",
		DocsURL:     "https://github.com/sourcegraph/python-langserver/blob/master/README.md",
		siteConfig:  schema.Langservers{Language: "python", Address: "tcp://python:2087"},
	},
	"java": {
		DisplayName: "Java",
		HomepageURL: "",
		IssuesURL:   "",
		DocsURL:     "",
		siteConfig:  schema.Langservers{Language: "php", Address: "tcp://java:2088"},
	},
	"php": {
		DisplayName: "PHP",
		HomepageURL: "https://github.com/felixfbecker/php-language-server",
		IssuesURL:   "https://github.com/felixfbecker/php-language-server/issues",
		DocsURL:     "https://github.com/felixfbecker/php-language-server/blob/master/README.md",
		siteConfig:  schema.Langservers{Language: "php", Address: "tcp://php:2088"},
	},
}

var debugContainerPorts = map[string]struct {
	HostPort, ContainerPort string
}{
	"go":         {"2081", ""},
	"typescript": {"2082", ""},
	"javascript": {"2082", ""},
	"python":     {"2083", ""},
	"java":       {"2084", ""},
	"php":        {"2085", ""},
}

func init() {
	for lang := range StaticInfo {
		Languages = append(Languages, lang)
	}
	sort.Strings(Languages)

	if envvar.DebugMode() {
		// Running in debug / development mode. In this case, the frontend is
		// not running inside of a Docker container itself, and we are not on
		// the 'lsp' network. So we must actually expose the language server's
		// ports to the network (via `docker run` `-p` flag).
		//
		// When running inside a Docker container, this is not desirable as it
		// needlessly exposes services to the network which is a security
		// concern.
		for lang, ls := range StaticInfo {
			// Save the port that the container is listening on internally so
			// that we can specify it later with the docker run `-p` flag.
			split := strings.Split(ls.siteConfig.Address, ":")
			p := debugContainerPorts[lang]
			p.ContainerPort = split[len(split)-1]
			debugContainerPorts[lang] = p

			// Since some language servers (typescript, java, php) listen on
			// the same internal port, we give all explicit host ports
			// (otherwise they would conflict).
			ls.siteConfig.Address = "tcp://localhost:" + debugContainerPorts[lang].HostPort
			StaticInfo[lang] = ls
		}
	}
}

// checkSupported checks if the specified language is in the list of supported
// Languages.
func checkSupported(language string) error {
	for _, supported := range Languages {
		if language == supported {
			return nil
		}
	}
	return fmt.Errorf("language not supported: %q", language)
}

// Update updates the language server for the specified language.
func Update(language string) error {
	language = mapLanguage(language)

	dockerContainerAccess.lock(language)
	defer dockerContainerAccess.unlock(language)

	if err := validate(language); err != nil {
		return err
	}

	// Pull the latest image.
	if err := pull(language); err != nil {
		return err
	}

	// Stop the running Docker container.
	if err := stop(language); err != nil {
		return err
	}

	// Start the Docker container.
	return start(language)
}

// pull pulls the latest image for the given language.
func pull(language string) error {
	dockerPulling.set(language)
	_, err := dockerCmd("pull", imageName(language))
	dockerPulling.delete(language)
	return err
}

// Start starts the language server Docker container for the specified language.
func Start(language string) error {
	language = mapLanguage(language)

	dockerContainerAccess.lock(language)
	defer dockerContainerAccess.unlock(language)

	if err := validate(language); err != nil {
		return err
	}

	// Check if the image exists locally or not. If it does not, we will pull
	// it ourselves manually instead of letting start() implicitly do it below.
	// This is so that dockerPulling is properly kept up-to-date on the initial
	// image pull.
	image, err := dockerInspectImage(imageName(language))
	if err != nil {
		return err
	}
	if image == nil {
		// Pull the latest image so that it is correctly marked as being pulled.
		if err := pull(language); err != nil {
			return err
		}
	}

	// Now start the container.
	return start(language)
}

// Stop stops the language server Docker container for the specified language.
func Stop(language string) error {
	language = mapLanguage(language)

	dockerContainerAccess.lock(language)
	defer dockerContainerAccess.unlock(language)

	if err := validate(language); err != nil {
		return err
	}
	return stop(language)
}

func start(language string) error {
	cmd := []string{"run", "--detach", "--rm", "--network=lsp", "--name=" + language}
	if envvar.DebugMode() {
		cmd = append(cmd, startDebugArgs(language)...)
	} else {
		cmd = append(cmd, startProdArgs(language)...)
	}
	cmd = append(cmd, imageName(language))
	_, err := dockerCmd(cmd...)
	if err != nil && strings.Contains(err.Error(), fmt.Sprintf(`The container name "/%s" is already in use by container`, language)) {
		// already started
		return nil
	}
	return err
}

func startDebugArgs(language string) (args []string) {
	if language == "go" {
		args = append(args, []string{"-e", "SRC_GIT_SERVERS=localhost:3178"}...)
	}

	p := debugContainerPorts[language]
	args = append(args, []string{"-p", p.HostPort + ":" + p.ContainerPort}...)
	return args
}

func startProdArgs(language string) (args []string) {
	if language == "go" {
		args = append(args, []string{"-e", "SRC_GIT_SERVERS=sourcegraph:3178"}...)
	}
	return args
}

func stop(language string) error {
	_, err := dockerCmd("stop", language)
	if err != nil && strings.Contains(err.Error(), "No such container") {
		// already stopped
		return nil
	}
	return err
}

// Restart restarts the language server for the given language.
func Restart(language string) error {
	language = mapLanguage(language)

	dockerContainerAccess.lock(language)
	defer dockerContainerAccess.unlock(language)

	if err := validate(language); err != nil {
		return err
	}
	if err := stop(language); err != nil {
		return err
	}
	return start(language)
}

// Status represents the status of a language server Docker container.
type Status string

const (
	// StatusNone indicates there is no status yet.
	StatusNone Status = "none"

	// StatusStarting indicates the container is starting.
	StatusStarting Status = "starting"

	// StatusHealthy indicates the container is healthy.
	StatusHealthy Status = "healthy"

	// StatusUnhealthy indicates the container is unhealthy.
	StatusUnhealthy Status = "unhealthy"
)

// LangInfo represents the current information about a language server.
type LangInfo struct {
	// Installed is true if any version of the language server Docker image is
	// installed.
	Installed bool

	// Pulling is true if the code intelligence image is currently being pulled.
	Pulling bool

	// Status is one of StatusNone, StatusStarting, StatusHealthy, or
	// StatusUnhealthy.
	Status Status

	// InstalledChecksum is the checksum of the installed docker image.
	// It is empty if the Docker image has not been installed.
	InstalledChecksum string
}

// Running tells if the language server is running.
func (i *LangInfo) Running() bool {
	return i.Installed && i.Status != StatusNone
}

// Info tells the current information of a language server Docker
// container/image.
func Info(language string) (*LangInfo, error) {
	language = mapLanguage(language)

	if err := validate(language); err != nil {
		return nil, err
	}

	container, err := dockerInspectContainer(language)
	if err != nil {
		return nil, err
	}

	installed := container != nil
	pulling := dockerPulling.get(language)
	checksum := ""
	if container != nil {
		checksum = container.Image
	}
	if checksum == "" || !installed {
		// We must execute dockerInspectImage now to get the true answer.
		image, err := dockerInspectImage(imageName(language))
		if err != nil {
			return nil, err
		}
		if image != nil {
			installed = true
			checksum = image.Config.Image
		}
	}

	info := &LangInfo{
		Installed:         installed,
		Pulling:           pulling,
		Status:            StatusNone,
		InstalledChecksum: checksum,
	}

	// Determine status.
	if container != nil && container.State != nil {
		state := container.State
		if state.Health != nil {
			// We mirror Docker status strings, so we can just do type conversion
			// here.
			info.Status = Status(state.Health.Status)
		} else if state.Running && !state.Paused && !state.OOMKilled && !state.Dead {
			info.Status = StatusHealthy
		} else {
			info.Status = StatusUnhealthy
		}
	}
	return info, nil
}

// dockerInspectContainer returns the result of `docker inspect <container>`.
//
// It returns nil, nil if the container does not exist.
func dockerInspectContainer(container string) (*containerInspection, error) {
	stdout, err := dockerCmd("inspect", container)
	if err != nil {
		if strings.Contains(err.Error(), "Error: No such object: ") {
			// This indicates the docker container does not exist.
			return nil, nil
		}
		return nil, err
	}
	var results []*containerInspection
	if err := json.Unmarshal(stdout, &results); err != nil {
		return nil, err
	}
	if len(results) == 0 {
		panic("unexpected state (never occurs, see no such object case above)")
	}
	return results[0], nil
}

// dockerInspectImage returns the result of `docker inspect <image>`.
//
// It returns nil, nil if the container does not exist.
func dockerInspectImage(image string) (*imageInspection, error) {
	stdout, err := dockerCmd("inspect", image)
	if err != nil {
		if strings.Contains(err.Error(), "Error: No such object: ") {
			// This indicates the docker container does not exist.
			return nil, nil
		}
		return nil, err
	}
	var results []*imageInspection
	if err := json.Unmarshal(stdout, &results); err != nil {
		return nil, err
	}
	if len(results) == 0 {
		panic("unexpected state (never occurs, see no such object case above)")
	}
	return results[0], nil
}

// validate returns an error if the language is invalid, or if running in Data
// Center mode, or if the Docker socket is not present.
func validate(language string) error {
	// Check if the language is supported.
	if err := checkSupported(language); err != nil {
		return err
	}

	// We do not support Data Center mode.
	if conf.IsDataCenter(conf.DeployType()) {
		return errors.New("data center is not supported")
	}

	// Check if Docker socket is present.
	haveSocket, err := haveDockerSocket()
	if err != nil {
		return err
	}
	if !haveSocket {
		return errors.New("Docker socket is not exposed / accessible")
	}
	return nil
}

// mapLanguage handles mapping languages like "javascript" to "typescript" as
// "sourcegraph/codeintel-javascript" does not exist but rather is the same
// "sourcegraph/codeintel-typescript" Docker image & container.
func mapLanguage(language string) string {
	switch language {
	case "javascript":
		return "typescript"
	default:
		return language
	}
}

func haveDockerSocket() (bool, error) {
	_, err := os.Stat("/var/run/docker.sock")
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// dockerCmd runs the given docker command and returns stdout or an error.
func dockerCmd(args ...string) ([]byte, error) {
	return cmd("docker", args...)
}

// cmd runs the given command and returns stdout or an error.
func cmd(name string, args ...string) ([]byte, error) {
	cmd := exec.Command(name, args...)
	stdout, err := cmd.Output()
	if err != nil {
		cmdStr := name + " " + strings.Join(args, " ")
		if e, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("exec: %q: error: %s stderr:\n%s", cmdStr, e, e.Stderr)
		}
		return nil, fmt.Errorf("exec: %q: error: %s", cmdStr, err)
	}
	return stdout, nil
}

// thisContainerID returns the name of the current Docker container this
// program is running in. It will only return correct results when running
// inside a Docker container. In other situations, it will return the current
// hostname.
func thisContainerID() (string, error) {
	output, err := cmd("hostname")
	return string(output), err
}

// imageName returns the Docker image name for the given language.
func imageName(language string) string {
	return "sourcegraph/codeintel-" + language
}

var canManage bool

// CanManage tells if language server Docker containers can be managed, or if
// they cannot be due due to some reason that has already been logged (e.g.
// running in Data Center, or a user has not exposed the docker socket to our
// container intentionally).
func CanManage() bool {
	return canManage
}

func init() {
	if !conf.DebugManageDocker() {
		return
	}
	if conf.IsDataCenter(conf.DeployType()) {
		// Do not run in Data Center, or else we would print log messages below
		// about not finding the docker socket.
		return
	}
	// Check if we have a docker socket or not. Situations where we may not
	// have this include:
	//
	// 	- Data center (if there is a regression in conf.DeployType detection)
	// 	- Users not trusting our Docker command ("wtf! I am not giving
	// 	  Sourcegraph access to manage my Docker containers!") and removing
	// 	  the Docker socket portion of our run command.
	//
	haveSocket, err := haveDockerSocket()
	if err != nil {
		log15.Error("langservers: error looking up /var/run/docker.sock", "error", err)
		return
	}
	if !haveSocket {
		log15.Error("langservers: /var/run/docker.sock not found, managing langservers disabled.\nSee https://about.sourcegraph.com/docs/code-intelligence/install")
		return
	}
	canManage = true

	goroutine.Go(func() {
		setContainerID()
		createLSPBridge()

		// Wait for our process to shutdown.
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGHUP)
		<-c

		stopAllLanguageServers()
		os.Exit(1)
	})
}

// setContainerID changes the name of our container to "sourcegraph" so that it is
// reachable on that name via the lsp network. This is needed so that
// e.g. the Go language server can reach the gitserver in our container.
//
// We do not do this in dev mode, since we are not running in a container.
func setContainerID() {
	if envvar.DebugMode() {
		return
	}
	containerID, err := thisContainerID()
	if err != nil {
		log15.Error("langservers: failed to get this container ID", "error", err)
		return
	}
	_, err = dockerCmd("container", "rename", containerID, "sourcegraph")
	if err != nil {
		if !strings.Contains(err.Error(), "same name as its current name") {
			log15.Error("langservers: error renaming Docker container", "error", err)
		}
		return
	}
}

// createLSPBridge creates a bridge network for the language servers to communicate with Sourcegraph Server.
func createLSPBridge() {
	// Create the necessary LSP bridge network.
	_, err := dockerCmd("network", "create", "--driver", "bridge", "lsp")
	if err != nil {
		if !strings.Contains(err.Error(), "network with name lsp already exists") {
			log15.Error("langservers: error creating Docker lsp bridge network", "error", err)
		}
		return
	}
	// Connect this container to the LSP bridge network we just created.
	_, err = dockerCmd("network", "connect", "lsp", "sourcegraph")
	if err != nil {
		if !strings.Contains(err.Error(), "already exists in network") {
			log15.Error("langservers: error connecting Docker container to lsp bridge network", "error", err)
		}
		return
	}
}

// stopAllLanguageServers stops all running language servers.
func stopAllLanguageServers() {
	var wg sync.WaitGroup
	for _, language := range Languages {
		language := language
		wg.Add(1)
		goroutine.Go(func() {
			defer wg.Done()
			info, err := Info(language)
			if err != nil {
				log15.Error("langservers: failed to get info for language server", "language", language, "error", err)
				return
			}
			if !info.Running() {
				// No container for this language running.
				return
			}
			if err := Stop(language); err != nil {
				log15.Error("langservers: error stopping language server", "language", language, "error", err)
			}
		})
	}
	wg.Wait()
}

// dockerPullingT represents which language images are currently being pulled
// (downloaded).
type dockerPullingT struct {
	mu         sync.RWMutex
	byLanguage map[string]struct{}
}

// set sets the given language as being pulled.
func (d *dockerPullingT) set(language string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.byLanguage[language] = struct{}{}
}

// delete sets the given language as no longer being pulled.
func (d *dockerPullingT) delete(language string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	delete(d.byLanguage, language)
}

// get returns true if the given language is being pulled.
func (d *dockerPullingT) get(language string) bool {
	d.mu.RLock()
	defer d.mu.RUnlock()
	_, ok := d.byLanguage[language]
	return ok
}

var dockerPulling = &dockerPullingT{
	byLanguage: make(map[string]struct{}),
}

// mutexByKey represents a map of keys to mutexes.
type mutexByKey struct {
	mu    sync.Mutex
	byKey map[string]*sync.Mutex
}

// get gets the mutex for the given key.
func (m *mutexByKey) get(key string) *sync.Mutex {
	m.mu.Lock()
	defer m.mu.Unlock()

	mu, ok := m.byKey[key]
	if !ok {
		mu = new(sync.Mutex)
		m.byKey[key] = mu
	}
	return mu
}

// lock locks the given key.
func (m *mutexByKey) lock(key string) {
	m.get(key).Lock()
}

// unlock unlocks the given key.
func (m *mutexByKey) unlock(key string) {
	m.get(key).Unlock()
}

// dockerContainerAccess represents access to start/stop/restart/update a given
// Docker container.
var dockerContainerAccess = &mutexByKey{
	byKey: map[string]*sync.Mutex{},
}
