// Package ctags provides a Go wrapper for universal-ctags.
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
	"path"
	"strings"
	"unicode/utf8"
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

type Parser interface {
	Parse(path string, content []byte) ([]*Entry, error)
	Close()
}

type Options struct {
	// Bin is the command to run. Defaults to "universal-ctags" if empty.
	Bin string

	// PatternLengthLimit is the cutoff length of the patterns output by
	// ctags. (--pattern-length-limit). If 0 defaults to 255.
	//
	// Note: --pattern-length-limit=0 disables this in universal-ctags. We don't
	// allow disabling it.
	PatternLengthLimit int

	// Info if non-nil will log info messages
	Info *log.Logger

	// Debug if non-nil will log debug messages
	Debug *log.Logger
}

func New(opts Options) (Parser, error) {
	if opts.Bin == "" {
		opts.Bin = "universal-ctags"
	}
	if opts.PatternLengthLimit == 0 {
		opts.PatternLengthLimit = 255
	}

	// ctagsArgs is the contents of ctags.d. ctags.d needs to be in a specific
	// location to be read correctly. We have accidently regressed on this
	// twice. Instead we pass in the arguments here.
	args := []string{"--_interactive=default", "--fields=*", fmt.Sprintf("--pattern-length-limit=%d", opts.PatternLengthLimit)}
	args = append(args, ctagsArgs...)

	// Some languages cause issues in universal-ctags. Stick to an allowlist of
	// known working languages.
	args = append(args, "--languages="+strings.Join(SupportedLanguages[:], ","))

	cmd := exec.Command(opts.Bin, args...)
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
		out:     &scanner{r: bufio.NewReaderSize(out, 4096)},
		outPipe: out,

		Info:  opts.Info,
		Debug: opts.Debug,
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
		return nil, fmt.Errorf("starting %s failed with: %s", opts.Bin, init.Message)
	}

	return &proc, nil
}

type ctagsProcess struct {
	cmd     *exec.Cmd
	in      io.WriteCloser
	out     *scanner
	outPipe io.ReadCloser

	Info  *log.Logger
	Debug *log.Logger
}

func (p *ctagsProcess) Close() {
	_ = p.cmd.Process.Kill()
	_ = p.outPipe.Close()
	_ = p.in.Close()
	_ = p.cmd.Wait()
}

func (p *ctagsProcess) read(rep *reply) error {
	if !p.out.Scan() {
		// Some errors do not kill the parser. We would deadlock if we waited
		// for the process to exit.
		err := p.out.Err()
		p.Close()
		return err
	}
	if p.Debug != nil {
		p.Debug.Printf("read %q", p.out.Bytes())
	}

	// See https://github.com/universal-ctags/ctags/issues/1493
	if bytes.Equal([]byte("(null)"), p.out.Bytes()) {
		return nil
	}

	err := json.Unmarshal(p.out.Bytes(), rep)
	if err != nil {
		return fmt.Errorf("unmarshal(%q): %v", p.out.Bytes(), err)
	}
	return nil
}

// universal-ctags line buffer size is only 1024.
const ctagsLineBufferSize = 1024

func (p *ctagsProcess) post(req *request, content []byte) (bool, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return false, err
	}
	body = append(body, '\n')

	// -1 for c-style string
	if len(body) > ctagsLineBufferSize-1 {
		return false, nil
	}

	if p.Debug != nil {
		p.Debug.Printf("post %q", body)
	}

	if _, err = p.in.Write(body); err != nil {
		return false, err
	}

	_, err = p.in.Write(content)
	if p.Debug != nil {
		p.Debug.Println(string(content))
	}
	return err == nil, err
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

	// error
	Message string `json:"message"`
	Fatal   bool   `json:"fatal"`

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
}

func (p *ctagsProcess) Parse(name string, content []byte) ([]*Entry, error) {
	filename := path.Base(name)
	req := request{
		Command:  "generate-tags",
		Size:     len(content),
		Filename: filename,
	}

	if !utf8.Valid(content) {
		if p.Info != nil {
			p.Info.Printf("ctags skipping file due not being utf-8 encoded: %s", name)
		}
		return nil, nil
	}

	if ok, err := p.post(&req, content); err != nil {
		return nil, err
	} else if !ok {
		if p.Info != nil {
			p.Info.Printf("ctags skipping file due to long filename: %s", name)
		}
		return nil, nil
	}

	// 250 is a better guess for initial size
	es := make([]*Entry, 0, 250)
	for {
		var rep reply
		if err := p.read(&rep); err != nil {
			return nil, err
		}
		switch rep.Typ {
		case "completed":
			return es, nil
		case "error":
			if rep.Fatal {
				return nil, fmt.Errorf("fatal ctags error for %s: %s", name, rep.Message)
			} else if p.Info != nil {
				p.Info.Printf("ignoring non-fatal ctags error for %s: %s", name, rep.Message)
			}
		case "tag":
			if rep.Path == filename {
				rep.Path = name
			}
			es = append(es, &Entry{
				Name:       rep.Name,
				Path:       rep.Path,
				Line:       rep.Line,
				Kind:       rep.Kind,
				Language:   rep.Language,
				Parent:     rep.Scope,
				ParentKind: rep.ScopeKind,
				Pattern:    rep.Pattern,
				Signature:  rep.Signature,
			})
		default:
			return nil, fmt.Errorf("ctags unexpected response %s for %s", rep.Typ, name)
		}
	}
}

// scanner is like bufio.Scanner but skips long lines instead of returning
// bufio.ErrTooLong.
//
// Additionally it will skip empty lines.
type scanner struct {
	r    *bufio.Reader
	line []byte
	err  error
}

func (s *scanner) Scan() bool {
	if s.err != nil {
		return false
	}

	var (
		err  error
		line []byte
	)

	for err == nil && len(line) == 0 {
		line, err = s.r.ReadSlice('\n')
		for err == bufio.ErrBufferFull {
			// make line empty so we ignore it
			line = nil
			_, err = s.r.ReadSlice('\n')
		}
		line = bytes.TrimSuffix(line, []byte{'\n'})
		line = bytes.TrimSuffix(line, []byte{'\r'})
	}

	s.line, s.err = line, err
	return len(line) > 0
}

func (s *scanner) Bytes() []byte {
	return s.line
}

func (s *scanner) Err() error {
	if s.err == io.EOF {
		return nil
	}
	return s.err
}
