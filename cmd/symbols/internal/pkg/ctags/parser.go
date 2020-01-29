package ctags

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"

	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/internal/env"
)

type Entry struct {
	Name       string
	Path       string
	Line       int
	Kind       string
	Language   string
	Parent     string
	ParentKind string
	Pattern    string
	Signature  string

	FileLimited bool
}

const debug = false

var logErrors = os.Getenv("DEPLOY_TYPE") == "dev"

type Parser interface {
	Parse(path string, content []byte) ([]Entry, error)
	Close()
}

func isCommandAvailable(name string) bool {
	cmd := exec.Command("/bin/sh", "-c", "command -v "+name)
	if err := cmd.Run(); err != nil {
		return false
	}
	return true
}

var ctagsCommand = env.Get("CTAGS_COMMAND", "universal-ctags", "ctags command (should point to universal-ctags executable compiled with JSON and seccomp support)")

// GetCommand returns the ctags command from the CTAGS_COMMAND environment
// variable, falling back to `universal-ctags`. Panics if the command doesn't
// exist.
func GetCommand() string {
	if !isCommandAvailable(ctagsCommand) {
		panic(fmt.Errorf("ctags command %s not found", ctagsCommand))
	}
	return ctagsCommand
}

func NewParser(ctagsCommand string) (Parser, error) {
	opt := "default"

	// TODO(sqs): Figure out why running with --_interactive=sandbox causes `Bad system call` inside Docker, and
	// reenable it.
	//
	// if runtime.GOOS == "linux" {
	//  opt = "sandbox"
	// }

	cmd := exec.Command(ctagsCommand, "--_interactive="+opt, "--fields=*",
		"--languages=Basic,C,C#,C++,Clojure,Cobol,CSS,CUDA,D,Elixir,elm,Erlang,Go,GraphQL,haskell,Java,JavaScript,kotlin,Lisp,Lua,MatLab,ObjectiveC,OCaml,Pascal,Perl,Perl6,PHP,Protobuf,Python,R,Ruby,Rust,scala,Scheme,Sh,swift,SystemVerilog,Tcl,typescript,tsx,Verilog,VHDL,Vim",
		"--map-CSS=+.scss", "--map-CSS=+.less", "--map-CSS=+.sass",
	)
	in, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}

	out, err := cmd.StdoutPipe()
	if err != nil {
		in.Close()
		return nil, err
	}
	cmd.Stderr = os.Stderr
	proc := ctagsProcess{
		cmd:     cmd,
		in:      in,
		out:     bufio.NewScanner(out),
		outPipe: out,
	}

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	var init reply
	if err := proc.read(&init); err != nil {
		proc.Close()
		return nil, err
	}

	if init.Typ == "error" {
		proc.Close()
		return nil, errors.Errorf("starting %s failed with: %s", ctagsCommand, init.Message)
	}

	return &proc, nil
}

type ctagsProcess struct {
	cmd     *exec.Cmd
	in      io.WriteCloser
	out     *bufio.Scanner
	outPipe io.ReadCloser
}

func (p *ctagsProcess) Close() {
	_ = p.cmd.Process.Kill()
	_ = p.outPipe.Close()
	_ = p.in.Close()
}

func (p *ctagsProcess) read(rep *reply) error {
	if !p.out.Scan() {
		err := p.out.Err()
		if err == nil {
			// p.out.Err() returns nil if the Scanner hit EOF,
			// but EOF is unexpected and means the process is bad and needs to be cleaned up
			err = errors.New("unexpected EOF from ctags")
		}
		return err
	}
	if debug {
		log.Printf("read %s", p.out.Text())
	}

	// See https://github.com/universal-ctags/ctags/issues/1493
	if bytes.Equal([]byte("(null)"), p.out.Bytes()) {
		return nil
	}

	err := json.Unmarshal(p.out.Bytes(), rep)
	if err != nil {
		return fmt.Errorf("unmarshal(%s): %v", p.out.Text(), err)
	}
	return nil
}

func (p *ctagsProcess) post(req *request, content []byte) error {
	body, err := json.Marshal(req)
	if err != nil {
		return err
	}
	body = append(body, '\n')
	if debug {
		log.Printf("post %q", body)
	}

	if _, err = p.in.Write(body); err != nil {
		return err
	}
	_, err = p.in.Write(content)
	if debug {
		log.Println(string(content))
	}
	return err
}

type request struct {
	Command  string `json:"command"`
	Filename string `json:"filename"`
	Size     int    `json:"size"`
}

type reply struct {
	// Init
	Typ     string `json:"_type"`
	Name    string `json:"name"`
	Version string `json:"version"`

	// completed
	Command string `json:"command"`

	Path      string `json:"path"`
	Language  string `json:"language"`
	Line      int    `json:"line"`
	Kind      string `json:"kind"`
	End       int    `json:"end"`
	Scope     string `json:"scope"`
	ScopeKind string `json:"scopeKind"`
	Access    string `json:"access"`
	File      bool   `json:"file"`
	Signature string `json:"signature"`
	Pattern   string `json:"pattern"`

	// error
	Message string `json:"message"`
}

func (p *ctagsProcess) Parse(name string, content []byte) (entries []Entry, err error) {
	req := request{
		Command:  "generate-tags",
		Size:     len(content),
		Filename: name,
	}

	if err := p.post(&req, content); err != nil {
		return nil, err
	}

	entries = make([]Entry, 0, 250)
	for {
		var rep reply
		if err := p.read(&rep); err != nil {
			return nil, err
		}
		if rep.Typ == "error" && logErrors {
			log.Printf("error parsing file %s: %s", name, rep.Message)
		}
		if rep.Typ == "completed" {
			break
		}

		entries = append(entries, Entry{
			Name:        rep.Name,
			Path:        rep.Path,
			Line:        rep.Line,
			Kind:        rep.Kind,
			Language:    rep.Language,
			Parent:      rep.Scope,
			ParentKind:  rep.ScopeKind,
			Pattern:     rep.Pattern,
			Signature:   rep.Signature,
			FileLimited: rep.File,
		})
	}

	return entries, nil
}
