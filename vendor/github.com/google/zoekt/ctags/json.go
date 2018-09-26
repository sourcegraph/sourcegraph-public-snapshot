// Copyright 2017 Google Inc. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
	"runtime"
	"strings"
	"sync"
)

const debug = false

type ctagsProcess struct {
	cmd     *exec.Cmd
	in      io.WriteCloser
	out     *bufio.Scanner
	outPipe io.ReadCloser
}

func newProcess(bin string) (*ctagsProcess, error) {
	opt := "default"
	if runtime.GOOS == "linux" {
		opt = "sandbox"
	}

	cmd := exec.Command(bin, "--_interactive="+opt, "--fields=*")
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
		return nil, err
	}

	return &proc, nil
}

func (p *ctagsProcess) Close() {
	p.cmd.Process.Kill()
	p.outPipe.Close()
	p.in.Close()
}

func (p *ctagsProcess) read(rep *reply) error {
	if !p.out.Scan() {
		// capture exit error.
		err := p.cmd.Wait()
		p.outPipe.Close()
		p.in.Close()
		return err
	}
	if debug {
		log.Printf("read %s", p.out.Text())
	}

	// See https://github.com/universal-ctags/ctags/issues/1493
	if bytes.Compare([]byte("(null)"), p.out.Bytes()) == 0 {
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

	// Ignore pattern: we don't use it and universal-ctags
	// sometimes generates 'false' as value.
	Path      string `json:"path"`
	Language  string `json:"language"`
	Line      int    `json:"line"`
	Kind      string `json:"kind"`
	End       int    `json:"end"`
	Scope     string `json:"scope"`
	ScopeKind string `json:"scopeKind"`
	Access    string `json:"access"`
	Signature string `json:"signature"`
}

func (p *ctagsProcess) Parse(name string, content []byte) ([]*Entry, error) {
	req := request{
		Command:  "generate-tags",
		Size:     len(content),
		Filename: name,
	}

	if err := p.post(&req, content); err != nil {
		return nil, err
	}

	var es []*Entry
	for {
		var rep reply
		if err := p.read(&rep); err != nil {
			return nil, err
		}
		if rep.Typ == "completed" {
			break
		}

		e := Entry{
			Sym:      rep.Name,
			Path:     rep.Path,
			Line:     rep.Line,
			Kind:     rep.Kind,
			Language: rep.Language,
		}

		es = append(es, &e)
	}

	return es, nil
}

type Parser interface {
	Parse(name string, content []byte) ([]*Entry, error)
}

type lockedParser struct {
	p Parser
	l sync.Mutex
}

func (lp *lockedParser) Parse(name string, content []byte) ([]*Entry, error) {
	lp.l.Lock()
	defer lp.l.Unlock()
	return lp.p.Parse(name, content)
}

// NewParser creates a parser that is implemented by the given
// universal-ctags binary. The parser is safe for concurrent use.
func NewParser(bin string) (Parser, error) {
	if strings.Contains(bin, "universal-ctags") {
		// todo: restart, parallelization.
		proc, err := newProcess(bin)
		if err != nil {
			return nil, err
		}
		return &lockedParser{p: proc}, nil
	}

	log.Fatal("not implemented")
	return nil, nil
}
