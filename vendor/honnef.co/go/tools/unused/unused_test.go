package unused

// Copyright (c) 2013 The Go Authors. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found at
// https://developers.google.com/open-source/licenses/bsd.

import (
	"go/parser"
	"go/token"
	"strings"
	"testing"

	"honnef.co/go/tools/lint/testutil"
)

func TestAll(t *testing.T) {
	checker := NewChecker(CheckAll)
	l := NewLintChecker(checker)
	testutil.TestAll(t, l, "")
}

type instruction struct {
	Line int // the line number this applies to
	IDs  []string
}

// parseInstructions parses instructions from the comments in a Go source file.
// It returns nil if none were parsed.
func parseInstructions(t *testing.T, filename string, src []byte) []instruction {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, filename, src, parser.ParseComments)
	if err != nil {
		t.Fatalf("Test file %v does not parse: %v", filename, err)
	}
	var ins []instruction
	for _, cg := range f.Comments {
		ln := fset.Position(cg.Pos()).Line
		raw := cg.Text()
		for _, line := range strings.Split(raw, "\n") {
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			if strings.Contains(line, "MATCH ") {
				ids := extractIDs(line)
				ins = append(ins, instruction{
					Line: ln,
					IDs:  ids,
				})
			}
		}
	}
	return ins
}

func extractIDs(line string) []string {
	const marker = "MATCH "
	idx := strings.Index(line, marker)
	line = line[idx+len(marker):]
	return strings.Split(line, ", ")
}
