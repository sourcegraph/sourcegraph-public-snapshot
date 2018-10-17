package langservers

import (
	"bytes"
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
	"time"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/pkg/conf"
	"github.com/sourcegraph/sourcegraph/pkg/env"
	"github.com/sourcegraph/sourcegraph/pkg/prefixsuffixsaver"
	"github.com/sourcegraph/sourcegraph/schema"
	log15 "gopkg.in/inconshreveable/log15.v2"
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

	// Experimental indicates that a language server may perform arbitrary code
	// execution, may have limited functionality, etc.
	Experimental bool

	// Whether or not the language server should be killed via
	// `docker kill <container>` or `docker stop <container>`. This is used for
	// some language servers that do not yet properly handle SIGTERM, as
	// otherwise `docker stop <container>` would do nothing and timeout after
	// 10s, ultimately just making e.g. the server restart process much slower.
	//
	// TODO: Remove the need for this: https://github.com/sourcegraph/sourcegraph/issues/10693
	kill bool

	// SiteConfig is the "langservers" site config entry for this language server.
	SiteConfig schema.Langservers
}

// StaticInfo maps language keys to static information about the language
// server. For each entry, the siteConfig.Address must match the port specified
// in the corresponding Dockerfile.
var StaticInfo = map[string]*StaticInfoT{
	"go": {
		DisplayName:  "Go",
		HomepageURL:  "https://github.com/sourcegraph/go-langserver",
		IssuesURL:    "https://github.com/sourcegraph/go-langserver/issues",
		DocsURL:      "https://github.com/sourcegraph/go-langserver/blob/master/README.md",
		Experimental: false,
		SiteConfig:   schema.Langservers{Language: "go", Address: "tcp://go:4389"},
	},
	"typescript": {
		DisplayName:  "TypeScript",
		HomepageURL:  "https://github.com/sourcegraph/javascript-typescript-langserver",
		IssuesURL:    "https://github.com/sourcegraph/javascript-typescript-langserver/issues",
		DocsURL:      "https://github.com/sourcegraph/javascript-typescript-langserver/blob/master/README.md",
		SiteConfig:   schema.Langservers{Language: "typescript", Address: "tcp://typescript:2088"},
		Experimental: false,
	},
	"javascript": {
		DisplayName:  "JavaScript",
		HomepageURL:  "https://github.com/sourcegraph/javascript-typescript-langserver",
		IssuesURL:    "https://github.com/sourcegraph/javascript-typescript-langserver/issues",
		DocsURL:      "https://github.com/sourcegraph/javascript-typescript-langserver/blob/master/README.md",
		SiteConfig:   schema.Langservers{Language: "javascript", Address: "tcp://typescript:2088"},
		Experimental: false,
	},
	"python": {
		DisplayName:  "Python",
		HomepageURL:  "https://github.com/sourcegraph/python-langserver",
		IssuesURL:    "https://github.com/sourcegraph/python-langserver/issues",
		DocsURL:      "https://github.com/sourcegraph/python-langserver/blob/master/README.md",
		SiteConfig:   schema.Langservers{Language: "python", Address: "tcp://python:2087"},
		Experimental: false,
	},
	"java": {
		DisplayName:  "Java",
		HomepageURL:  "https://github.com/sourcegraph/java-langserver-docs",
		IssuesURL:    "https://github.com/sourcegraph/java-langserver-docs/issues",
		DocsURL:      "https://github.com/sourcegraph/java-langserver-docs/blob/master/README.md",
		SiteConfig:   schema.Langservers{Language: "java", Address: "tcp://java:2088"},
		Experimental: false,
		kill:         true,
	},
	"php": {
		DisplayName:  "PHP",
		HomepageURL:  "https://github.com/felixfbecker/php-language-server",
		IssuesURL:    "https://github.com/felixfbecker/php-language-server/issues",
		DocsURL:      "https://github.com/felixfbecker/php-language-server/blob/master/README.md",
		Experimental: false,
		SiteConfig:   schema.Langservers{Language: "php", Address: "tcp://php:2088"},
	},
	"bash": {
		DisplayName:  "Bash",
		HomepageURL:  "https://github.com/mads-hartmann/bash-language-server",
		IssuesURL:    "https://github.com/mads-hartmann/bash-language-server/issues",
		DocsURL:      "https://github.com/mads-hartmann/bash-language-server/blob/master/README.md",
		Experimental: true,
		SiteConfig:   schema.Langservers{Language: "bash", Address: "tcp://bash:8080"},
	},
	"clojure": {
		DisplayName:  "Clojure",
		HomepageURL:  "https://github.com/snoe/clojure-lsp",
		IssuesURL:    "https://github.com/snoe/clojure-lsp/issues",
		DocsURL:      "https://github.com/snoe/clojure-lsp/blob/master/README.md",
		Experimental: true,
		SiteConfig:   schema.Langservers{Language: "clojure", Address: "tcp://clojure:8080"},
	},
	"cpp": {
		DisplayName:  "C++",
		HomepageURL:  "https://github.com/Chilledheart/vim-clangd",
		IssuesURL:    "https://github.com/Chilledheart/vim-clangd/issues",
		DocsURL:      "https://github.com/Chilledheart/vim-clangd/blob/master/README.md",
		Experimental: true,
		SiteConfig:   schema.Langservers{Language: "cpp", Address: "tcp://cpp:8080"},
	},
	"cs": {
		DisplayName:  "C#",
		HomepageURL:  "https://github.com/OmniSharp/omnisharp-node-client",
		IssuesURL:    "https://github.com/OmniSharp/omnisharp-node-client/issues",
		DocsURL:      "https://github.com/OmniSharp/omnisharp-node-client/blob/master/readme.md",
		Experimental: true,
		SiteConfig:   schema.Langservers{Language: "cs", Address: "tcp://cs:8080"},
	},
	"css": {
		DisplayName:  "CSS",
		HomepageURL:  "https://github.com/vscode-langservers/vscode-css-languageserver-bin",
		IssuesURL:    "https://github.com/vscode-langservers/vscode-css-languageserver-bin/issues",
		DocsURL:      "https://github.com/vscode-langservers/vscode-css-languageserver-bin/blob/master/README.md",
		Experimental: true,
		SiteConfig:   schema.Langservers{Language: "css", Address: "tcp://css:8080"},
	},
	"dockerfile": {
		DisplayName:  "Dockerfile",
		HomepageURL:  "https://github.com/rcjsuen/dockerfile-language-server-nodejs",
		IssuesURL:    "https://github.com/rcjsuen/dockerfile-language-server-nodejs/issues",
		DocsURL:      "https://github.com/rcjsuen/dockerfile-language-server-nodejs/blob/master/README.md",
		Experimental: true,
		SiteConfig:   schema.Langservers{Language: "dockerfile", Address: "tcp://docker:8080"},
	},
	"elixir": {
		DisplayName:  "Elixir",
		HomepageURL:  "https://github.com/JakeBecker/elixir-ls",
		IssuesURL:    "https://github.com/JakeBecker/elixir-ls/issues",
		DocsURL:      "https://github.com/JakeBecker/elixir-ls/blob/master/README.md",
		Experimental: true,
		SiteConfig:   schema.Langservers{Language: "elixir", Address: "tcp://elixir:8080"},
	},
	"haskell": {
		DisplayName:  "Haskell",
		HomepageURL:  "https://github.com/haskell/haskell-ide-engine",
		IssuesURL:    "https://github.com/haskell/haskell-ide-engine/issues",
		DocsURL:      "https://github.com/haskell/haskell-ide-engine/blob/master/README.md",
		Experimental: true,
		SiteConfig:   schema.Langservers{Language: "haskell", Address: "tcp://haskell:8080"},
	},
	"html": {
		DisplayName:  "HTML",
		HomepageURL:  "https://github.com/vscode-langservers/vscode-html-languageserver-bin",
		IssuesURL:    "https://github.com/vscode-langservers/vscode-html-languageserver-bin/issues",
		DocsURL:      "https://github.com/vscode-langservers/vscode-html-languageserver-bin/blob/master/README.md",
		Experimental: true,
		SiteConfig:   schema.Langservers{Language: "html", Address: "tcp://html:8080"},
	},
	"lua": {
		DisplayName:  "Lua",
		HomepageURL:  "https://github.com/Alloyed/lua-lsp",
		IssuesURL:    "https://github.com/Alloyed/lua-lsp/issues",
		DocsURL:      "https://github.com/Alloyed/lua-lsp/blob/master/readme.md",
		Experimental: true,
		SiteConfig:   schema.Langservers{Language: "lua", Address: "tcp://lua:8080"},
	},
	"ocaml": {
		DisplayName:  "OCaml",
		HomepageURL:  "https://github.com/freebroccolo/ocaml-language-server",
		IssuesURL:    "https://github.com/freebroccolo/ocaml-language-server/issues",
		DocsURL:      "https://github.com/freebroccolo/ocaml-language-server/blob/master/README.md",
		Experimental: true,
		SiteConfig:   schema.Langservers{Language: "ocaml", Address: "tcp://ocaml:8080"},
	},
	"r": {
		DisplayName:  "R",
		HomepageURL:  "https://github.com/REditorSupport/languageserver",
		IssuesURL:    "https://github.com/REditorSupport/languageserver/issues",
		DocsURL:      "https://github.com/REditorSupport/languageserver/blob/master/README.md",
		Experimental: true,
		SiteConfig:   schema.Langservers{Language: "r", Address: "tcp://rlang:8080"},
	},
	"ruby": {
		DisplayName:  "Ruby",
		HomepageURL:  "https://github.com/castwide/solargraph",
		IssuesURL:    "https://github.com/castwide/solargraph/issues",
		DocsURL:      "https://github.com/castwide/solargraph/blob/master/README.md",
		Experimental: true,
		SiteConfig:   schema.Langservers{Language: "ruby", Address: "tcp://ruby:8080"},
	},
	"rust": {
		DisplayName:  "Rust",
		HomepageURL:  "https://github.com/rust-lang-nursery/rls",
		IssuesURL:    "https://github.com/rust-lang-nursery/rls/issues",
		DocsURL:      "https://github.com/rust-lang-nursery/rls/blob/master/README.md",
		Experimental: true,
		SiteConfig:   schema.Langservers{Language: "rust", Address: "tcp://rust:8080"},
	},
}

