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
	"fmt"
	"log"
	"os"
	"time"

	goctags "github.com/sourcegraph/go-ctags"
)

type Entry = goctags.Entry

// CTagsParser wraps go-ctags and delegates to the right process (like universal-ctags or scip-ctags).
// It is only safe for single-threaded use. This wrapper also enforces a timeout on parsing a single
// document, which is important since documents can occasionally hang universal-ctags.
// documents which hang universal-ctags.
type CTagsParser struct {
	bins    ParserBinMap
	parsers map[CTagsParserType]goctags.Parser
}

// parseTimeout is how long we wait for a response for parsing a single file
// in ctags. 1 minute is a very conservative timeout which we should only hit
// if ctags hangs.
const parseTimeout = time.Minute

func NewCTagsParser(bins ParserBinMap) CTagsParser {
	return CTagsParser{bins: bins, parsers: make(map[CTagsParserType]goctags.Parser)}
}

type parseResult struct {
	entries []*Entry
	err     error
}

func (lp *CTagsParser) Parse(name string, content []byte, typ CTagsParserType) ([]*Entry, error) {
	if lp.parsers[typ] == nil {
		parser, err := lp.newParserProcess(typ)
		if parser == nil || err != nil {
			return nil, err
		}
		lp.parsers[typ] = parser
	}

	deadline := time.NewTimer(parseTimeout)
	defer deadline.Stop()

	parser := lp.parsers[typ]
	recv := make(chan parseResult, 1)
	go func() {
		entry, err := parser.Parse(name, content)
		recv <- parseResult{entries: entry, err: err}
	}()

	select {
	case resp := <-recv:
		return resp.entries, resp.err
	case <-deadline.C:
		// Error out since ctags hanging is a sign something bad is happening.
		return nil, fmt.Errorf("ctags timedout after %s parsing %s", parseTimeout, name)
	}
}

func (lp *CTagsParser) newParserProcess(typ CTagsParserType) (goctags.Parser, error) {
	bin := lp.bins[typ]
	if bin == "" {
		// This happens if CTagsMustSucceed is false and we didn't find the binary
		return nil, nil
	}

	opts := goctags.Options{Bin: bin}
	parserType := ParserToString(typ)
	if debug {
		opts.Info = log.New(os.Stderr, "CTAGS ("+parserType+") INF: ", log.LstdFlags)
		opts.Debug = log.New(os.Stderr, "CTAGS ("+parserType+") DBG: ", log.LstdFlags)
	}
	return goctags.New(opts)
}

func (lp *CTagsParser) Close() {
	for _, parser := range lp.parsers {
		parser.Close()
	}
}
