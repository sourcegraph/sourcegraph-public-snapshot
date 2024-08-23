/* Copyright 2022 The Bazel Authors. All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

   http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package golang

import (
	"bufio"
	"bytes"
	"fmt"
	"go/build/constraint"
	"os"
	"strings"
)

// readTags reads and extracts build tags from the block of comments
// and blank lines at the start of a file which is separated from the
// rest of the file by a blank line. Each string in the returned slice
// is the trimmed text of a line after a "+build" prefix.
// Based on go/build.Context.shouldBuild.
func readTags(path string) (*buildTags, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	content, err := readComments(f)
	if err != nil {
		return nil, err
	}

	content, goBuild, _, err := parseFileHeader(content)
	if err != nil {
		return nil, err
	}

	if goBuild != nil {
		x, err := constraint.Parse(string(goBuild))
		if err != nil {
			return nil, err
		}

		return newBuildTags(x)
	}

	var fullConstraint constraint.Expr
	// Search and parse +build tags
	scanner := bufio.NewScanner(bytes.NewReader(content))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if !constraint.IsPlusBuild(line) {
			continue
		}

		x, err := constraint.Parse(line)
		if err != nil {
			return nil, err
		}

		if fullConstraint != nil {
			fullConstraint = &constraint.AndExpr{
				X: fullConstraint,
				Y: x,
			}
		} else {
			fullConstraint = x
		}
	}

	if scanner.Err() != nil {
		return nil, scanner.Err()
	}

	if fullConstraint == nil {
		return nil, nil
	}

	return newBuildTags(fullConstraint)
}

// buildTags represents the build tags specified in a file.
type buildTags struct {
	// expr represents the parsed constraint expression
	// that can be used to evaluate a file against a set
	// of tags.
	expr constraint.Expr
	// rawTags represents the concrete tags that make up expr.
	rawTags []string
}

// newBuildTags will return a new buildTags structure with any
// ignored tags filtered out from the provided constraints.
func newBuildTags(x constraint.Expr) (*buildTags, error) {
	modified, err := dropNegationForIgnoredTags(pushNot(x, false))
	if err != nil {
		return nil, err
	}

	rawTags, err := collectTags(modified)
	if err != nil {
		return nil, err
	}

	return &buildTags{
		expr:    modified,
		rawTags: rawTags,
	}, nil
}

func (b *buildTags) tags() []string {
	if b == nil {
		return nil
	}

	return b.rawTags
}

func (b *buildTags) eval(ok func(string) bool) bool {
	if b == nil || b.expr == nil {
		return true
	}

	return b.expr.Eval(ok)
}

func (b *buildTags) empty() bool {
	if b == nil {
		return true
	}

	return len(b.rawTags) == 0
}

// dropNegationForIgnoredTags drops negations for any concrete tags that should be ignored.
// This is done to ensure that when ignored tags are evaluated, they can always return true
// without having to worry that the result will be negated later on. Ignored tags should always
// evaluate to true, regardless of whether they are negated or not leaving the final evaluation
// to happen at compile time by the compiler.
func dropNegationForIgnoredTags(expr constraint.Expr) (constraint.Expr, error) {
	if expr == nil {
		return nil, nil
	}

	switch x := expr.(type) {
	case *constraint.TagExpr:
		return &constraint.TagExpr{
			Tag: x.Tag,
		}, nil

	case *constraint.NotExpr:
		var toRet constraint.Expr
		// flip nots on any ignored tags
		if tag, ok := x.X.(*constraint.TagExpr); ok && isIgnoredTag(tag.Tag) {
			toRet = &constraint.TagExpr{
				Tag: tag.Tag,
			}
		} else {
			fixed, err := dropNegationForIgnoredTags(x.X)
			if err != nil {
				return nil, err
			}
			toRet = &constraint.NotExpr{X: fixed}
		}

		return toRet, nil

	case *constraint.AndExpr:
		a, err := dropNegationForIgnoredTags(x.X)
		if err != nil {
			return nil, err
		}

		b, err := dropNegationForIgnoredTags(x.Y)
		if err != nil {
			return nil, err
		}

		return &constraint.AndExpr{
			X: a,
			Y: b,
		}, nil

	case *constraint.OrExpr:
		a, err := dropNegationForIgnoredTags(x.X)
		if err != nil {
			return nil, err
		}

		b, err := dropNegationForIgnoredTags(x.Y)
		if err != nil {
			return nil, err
		}

		return &constraint.OrExpr{
			X: a,
			Y: b,
		}, nil

	default:
		return nil, fmt.Errorf("unknown constraint type: %T", x)
	}
}

// filterTags will traverse the provided constraint.Expr, recursively, and call
// the user provided ok func on concrete constraint.TagExpr structures. If the provided
// func returns true, the tag in question is kept, otherwise it is filtered out.
func visitTags(expr constraint.Expr, visit func(string)) (err error) {
	if expr == nil {
		return nil
	}

	switch x := expr.(type) {
	case *constraint.TagExpr:
		visit(x.Tag)

	case *constraint.NotExpr:
		err = visitTags(x.X, visit)

	case *constraint.AndExpr:
		err = visitTags(x.X, visit)
		if err == nil {
			err = visitTags(x.Y, visit)
		}

	case *constraint.OrExpr:
		err = visitTags(x.X, visit)
		if err == nil {
			err = visitTags(x.Y, visit)
		}

	default:
		return fmt.Errorf("unknown constraint type: %T", x)
	}

	return
}

func collectTags(expr constraint.Expr) ([]string, error) {
	var tags []string
	err := visitTags(expr, func(tag string) {
		tags = append(tags, tag)
	})
	if err != nil {
		return nil, err
	}

	return tags, err
}

// cgoTagsAndOpts contains compile or link options which should only be applied
// if the given set of build tags are satisfied. These options have already
// been tokenized using the same algorithm that "go build" uses, then joined
// with OptSeparator.
type cgoTagsAndOpts struct {
	*buildTags
	opts string
}

func (c *cgoTagsAndOpts) tags() []string {
	if c == nil {
		return nil
	}

	return c.buildTags.tags()
}

func (c *cgoTagsAndOpts) eval(ok func(string) bool) bool {
	if c == nil {
		return true
	}

	return c.buildTags.eval(ok)
}

// matchAuto interprets text as either a +build or //go:build expression (whichever works).
// Forked from go/build.Context.matchAuto
func matchAuto(tokens []string) (*buildTags, error) {
	if len(tokens) == 0 {
		return nil, nil
	}

	text := strings.Join(tokens, " ")
	if strings.ContainsAny(text, "&|()") {
		text = "//go:build " + text
	} else {
		text = "// +build " + text
	}

	x, err := constraint.Parse(text)
	if err != nil {
		return nil, err
	}

	return newBuildTags(x)
}

// isIgnoredTag returns whether the tag is "cgo" or is a release tag.
// Release tags match the pattern "go[0-9]\.[0-9]+".
// Gazelle won't consider whether an ignored tag is satisfied when evaluating
// build constraints for a file and will instead defer to the compiler at compile
// time.
func isIgnoredTag(tag string) bool {
	if tag == "cgo" || tag == "race" || tag == "msan" {
		return true
	}
	if len(tag) < 5 || !strings.HasPrefix(tag, "go") {
		return false
	}
	if tag[2] < '0' || tag[2] > '9' || tag[3] != '.' {
		return false
	}
	for _, c := range tag[4:] {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

// pushNot applies DeMorgan's law to push negations down the expression,
// so that only tags are negated in the result.
// (It applies the rewrites !(X && Y) => (!X || !Y) and !(X || Y) => (!X && !Y).)
// Forked from go/build/constraint.pushNot
func pushNot(x constraint.Expr, not bool) constraint.Expr {
	switch x := x.(type) {
	default:
		// unreachable
		return x
	case *constraint.NotExpr:
		if _, ok := x.X.(*constraint.TagExpr); ok && !not {
			return x
		}
		return pushNot(x.X, !not)
	case *constraint.TagExpr:
		if not {
			return &constraint.NotExpr{X: x}
		}
		return x
	case *constraint.AndExpr:
		x1 := pushNot(x.X, not)
		y1 := pushNot(x.Y, not)
		if not {
			return or(x1, y1)
		}
		if x1 == x.X && y1 == x.Y {
			return x
		}
		return and(x1, y1)
	case *constraint.OrExpr:
		x1 := pushNot(x.X, not)
		y1 := pushNot(x.Y, not)
		if not {
			return and(x1, y1)
		}
		if x1 == x.X && y1 == x.Y {
			return x
		}
		return or(x1, y1)
	}
}

func or(x, y constraint.Expr) constraint.Expr {
	return &constraint.OrExpr{X: x, Y: y}
}

func and(x, y constraint.Expr) constraint.Expr {
	return &constraint.AndExpr{X: x, Y: y}
}