// debugContainerPorts specifies which port to expose on localhost during
// development. The ContainerPort field is set to the corresponding port in
// StaticInfo.<lang>.siteConfig.Address at runtime.
//
// For each language server in StaticInfo, there must be a corresponding entry
// in debugContainerPorts. If an entry is missing, you'll get a "can't assign
// requested address" error in the terminal when you enable the language server
// because the port will default to the empty string.
//
// Also, these ports must be unique. If they aren't, then docker run will fail
// due to a port conflict.
var debugContainerPorts = map[string]struct {
	HostPort, ContainerPort string
}{
	"go":         {"2081", ""},
	"typescript": {"2082", ""},
	"javascript": {"2082", ""},
	"python":     {"2083", ""},
	"java":       {"2084", ""},
	"php":        {"2085", ""},
	"bash":       {"2086", ""},
	"clojure":    {"2087", ""},
	"cpp":        {"2088", ""},
	"cs":         {"2089", ""},
	"css":        {"2090", ""},
	"dockerfile": {"2091", ""},
	"elixir":     {"2092", ""},
	"html":       {"2093", ""},
	"lua":        {"2094", ""},
	"ocaml":      {"2095", ""},
	"r":          {"2096", ""},
	"ruby":       {"2097", ""},
	"rust":       {"2098", ""},
	"haskell":    {"2099", ""},
}

