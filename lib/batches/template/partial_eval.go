package template

import (
	"bytes"
	"fmt"
	"reflect"
	"strings"
	"text/template"
	"text/template/parse"
)

// IsStaticBool parses the input as a text/template and attempts to evaluate it
// with only the ahead-of-execution information available in StepContext.
//
// To do that it first calls parseAndPartialEval to evaluate the template as
// much as possible.
//
// If, after evaluation, more than text is left (i.e. because the template
// requires information that's only available later) the function returns with
// the first return value being false, because the template is not "static".
// first return value is true.
//
// If only text is left we check whether that text equals "true". The result of
// that check is the second return value.
func IsStaticBool(input string, ctx *StepContext) (isStatic bool, boolVal bool, err error) {
	t, err := parseAndPartialEval(input, ctx)
	if err != nil {
		return false, false, err
	}

	isStatic = true
	for _, n := range t.Tree.Root.Nodes {
		if n.Type() != parse.NodeText {
			isStatic = false
			break
		}
	}
	if !isStatic {
		return isStatic, false, nil
	}

	return true, isTrueOutput(t.Tree.Root), nil
}

// parseAndPartialEval parses input as a text/template and then attempts to
// partially evaluate the parts of the template it can evaluate ahead of time
// (meaning: before we've executed any batch spec steps and have a full
// StepContext available).
//
// If it's possible to evaluate a parse.ActionNode (which is what sits between
// delimiters in a text/template), the node is rewritten into a parse.TextNode,
// to make it look like it's always been text in the template.
//
// Partial evaluation is done in a best effort manner: if it's not possible to
// evaluate a node (because it requires information that we only later get, or
// because it's too complex, etc.) we degrade gracefully and simply abort the
// partial evaluation and leave the node as is.
//
// It also should be noted that we don't do "full" partial evaluation: if we
// come across value that we can't partially evaluate we abort the process *for
// the whole node* without replacing the sub-nodes that we've successfully
// evaluated. Why? Because we can't construct correct `*parse.Node` from
// outside the `parse` package. In other words: we evaluate
// all-parse.ActionNode-or-nothing.
func parseAndPartialEval(input string, ctx *StepContext) (*template.Template, error) {
	t, err := template.
		New("partial-eval").
		Delims(startDelim, endDelim).
		Funcs(builtins).
		Funcs(ctx.ToFuncMap()).
		Parse(input)

	if err != nil {
		return nil, err
	}

	for i, n := range t.Tree.Root.Nodes {
		t.Tree.Root.Nodes[i] = rewriteNode(n, ctx)
	}

	return t, nil
}

// rewriteNode takes the given parse.Parse and tries to partially evaluate it.
// If that's possible, the output of the evaluation is turned into text and
// instead of the node that was passed in a new parse.TextNode is returned that
// represents the output of the evaluation.
func rewriteNode(n parse.Node, ctx *StepContext) parse.Node {
	switch n := n.(type) {
	case *parse.ActionNode:
		if val, ok := evalPipe(ctx, n.Pipe); ok {
			var out bytes.Buffer
			fmt.Fprint(&out, val.Interface())
			return &parse.TextNode{
				Text:     out.Bytes(),
				Pos:      n.Pos,
				NodeType: parse.NodeText,
			}
		}

		return n

	default:
		return n
	}
}

// noValue is returned by the functions that partially evaluate a parse.Node
// to signify that evaluation was not possible or did not yield a value.
var noValue reflect.Value

func evalPipe(ctx *StepContext, p *parse.PipeNode) (finalVal reflect.Value, ok bool) {
	// If the pipe contains declaration we abort evaluation.
	if len(p.Decl) > 0 {
		return noValue, false
	}

	// TODO: Support finalVal and pass it in to evalCmd
	// finalVal is the value of the previous Cmd in a pipe (i.e. `${{ 3 + 3 | eq 6 }}`)
	// It needs to be the final (fixed) argument of a call if it's set.

	for _, c := range p.Cmds {
		finalVal, ok = evalCmd(ctx, c)
		if !ok {
			return noValue, false
		}
	}

	return finalVal, ok
}

func evalCmd(ctx *StepContext, c *parse.CommandNode) (reflect.Value, bool) {
	switch first := c.Args[0].(type) {
	case *parse.BoolNode, *parse.NumberNode, *parse.StringNode, *parse.ChainNode:
		if len(c.Args) == 1 {
			return evalNode(ctx, first)
		}
		return noValue, false

	case *parse.IdentifierNode:
		// A function call always starts with an identifier
		return evalFunction(ctx, first.Ident, c.Args)

	default:
		// Node type that we don't care about, so we don't even try to evaluate it
		return noValue, false
	}
}

