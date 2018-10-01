package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/joho/godotenv"
	"gopkg.in/yaml.v2"
)

const version = "0.2.0"

func usage() {
	fmt.Fprint(os.Stderr, `Tasks:
  goreman check                      # Show entries in Procfile
  goreman help [TASK]                # Show this help
  goreman export [FORMAT] [LOCATION] # Export the apps to another process
                                       (upstart)
  goreman run COMMAND [PROCESS...]   # Run a command
                                       start
                                       stop
                                       stop-all
                                       restart
                                       restart-all
                                       list
                                       status
  goreman start [PROCESS]            # Start the application
  goreman version                    # Display Goreman version

Options:
`)
	flag.PrintDefaults()
	os.Exit(0)
}

// -- process information structure.
type procInfo struct {
	proc       string
	cmdline    string
	cmd        *exec.Cmd
	port       uint
	setPort    bool
	colorIndex int

	// True if we called stopProc to kill the process, in which case an
	// *os.ExitError is not the fault of the subprocess
	stoppedBySupervisor bool

	mu      sync.Mutex
	cond    *sync.Cond
	waitErr error
}

// process informations named with proc.
var procs map[string]*procInfo

// filename of Procfile.
var procfile = flag.String("f", "Procfile", "proc file")

// rpc port number.
var port = flag.Uint("p", defaultPort(), "port")

// base directory
var basedir = flag.String("basedir", "", "base directory")

// base of port numbers for app
var baseport = flag.Uint("b", 5000, "base number of port")

var setPorts = flag.Bool("set-ports", true, "False to avoid setting PORT env var for each subprocess")

// true to exit the supervisor
var exitOnError = flag.Bool("exit-on-error", false, "Exit goreman if a subprocess quits with a nonzero return code")

var maxProcNameLength = 0

var re = regexp.MustCompile(`\$([a-zA-Z]+[a-zA-Z0-9_]+)`)

type config struct {
	Procfile string `yaml:"procfile"`
	// Port for RPC server
	Port     uint   `yaml:"port"`
	BaseDir  string `yaml:"basedir"`
	BasePort uint   `yaml:"baseport"`
	Args     []string
	// If true, exit the supervisor process if a subprocess exits with an error.
	ExitOnError bool `yaml:"exit_on_error"`
}

func readConfig() *config {
	var cfg config

	flag.Parse()
	if flag.NArg() == 0 {
		usage()
	}

	cfg.Procfile = *procfile
	cfg.Port = *port
	cfg.BaseDir = *basedir
	cfg.BasePort = *baseport
	cfg.ExitOnError = *exitOnError
	cfg.Args = flag.Args()

	b, err := ioutil.ReadFile(".goreman")
	if err == nil {
		yaml.Unmarshal(b, &cfg)
	}
	return &cfg
}

// read Procfile and parse it.
func readProcfile(cfg *config) error {
	content, err := ioutil.ReadFile(cfg.Procfile)
	if err != nil {
		return err
	}
	procs = map[string]*procInfo{}
	index := 0
	for _, line := range strings.Split(string(content), "\n") {
		tokens := strings.SplitN(line, ":", 2)
		if len(tokens) != 2 || tokens[0][0] == '#' {
			continue
		}
		k, v := strings.TrimSpace(tokens[0]), strings.TrimSpace(tokens[1])
		if runtime.GOOS == "windows" {
			v = re.ReplaceAllStringFunc(v, func(s string) string {
				return "%" + s[1:] + "%"
			})
		}
		p := &procInfo{proc: k, cmdline: v, colorIndex: index}
		if *setPorts == true {
			p.setPort = true
			p.port = cfg.BasePort
			cfg.BasePort += 100
		}
		p.cond = sync.NewCond(&p.mu)
		procs[k] = p
		if len(k) > maxProcNameLength {
			maxProcNameLength = len(k)
		}
		index++
		if index >= len(colors) {
			index = 0
		}
	}
	if len(procs) == 0 {
		return errors.New("no valid entry")
	}
	return nil
}

func defaultServer(serverPort uint) string {
	if s, ok := os.LookupEnv("GOREMAN_RPC_SERVER"); ok {
		return s
	}
	return fmt.Sprintf("127.0.0.1:%d", defaultPort())
}

func defaultAddr() string {
	if s, ok := os.LookupEnv("GOREMAN_RPC_ADDR"); ok {
		return s
	}
	return "0.0.0.0"
}

// default port
func defaultPort() uint {
	s := os.Getenv("GOREMAN_RPC_PORT")
	if s != "" {
		i, err := strconv.Atoi(s)
		if err == nil {
			return uint(i)
		}
	}
	return 8555
}

// command: check. show Procfile entries.
func check(cfg *config) error {
	err := readProcfile(cfg)
	if err != nil {
		return err
	}
	keys := make([]string, len(procs))
	i := 0
	for k := range procs {
		keys[i] = k
		i++
	}
	sort.Strings(keys)
	fmt.Printf("valid procfile detected (%s)\n", strings.Join(keys, ", "))
	return nil
}

// command: start. spawn procs.
func start(ctx context.Context, sig <-chan os.Signal, cfg *config) error {
	err := readProcfile(cfg)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithCancel(ctx)
	if len(cfg.Args) > 1 {
		tmp := make(map[string]*procInfo, len(cfg.Args[1:]))
		maxProcNameLength = 0
		for _, v := range cfg.Args[1:] {
			p, ok := procs[v]
			if !ok {
				return errors.New("unknown proc: " + v)
			}
			tmp[v] = p
			if len(v) > maxProcNameLength {
				maxProcNameLength = len(v)
			}
		}
		procs = tmp
	}
	godotenv.Load()
	rpcChan := make(chan *rpcMessage, 10)
	go startServer(ctx, rpcChan, cfg.Port)
	procsErr := startProcs(sig, rpcChan, cfg.ExitOnError)
	cancel() // If procs have returned/errored, cancel the RPC server.
	return procsErr
}

func main() {
	var err error
	cfg := readConfig()

	if cfg.BaseDir != "" {
		err = os.Chdir(cfg.BaseDir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "goreman: %s\n", err.Error())
			os.Exit(1)
		}
	}

	cmd := cfg.Args[0]
	switch cmd {
	case "check":
		err = check(cfg)
		break
	case "help":
		usage()
		break
	case "run":
		if len(cfg.Args) >= 2 {
			cmd, args := cfg.Args[1], cfg.Args[2:]
			err = run(cmd, args, cfg.Port)
		} else {
			usage()
		}
		break
	case "export":
		if len(cfg.Args) == 3 {
			format, path := cfg.Args[1], cfg.Args[2]
			err = export(cfg, format, path)
		} else {
			usage()
		}
		break
	case "start":
		c := notifyCh()
		err = start(context.Background(), c, cfg)
		break
	case "version":
		fmt.Println(version)
		break
	default:
		usage()
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %v\n", os.Args[0], err.Error())
		os.Exit(1)
	}
}