func init() {
	for lang := range StaticInfo {
		Languages = append(Languages, lang)
	}
	sort.Strings(Languages)

	if env.InsecureDev {
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
			split := strings.Split(ls.SiteConfig.Address, ":")
			p := debugContainerPorts[lang]
			p.ContainerPort = split[len(split)-1]
			debugContainerPorts[lang] = p

			// Since some language servers (typescript, java, php) listen on
			// the same internal port, we give all explicit host ports
			// (otherwise they would conflict).
			ls.SiteConfig.Address = "tcp://localhost:" + debugContainerPorts[lang].HostPort
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
	if err := CanManage(); err != nil {
		return err
	}
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
	if err := CanManage(); err != nil {
		return err
	}
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

// Stop stops the language server Docker container for the specified language,
// if the specified language is the only active user of it. For example, if
// javascript is being stopped but typescript is still enabled in the site
// config, this function is no-op.
func Stop(language string) error {
	if err := CanManage(); err != nil {
		return err
	}

	// Determine if the container is in use according to the site config.
	inUse := false
	for _, lang := range relatedLanguages(language) {
		state, err := State(lang)
		if err != nil {
			return err
		}
		if state == StateEnabled {
			inUse = true
		}
	}
	if inUse {
		return nil // Container is in use, so don't stop it.
	}
	return ForceStop(language)
}

// ForceStop stops the language server Docker container for the specified language.
//
// If the language server for the specified language provides multiple languages
// (e.g. javascript and typescript are provided by the same language server),
// both will be stopped.
func ForceStop(language string) error {
	if err := CanManage(); err != nil {
		return err
	}
	language = mapLanguage(language)

	dockerContainerAccess.lock(language)
	defer dockerContainerAccess.unlock(language)

	if err := validate(language); err != nil {
		return err
	}
	return stop(language)
}

func start(language string) error {
	// Remove the container if it already exists. This effectively gives us the same behavior
	// as `docker run --rm` (which we cannot use as `--rm` and `--restart=always` are mutually
	// exclusive options). If we did not do this here, the container name may already be in
	// use by a container that was previously started but not running. i.e., docker run will
	// never *start* a container that is OK but not running (`docker start` would have to be
	// used):
	//
	// 	$ docker ps
	// 	[...] IMAGE                               STATUS                         NAMES
	// 	[...] sourcegraph/codeintel-typescript    Exited (137) 32 minutes ago    typescript
	//
	// 	$ docker run --name=typescript --restart=always sourcegraph/codeintel-typescript
	// 	docker: Error response from daemon: Conflict. The container name "/typescript" is already in use by container "cd64c8a4dfdc811e945d9d668026fd9552e6a4d9b28353a38c10446dc16ecaf5". You have to remove (or rename) that container to be able to reuse that name.
	// 	See 'docker run --help'.
	//
	// NOTE: We do not yet handle the fact that we could be removing someone else's container
	// whose name conflicts with ours. This has *always* been true historically, ever since we
	// suggested --rm to users before we had managed Docker. To fix this, we should switch to
	// container names that are unlikely to conflict like "sourcegraph-typescript" or explore
	// DIND (https://github.com/sourcegraph/sourcegraph/issues/11616). Note however that you
	// cannot `docker rm` a running container, so the issue would be in Sourcegraph and not
	// the user's conflicting container generally.
	_, err := dockerCmd("rm", containerName(language))
	if err != nil && !strings.Contains(err.Error(), "No such container") {
		// Also ignore running container warnings -- since that means it is most likely ours
		// and we will hit the 'container name is already in use' error case below.
		if !strings.Contains(err.Error(), "You cannot remove a running container") {
			return err
		}
	}

	cmd := []string{"run", "--detach", "--restart=always", "--network=lsp", "--name=" + containerName(language)}
	if env.InsecureDev {
		cmd = append(cmd, startDebugArgs(language)...)
	} else {
		cmd = append(cmd, startProdArgs(language)...)
	}
	cmd = append(cmd, imageName(language))
	_, err = dockerCmd(cmd...)
	if err != nil && strings.Contains(err.Error(), fmt.Sprintf(`The container name "/%s" is already in use by container`, containerName(language))) {
		// already started
		return nil
	}
	if err != nil {
		return err
	}

	// Force an update of the info cache. We do not usually need to do this,
	// but in the case of starting a language server we do. Otherwise the status
	// would be StatusNone which is inherently unhealthy and the caller would
	// be forced to display the langserver as "unhealthy" until the cache is
	// updated in a second or so.
	_, err = updateInfoCache(language)
	return err
}

func startDebugArgs(language string) (args []string) {
	if language == "go" {
		// The address where gitserver is accessible from the Go language server.
		host := os.Getenv("GOLANGSERVER_SRC_GIT_SERVERS")
		if host == "" {
			host = "localhost:3178"
		}
		args = append(args, []string{"-e", fmt.Sprintf("SRC_GIT_SERVERS=%s", host)}...)
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
	cmdName := "stop"
	if StaticInfo[language].kill {
		cmdName = "kill"
	}
	startTime := time.Now()
	_, err := dockerCmd(cmdName, containerName(language))
	timeToKill := time.Since(startTime)
	if err != nil {
		if strings.Contains(err.Error(), "No such container") {
			// already stopped
			return nil
		}
		return err
	}
	if timeToKill > 10*time.Second { // 'docker stop' kills after 10s by default.
		log15.Warn("langservers: stopping the language server took unusually long and docker may have had to SIGKILL it; is it responding to signals properly?", "duration", time.Since(startTime), "language", language)
	}
	return nil
}

// Restart restarts the language server for the given language.
func Restart(language string) error {
	if err := CanManage(); err != nil {
		return err
	}
	language = mapLanguage(language)

	dockerContainerAccess.lock(language)
	defer dockerContainerAccess.unlock(language)

	if err := validate(language); err != nil {
		return err
	}
	_, err := dockerCmd("restart", containerName(language))
	return err
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
	return i.Installed && i.Status != StatusNone && i.Status != StatusUnhealthy
}

type cacheEntry struct {
	info *LangInfo
	err  error
}

var latestInfo = struct {
	sync.RWMutex
	byLanguage map[string]cacheEntry
}{
	byLanguage: map[string]cacheEntry{},
}

func queryContainerInfoWorker() {
	// Monitor events to containers, when a containers state changes we
	// refresh our state from docker. We use a buffer size of 1 so if an event
	// happens while we are running updateInfoCache we will run
	// updateInfoCache again.
	event := make(chan struct{}, 1)
	go dockerEventNotify(event, "type=container")
	// Periodically refresh everything anyways, just in case we aren't
	// notified of an event.
	refresh := time.Tick(time.Minute)
	for {
		select {
		case <-event:
		case <-refresh:
		}

		for _, language := range Languages {
			updateInfoCache(language)
		}

		// Always sleep at least 1 second incase we have a lot of events
		// leading to use refreshing state at a high rate.
		time.Sleep(1 * time.Second)
	}
}

// Info tells the current information of a language server Docker
// container/image.
func Info(language string) (*LangInfo, error) {
	if err := CanManage(); err != nil {
		return nil, err
	}

	// Check if info exists in the cache already and use it if so.
	latestInfo.RLock()
	e, ok := latestInfo.byLanguage[language]
	latestInfo.RUnlock()
	if ok {
		return e.info, e.err
	}

	// No info in the cache yet (e.g. maybe the queryContainerInfoWorker has
	// not had a chance to run yet), query the info directly.
	return updateInfoCache(language)
}

// updateInfoCache calls infoUncached and caches the result for future use.
func updateInfoCache(language string) (*LangInfo, error) {
	info, err := infoUncached(language)
	latestInfo.Lock()
	latestInfo.byLanguage[language] = cacheEntry{info: info, err: err}
	latestInfo.Unlock()
	return info, err
}

// infoUncached queries the current information of a language server directly
// from Docker.
func infoUncached(language string) (*LangInfo, error) {
	language = mapLanguage(language)

	if err := validate(language); err != nil {
		return nil, err
	}

	container, err := dockerInspectContainer(containerName(language))
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

	if reason, ok := conf.SupportsManagingLanguageServers(); !ok {
		return errors.New(reason)
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

// multiLanguageProviders maps a single sourcegraph/codeintel-$KEY container to the
// multiple languages it provides. For example, sourcegraph/codeintel-typescript
// provides both "typescript" and "javascript" languages.
//
// Docker images that do not provide multiple languages do not need to be
// listed here.
var multiLanguageProviders = map[string][]string{
	// sourcegraph/codeintel-typescript provides "typescript" and "javascript" languages.
	"typescript": []string{
		"typescript",
		"javascript",
	},
}

// mapLanguage handles mapping languages like "javascript" to "typescript" as
// "sourcegraph/codeintel-javascript" does not exist but rather is the same
// "sourcegraph/codeintel-typescript" Docker image & container.
func mapLanguage(language string) string {
	for provider, multiLanguages := range multiLanguageProviders {
		for _, multiLanguage := range multiLanguages {
			if language == multiLanguage {
				return provider
			}
		}
	}
	return language
}

// relatedLanguages returns a list of all languages that are provided by the
// specified language's Docker image. Examples:
//
// 	relatedLanguages("typescript") == []string{"typescript", "javascript"}
// 	relatedLanguages("javascript") == []string{"typescript", "javascript"}
// 	relatedLanguages("go") == []string{"go"}
//
func relatedLanguages(language string) []string {
	for provider, multiLanguages := range multiLanguageProviders {
		if language == provider {
			return multiLanguages
		}
		for _, multiLanguage := range multiLanguages {
			if language == multiLanguage {
				return multiLanguages
			}
		}
	}
	return []string{language}
}

// notifyNewLine notifies c each time a read of Stdout contains a new line,
// but not every newline. It is meant to wake up processes waiting for new
// output. Additionally once the process is started it notifies c.
//
// It sets cmd.Stdout and starts cmd and blocks until cmd is finished.
func notifyNewLine(cmd *exec.Cmd, c chan<- struct{}) error {
	if cmd.Stderr == nil {
		cmd.Stderr = &prefixsuffixsaver.Writer{N: 32 << 10}
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	err = cmd.Start()
	if err != nil {
		return err
	}

	done := make(chan struct{})
	go func() {
		defer close(done)

		// notify we have started
		select {
		case c <- struct{}{}:
		default:
		}

		buf := make([]byte, 4096)
		for {
			n, err := stdout.Read(buf)
			// Each event is on its own line, so just look for a line in
			// the buffer.
			if bytes.IndexByte(buf[:n], '\n') >= 0 {
				select {
				case c <- struct{}{}:
				default:
				}
			}
			if err != nil {
				return
			}
		}
	}()

	err = cmd.Wait()

	// wait until goroutine is done so we stop writing to c when we return.
	<-done

	if err != nil && cmd.Stderr != nil {
		if w, ok := cmd.Stderr.(interface{ Bytes() []byte }); ok {
			return fmt.Errorf("exec: docker events: error: %s stderr:\n%s", err, w.Bytes())
		}
	}

	return err
}

// dockerEventNotify notifies c each time there is one or more docker event(s)
// matching a filter.
//
// dockerEventNotify will not block sending to c and is best-effort. On the
// underlying "docker events" command starting it will notify, and will then
// notify for each event received. It will restart docker event in the case of
// the process dieing.
//
// See the documentation for "docker event" for more information on the
// filters.
func dockerEventNotify(c chan<- struct{}, filter ...string) {
	args := []string{"events"}
	for _, f := range filter {
		args = append(args, "--filter="+f)
	}

	for {
		err := notifyNewLine(exec.Command("docker", args...), c)
		if err != nil {
			// If we failed with error, wait a bit longer to prevent fast
			// spamming.
			log15.Warn("docker events failed. Will try again in 30s", "error", err)
			time.Sleep(30 * time.Second)
			continue
		}

		time.Sleep(1 * time.Second)
	}
}

func haveDockerSocket() (bool, error) {
	if conf.IsDev(conf.DeployType()) && conf.DebugNoDockerSocket() {
		return false, nil
	}
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
	switch language {
	case "cs":
		return "sourcegraph/codeintel-csharp"
	case "dockerfile":
		return "sourcegraph/codeintel-docker"
	default:
		return "sourcegraph/codeintel-" + language
	}
}

// containerName returns the Docker container name for the given language.
func containerName(language string) string {
	switch language {
	case "r":
		// Single-character container names are not allowed by Docker.
		return "rlang"
	default:
		return language
	}
}

var (
	canManage error
)

// CanManage reports whether language servers can be managed via the functions
// in this package:
//
// 	Update
// 	Start
// 	Stop
// 	Info
// 	Restart
//
// If this function returns an error, it indicates that they cannot be managed
// for some reason. Otherwise, nil is returned. The error message is suitable
// for display directly to e.g. a site admin.
//
// The most common reasons for lacking this capability are being deployed to a cluster or the admin
// intentionally not exposing the Docker socket to the Sourcegraph container (for security reasons).
func CanManage() error {
	return canManage
}

func init() {
	reason, ok := conf.SupportsManagingLanguageServers()
	if !ok {
		canManage = errors.New(reason)
		return
	}

	// Check if we have a docker socket or not. Situations where we may not
	// have this include:
	//
	// 	- Cluster deployments (if there is a regression in conf.DeployType detection)
	// 	- Users not trusting our Docker command ("wtf! I am not giving
	// 	  Sourcegraph access to manage my Docker containers!") and removing
	// 	  the Docker socket portion of our run command.
	//
	haveSocket, err := haveDockerSocket()
	if err != nil {
		const msg = "Language server management capabilities disabled due to an error looking up /var/run/docker.sock"
		canManage = fmt.Errorf("%s: %s.", msg, err)
		log15.Error(msg+".", "error", err)
		return
	}
	if !haveSocket {
		const msg = "Language server management capabilities disabled because /var/run/docker.sock was not found. See https://about.sourcegraph.com/docs/code-intelligence/install for help." // TODO!(sqs): update this
		canManage = errors.New(msg)
		log15.Error(msg)
		return
	}

	setupNetworking()

	goroutine.Go(queryContainerInfoWorker)
	goroutine.Go(func() {
		// Wait for our process to shutdown.
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGHUP)
		<-c

		stopAllLanguageServers()
		os.Exit(1)
	})
}

// setupNetworking handles setting up the networking required for our
// sourcegraph/server container and language server containers.
//
// This function MUST be invoked inside at init.
func setupNetworking() {
	// 🦄🐉🦄 Docker for Mac bug / issue workaround 🦄🐉🦄
	//
	// For some reason, in Docker for Mac versions as recent as 18.03.0-ce and
	// probably much later versions than this, the lsp bridge network we create
	// gets into a completely broken DNS state:
	//
	// 	dial tcp: lookup go on 127.0.0.11:53: no such host
	//
	// It is completely unclear to us why this happens, but what we do know is:
	//
	// 1. It seems to be a bug in Docker for Mac, specifically around running
	//    multiple `docker` commands at once. For example, we do not have this
	//    managed Docker running in development mode because process restarting
	//    would trigger this situation much more frequently (see https://github.com/sourcegraph/sourcegraph/pull/10600)
	//
	// 2. When the network does become bugged, it is still connected but
	//    experiences EXTREME packet loss (about 79%), i.e. `docker exec sourcegraph ping go`
	//    does work, but only very rarely and only after multiple minutes.
	//
	// 3. The issue is not reliably reproducible, but does happen.
	//
	// 4. Deleting the lsp network via `docker network rm lsp` resolves the
	//    issue.
	//
	// Bugs that may be related, but are not specifically this issue:
	//
	// - https://github.com/docker/for-mac/issues/997
	// - https://github.com/moby/moby/issues/24344
	// - https://github.com/theupdateframework/notary/pull/753
	//

	// Workaround #1: Do not allow Docker container modification commands (like
	// Start/Stop/Restart/Pull funcs in this package) to run while we create
	// the network. We do this by locking every language.
	for _, lang := range Languages {
		dockerContainerAccess.lock(lang)
	}

	// Do not block server startup, since we're running inside init.
	goroutine.Go(func() {
		// Unlock container access for all languages once we're done.
		defer func() {
			for _, lang := range Languages {
				dockerContainerAccess.unlock(lang)
			}
		}()

		// Set our container ID now, so we can attach it to the network later.
		setContainerID()

		// Workaround #2: Since we're creating the LSP network below anyway,
		// we may as well delete the existing one as it could be in a bugged
		// state.
		deleteLSPBridge()

		// Now create the LSP bridge network.
		createLSPBridge()
	})
}

// setContainerID changes the name of our container to "sourcegraph" so that it is
// reachable on that name via the lsp network. This is needed so that
// e.g. the Go language server can reach the gitserver in our container.
func setContainerID() {
	// We do not do this in dev mode, since we are not running in a container.
	if env.InsecureDev {
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

// deleteLSPBridge deletes the lsp bridge network if it exists.
func deleteLSPBridge() {
	_, err := dockerCmd("network", "rm", "lsp")
	if err != nil && !strings.Contains(err.Error(), "No such network") && !strings.Contains(err.Error(), "not found") {
		// In dev mode, deleting the LSP bridge almost always fails because goreman doesn't
		// send us SIGINT or SIGHUP(!) so we cannot do graceful shutdown of containers, and
		// hence the network would always have active endpoints.
		if env.InsecureDev {
			return
		}
		log15.Error("langservers: error deleting Docker lsp bridge network", "error", err)
	}
}

// createLSPBridge creates a bridge network for the language servers to communicate with sourcegraph/server.
func createLSPBridge() {
	// Create the necessary LSP bridge network.
	_, err := dockerCmd("network", "create", "--driver", "bridge", "lsp")
	if err != nil {
		if !strings.Contains(err.Error(), "network with name lsp already exists") {
			log15.Error("langservers: error creating Docker lsp bridge network", "error", err)
		}
		// Don't return here because we want to try connecting our container to the network below.
	}

	// Connect this container to the LSP bridge network we just created.
	//
	// We do not do this in dev mode, since we are not running in a container.
	if env.InsecureDev {
		return
	}
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
			if err := ForceStop(language); err != nil {
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