func evalNode(ctx *StepContext, n parse.Node) (reflect.Value, bool) {
	switch n := n.(type) {
	case *parse.BoolNode:
		return reflect.ValueOf(n.True), true

	case *parse.NumberNode:
		// This case branch is lifted from Go's text/template execution engine:
		// https://sourcegraph.com/github.com/golang/go@2c9f5a1da823773c436f8b2c119635797d6db2d3/-/blob/src/text/template/exec.go#L493-530
		// The difference is that we don't do any error handling but simply abort.
		switch {
		case n.IsComplex:
			return reflect.ValueOf(n.Complex128), true

		case n.IsFloat &&
			!isHexInt(n.Text) && !isRuneInt(n.Text) &&
			strings.ContainsAny(n.Text, ".eEpP"):
			return reflect.ValueOf(n.Float64), true

		case n.IsInt:
			num := int(n.Int64)
			if int64(num) != n.Int64 {
				return noValue, false
			}
			return reflect.ValueOf(num), true

		case n.IsUint:
			return noValue, false
		}

	case *parse.StringNode:
		return reflect.ValueOf(n.Text), true

	case *parse.ChainNode:
		// For now we only support fields that are 1 level deep (see below).
		// Should we ever want to support more than one level, we need to
		// revise this.
		if len(n.Field) != 1 {
			return noValue, false
		}

		if ident, ok := n.Node.(*parse.IdentifierNode); ok {
			switch ident.Ident {
			case "repository":
				switch n.Field[0] {
				case "search_result_paths":
					// TODO: We don't eval search_result_paths for now, since it's a
					// "complex" value, a slice of strings, and turning that
					// into text might not be useful to the user. So we abort.
					return noValue, false
				case "name":
					return reflect.ValueOf(ctx.Repository.Name), true
				}

			case "batch_change":
				switch n.Field[0] {
				case "name":
					return reflect.ValueOf(ctx.BatchChange.Name), true
				case "description":
					return reflect.ValueOf(ctx.BatchChange.Description), true
				}
			}
		}
		return noValue, false

	case *parse.PipeNode:
		return evalPipe(ctx, n)
	}

	return noValue, false
}

func isRuneInt(s string) bool {
	return len(s) > 0 && s[0] == '\''
}

func isHexInt(s string) bool {
	return len(s) > 2 && s[0] == '0' && (s[1] == 'x' || s[1] == 'X') && !strings.ContainsAny(s, "pP")
}

func evalFunction(ctx *StepContext, name string, args []parse.Node) (val reflect.Value, success bool) {
	defer func() {
		if r := recover(); r != nil {
			val = noValue
			success = false
		}
	}()

	switch name {
	case "eq":
		return evalEqCall(ctx, args[1:])

	case "ne":
		equal, ok := evalEqCall(ctx, args[1:])
		if !ok {
			return noValue, false
		}
		return reflect.ValueOf(!equal.Bool()), true

	case "not":
		return evalNotCall(ctx, args[1:])

	default:
		concreteFn, ok := builtins[name]
		if !ok {
			return noValue, false
		}

		fn := reflect.ValueOf(concreteFn)

		// We can eval only if all args are static:
		var evaluatedArgs []reflect.Value
		for _, a := range args[1:] {
			v, ok := evalNode(ctx, a)
			if !ok {
				// One of the args is not static, abort
				return noValue, false
			}
			evaluatedArgs = append(evaluatedArgs, v)

		}

		ret := fn.Call(evaluatedArgs)
		if len(ret) == 2 && !ret[1].IsNil() {
			return noValue, false
		}
		return ret[0], true
	}
}

func evalNotCall(ctx *StepContext, args []parse.Node) (reflect.Value, bool) {
	// We only support 1 arg for now:
	if len(args) != 1 {
		return noValue, false
	}

	arg, ok := evalNode(ctx, args[0])
	if !ok {
		return noValue, false
	}

	return reflect.ValueOf(!isTrue(arg)), true
}

func evalEqCall(ctx *StepContext, args []parse.Node) (reflect.Value, bool) {
	// We only support 2 args for now:
	if len(args) != 2 {
		return noValue, false
	}

	// We only eval `eq` if all args are static:
	var evaluatedArgs []reflect.Value
	for _, a := range args {
		v, ok := evalNode(ctx, a)
		if !ok {
			// One of the args is not static, abort
			return noValue, false
		}
		evaluatedArgs = append(evaluatedArgs, v)
	}

	if len(evaluatedArgs) != 2 {
		// safety check
		return noValue, false
	}

	isEqual := evaluatedArgs[0].Interface() == evaluatedArgs[1].Interface()
	return reflect.ValueOf(isEqual), true
}

// isTrue is taken from Go's text/template/exec.go and simplified
func isTrue(val reflect.Value) (truth bool) {
	if !val.IsValid() {
		// Something like var x interface{}, never set. It's a form of nil.
		return false
	}
	switch val.Kind() {
	case reflect.Array, reflect.Map, reflect.Slice, reflect.String:
		return val.Len() > 0
	case reflect.Bool:
		return val.Bool()
	case reflect.Complex64, reflect.Complex128:
		return val.Complex() != 0
	case reflect.Chan, reflect.Func, reflect.Ptr, reflect.Interface:
		return !val.IsNil()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return val.Int() != 0
	case reflect.Float32, reflect.Float64:
		return val.Float() != 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return val.Uint() != 0
	case reflect.Struct:
		return true // Struct values are always true.
	default:
		return false
	}
}
