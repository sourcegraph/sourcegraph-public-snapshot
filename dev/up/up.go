package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"strings"

	yaml "gopkg.in/yaml.v2"
)

// config represents a Docker Compose configuration (as is defined in
// a docker-compose.yml).  Note that this struct does not capture all
// possible structures allowed in docker-compose.yml, but is
// specifically tailored to the structure of our specific
// docker-compose.yml. In other words, you might have to change this
// type if our docker-compose.yml changes.
type config struct {
	Version  string              `yaml:"version"`
	Services map[string]*service `yaml:"services"`
}

// service is a service in a Docker Compose configuration
type service struct {
	Image       string                 `yaml:"image"`
	Entrypoint  []string               `yaml:"entrypoint,omitempty"`
	Command     []string               `yaml:"command,omitempty"`
	Environment map[string]interface{} `yaml:"environment,omitempty"`
	Restart     string                 `yaml:"restart,omitempty"`
	Ports       []string               `yaml:"ports,omitempty"`
	DependsOn   []string               `yaml:"depends_on,omitempty"`
	Volumes     []string               `yaml:"volumes,omitempty"`
}

// manualCmds is a mapping from Docker Compose service to the
// corresponding commands that should be run on the host machine if we
// want to run these services on the host instead of inside Docker
// containers.
var manualCmds = map[string][]string{
	"frontend": []string{
		`curl -Ss -o /dev/null "$WEBPACK_DEV_SERVER_URL" || (cd ui && yarn && yarn run start)`,
		`PGSSLMODE=disable VSCODE_BROWSER_PKG=/tmp/VSCode-browser USE_VSCODE_UI=1 vendor/.bin/rego -installenv=GOGC=off,GODEBUG=sbrk=1 -tags="${GOTAGS-}" -extra-watches='app/templates/*' sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend`,
	},
	"gitserver-1":  []string{"go install ./cmd/gitserver && gitserver"},
	"indexer":      []string{"go install ./cmd/indexer && indexer"},
	"searcher":     []string{"go install ./cmd/searcher && searcher"},
	"github-proxy": []string{"go install ./cmd/github-proxy && github-proxy"},
	"lsp-proxy":    []string{"go install ./cmd/lsp-proxy && lsp-proxy -prof-http=:6061"},
}

// hostCommand defines a command that has been extracted from the
// Docker Compose configuration and should be run on the host machine.
type hostCommand struct {
	// Command is the CLI comand
	Command string

	// Env is the environment variables that accompany the command
	Env map[string]interface{}

	// Service is the name of the original service in docker-compose
	Service string
}

// String returns a CLI representation of the hostCommand
func (h *hostCommand) String() string {
	var cmps []string
	if h.Env != nil {
		for k, v := range h.Env {
			cmps = append(cmps, fmt.Sprintf("export %s=%v", k, v))
		}
	}
	cmps = append(cmps, h.Command)
	return strings.Join(cmps, ";")
}

var (
	dockerComposeFile = flag.String("f", "", "path to the docker-compose.yml file")
)

func main() {
	flag.Parse()
	if err := run(flag.Args()); err != nil {
		log.Fatal(err)
	}
}

func run(hostSvcs []string) error {
	cfgFile := *dockerComposeFile
	if cfgFile == "" {
		return fmt.Errorf("must specify docker-compose.yml file")
	}

	hostIP, err := discoverLocalhostIP()
	if err != nil {
		return fmt.Errorf("could not find localhost external IP: %s", err)
	}
	var cfg config
	if err := readConfig(cfgFile, &cfg); err != nil {
		return err
	}
	hostCmds, err := transformConfig(&cfg, hostIP, hostSvcs)
	if err != nil {
		return err
	}

	if len(hostCmds) > 0 {
		fmt.Fprintln(os.Stderr, "\nRUN the following on the host machine:")
		for _, hostCmd := range hostCmds {
			fmt.Fprintf(os.Stderr, "  %s\n", hostCmd)
		}
		fmt.Fprintln(os.Stderr, "")
	} else {
		fmt.Fprintln(os.Stderr, "No services running on host")
	}

	out, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	fmt.Println("##################################################")
	fmt.Println("# !!! AUTO-GENERATED FILE via `go run up.go` !!! #")
	fmt.Println("##################################################")
	fmt.Println()
	fmt.Printf("%s\n", string(out))
	return nil
}

// transformConfig transforms the Docker Compose configuration (cfg)
// by removing the host services (hostSvcs) and rewriting references
// to those services' ports to use `hostIP`. transformConfig modifies
// the config to remove the host services and returns the list of
// commands you should run on the host.
func transformConfig(cfg *config, hostIP string, hostSvcs []string) ([]*hostCommand, error) {
	// Compute host commands
	var hostCmds []*hostCommand

	// original service names
	originalSvcs := make(map[string]struct{})
	for s, _ := range cfg.Services {
		originalSvcs[s] = struct{}{}
	}

	hostSvcSet := make(map[string]struct{})
	for _, svc := range hostSvcs {
		hostSvcSet[svc] = struct{}{}
	}

	// delete overridden services
	for _, hostSvc := range hostSvcs {
		if _, exist := cfg.Services[hostSvc]; !exist {
			return nil, fmt.Errorf("did not find service %q in docker-compose.yml", hostSvc)
		}
		svc := cfg.Services[hostSvc]
		for _, mc := range manualCmds[hostSvc] {
			hostCmds = append(hostCmds, &hostCommand{
				Command: mc,
				Env:     svc.Environment,
				Service: hostSvc,
			})
		}
		delete(cfg.Services, hostSvc)
		if _, exist := manualCmds[hostSvc]; !exist {
			log.Printf("WARNING: did not find manual command for disabled service %q. Are you sure you know what you're doing?", hostSvc)
		}
	}
	for _, svc := range cfg.Services {
		// rewrite overriden service references to use hostIP
		for k, v := range svc.Environment {
			if vstr, ok := v.(string); ok {
				for _, replacedSvc := range hostSvcs {
					vstr = strings.Replace(vstr, replacedSvc, hostIP, 1)
				}
				svc.Environment[k] = vstr
			}
		}
		if len(svc.DependsOn) > 0 {
			newDependsOn := make([]string, 0, len(svc.DependsOn))
			for _, dep := range svc.DependsOn {
				if _, in := hostSvcSet[dep]; !in {
					newDependsOn = append(newDependsOn, dep)
				}
			}
			svc.DependsOn = newDependsOn
		}
	}

	// always remove "initializer" in dev
	if _, exist := cfg.Services["initializer"]; exist {
		delete(cfg.Services, "initializer")
	}

	return hostCmds, nil
}

// readConfig reads a Docker Compose configuration from the file at
// `filename` into `cfg`. `cfg` should either be of type *config or
// *map[string]interface{}.
func readConfig(filename string, cfg interface{}) error {
	var (
		b   []byte
		err error
	)
	b, err = ioutil.ReadFile(filename)
	if err != nil {
		return err
	}
	return yaml.Unmarshal(b, cfg)
}

// discoverLocalhostIP returns the external (local) IP of the host
// machine. This is the IP that docker containers can use to
// communicate with the host machine.
func discoverLocalhostIP() (string, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}
	for _, i := range ifaces {
		if addrs, err := i.Addrs(); err == nil {
			for _, addr := range addrs {
				var ip net.IP
				switch v := addr.(type) {
				case *net.IPNet:
					ip = v.IP
				}
				if ip.To4() != nil && ip.String() != "127.0.0.1" {
					return ip.String(), nil
				}
			}
		}
	}
	return "", fmt.Errorf("localhost IP address not found")
}
