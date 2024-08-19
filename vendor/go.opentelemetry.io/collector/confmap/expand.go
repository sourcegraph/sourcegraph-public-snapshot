// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package confmap // import "go.opentelemetry.io/collector/confmap"

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"go.opentelemetry.io/collector/confmap/internal"
)

// schemePattern defines the regexp pattern for scheme names.
// Scheme name consist of a sequence of characters beginning with a letter and followed by any
// combination of letters, digits, plus ("+"), period ("."), or hyphen ("-").
const schemePattern = `[A-Za-z][A-Za-z0-9+.-]+`

var (
	// Need to match new line as well in the OpaqueValue, so setting the "s" flag. See https://pkg.go.dev/regexp/syntax.
	uriRegexp = regexp.MustCompile(`(?s:^(?P<Scheme>` + schemePattern + `):(?P<OpaqueValue>.*)$)`)

	errTooManyRecursiveExpansions = errors.New("too many recursive expansions")
)

func (mr *Resolver) expandValueRecursively(ctx context.Context, value any) (any, error) {
	for i := 0; i < 100; i++ {
		val, changed, err := mr.expandValue(ctx, value)
		if err != nil {
			return nil, err
		}
		if !changed {
			return val, nil
		}
		value = val
	}
	return nil, errTooManyRecursiveExpansions
}

func (mr *Resolver) expandValue(ctx context.Context, value any) (any, bool, error) {
	switch v := value.(type) {
	case string:
		if !strings.Contains(v, "${") || !strings.Contains(v, "}") {
			// No URIs to expand.
			return value, false, nil
		}
		// Embedded or nested URIs.
		return mr.findAndExpandURI(ctx, v)
	case []any:
		nslice := make([]any, 0, len(v))
		nchanged := false
		for _, vint := range v {
			val, changed, err := mr.expandValue(ctx, vint)
			if err != nil {
				return nil, false, err
			}
			nslice = append(nslice, val)
			nchanged = nchanged || changed
		}
		return nslice, nchanged, nil
	case map[string]any:
		nmap := map[string]any{}
		nchanged := false
		for mk, mv := range v {
			val, changed, err := mr.expandValue(ctx, mv)
			if err != nil {
				return nil, false, err
			}
			nmap[mk] = val
			nchanged = nchanged || changed
		}
		return nmap, nchanged, nil
	}
	return value, false, nil
}

// findURI attempts to find the first potentially expandable URI in input. It returns a potentially expandable
// URI, or an empty string if none are found.
// Note: findURI is only called when input contains a closing bracket.
func (mr *Resolver) findURI(input string) string {
	closeIndex := strings.Index(input, "}")
	remaining := input[closeIndex+1:]
	openIndex := strings.LastIndex(input[:closeIndex+1], "${")

	// if there is any of:
	//  - a missing "${"
	//  - there is no default scheme AND no scheme is detected because no `:` is found.
	// then check the next URI.
	if openIndex < 0 || (mr.defaultScheme == "" && !strings.Contains(input[openIndex:closeIndex+1], ":")) {
		// if remaining does not contain "}", there are no URIs left: stop recursion.
		if !strings.Contains(remaining, "}") {
			return ""
		}
		return mr.findURI(remaining)
	}

	return input[openIndex : closeIndex+1]
}

// findAndExpandURI attempts to find and expand the first occurrence of an expandable URI in input. If an expandable URI is found it
// returns the input with the URI expanded, true and nil. Otherwise, it returns the unchanged input, false and the expanding error.
// This method expects input to start with ${ and end with }
func (mr *Resolver) findAndExpandURI(ctx context.Context, input string) (any, bool, error) {
	uri := mr.findURI(input)
	if uri == "" {
		// No URI found, return.
		return input, false, nil
	}
	if uri == input {
		// If the value is a single URI, then the return value can be anything.
		// This is the case `foo: ${file:some_extra_config.yml}`.
		ret, err := mr.expandURI(ctx, input)
		if err != nil {
			return input, false, err
		}

		expanded, err := ret.AsRaw()
		if err != nil {
			return input, false, err
		}
		return expanded, true, err
	}
	expanded, err := mr.expandURI(ctx, uri)
	if err != nil {
		return input, false, err
	}

	var repl string
	if internal.StrictlyTypedInputGate.IsEnabled() {
		repl, err = expanded.AsString()
	} else {
		repl, err = toString(expanded)
	}
	if err != nil {
		return input, false, fmt.Errorf("expanding %v: %w", uri, err)
	}
	return strings.ReplaceAll(input, uri, repl), true, err
}

// toString attempts to convert input to a string.
func toString(ret *Retrieved) (string, error) {
	// This list must be kept in sync with checkRawConfType.
	input, err := ret.AsRaw()
	if err != nil {
		return "", err
	}

	val := reflect.ValueOf(input)
	switch val.Kind() {
	case reflect.String:
		return val.String(), nil
	case reflect.Int, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(val.Int(), 10), nil
	case reflect.Float32, reflect.Float64:
		return strconv.FormatFloat(val.Float(), 'f', -1, 64), nil
	case reflect.Bool:
		return strconv.FormatBool(val.Bool()), nil
	default:
		return "", fmt.Errorf("expected convertable to string value type, got %q(%T)", input, input)
	}
}

func (mr *Resolver) expandURI(ctx context.Context, input string) (*Retrieved, error) {
	// strip ${ and }
	uri := input[2 : len(input)-1]

	if !strings.Contains(uri, ":") {
		uri = fmt.Sprintf("%s:%s", mr.defaultScheme, uri)
	}

	lURI, err := newLocation(uri)
	if err != nil {
		return nil, err
	}

	if strings.Contains(lURI.opaqueValue, "$") {
		return nil, fmt.Errorf("the uri %q contains unsupported characters ('$')", lURI.asString())
	}
	ret, err := mr.retrieveValue(ctx, lURI)
	if err != nil {
		return nil, err
	}
	mr.closers = append(mr.closers, ret.Close)
	return ret, nil
}

type location struct {
	scheme      string
	opaqueValue string
}

func (c location) asString() string {
	return c.scheme + ":" + c.opaqueValue
}

func newLocation(uri string) (location, error) {
	submatches := uriRegexp.FindStringSubmatch(uri)
	if len(submatches) != 3 {
		return location{}, fmt.Errorf("invalid uri: %q", uri)
	}
	return location{scheme: submatches[1], opaqueValue: submatches[2]}, nil
}
