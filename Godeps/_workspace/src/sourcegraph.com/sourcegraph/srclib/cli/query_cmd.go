package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"sort"

	"sourcegraph.com/sourcegraph/srclib/graph"

	"github.com/peterh/liner"
)

func init() {
	queryGroup, err := CLI.AddCommand("query",
		"query REPL for build data",
		"The query (q) command is a readline-like interface for interacting with the build data for a project.",
		&queryCmd,
	)
	if err != nil {
		log.Fatal(err)
	}
	queryGroup.Aliases = append(queryGroup.Aliases, "q")
}

type QueryCmd struct {
	Args struct {
		Rest []string `name:"ARGS"`
	} `positional-args:"yes"`
}

var queryCmd QueryCmd

var historyFile = "/tmp/.srclibq_history"

var activeContext commandContext

func (c *QueryCmd) Execute(args []string) error {
	if err := setActiveContext("."); err != nil {
		// TODO: log error somewhere
		log.Println("Errors were found building this project. Some things may be broken. Continuing...")
	}
	if len(c.Args.Rest) != 0 {
		// If args are provided, evaluate the args and do not
		// enter the interactive interface.
		output, err := eval(strings.Join(c.Args.Rest, " "))
		// Always print output, even if err is non-nil.
		if output != "" {
			fmt.Print(cleanOutput(output))
		}
		if err != nil {
			return err
		}
		return nil
	}
	fmt.Println(`src query - ":help" for help`)
	term, err := terminal(historyFile)
	if err != nil {
		return err
	}
	defer persist(term, historyFile)
	term.SetWordCompleter(wordCompleter)
	term.SetTabCompletionStyle(liner.TabPrints)

	for {
		line, err := term.Prompt("src> ")
		if err != nil {
			if err == io.EOF {
				fmt.Println()
				return nil
			}
			return err
		}
		term.AppendHistory(line)
		output, err := eval(line)
		if err != nil {
			fmt.Println("Error:", err)
			if output != "" {
				fmt.Print(cleanOutput(output))
			}
		} else {
			fmt.Print(cleanOutput(output))
		}
	}
}

func setActiveContext(repoPath string) error {
	fmt.Printf("Analyzing project...")
	// Build project concurrently so we can update the UI.
	type maybeContext struct {
		context commandContext
		err     error
	}
	done := make(chan maybeContext)
	go func() {
		context, err := prepareCommandContext(repoPath)
		done <- maybeContext{context, err}
	}()
OuterLoop:
	for {
		select {
		case <-time.Tick(time.Second):
			fmt.Print(".")
		case m := <-done:
			if m.err != nil {
				fmt.Println()
				return m.err
			}
			activeContext = m.context
			break OuterLoop
		}
	}
	fmt.Println()
	// Invariant: activeContext is the result of prepareCommandContext
	// after the loop above.
	return nil
}

// matchWithKeyword matches lines that include a colon, ':'.
//
// Groups:
//   1. line prefix
//   2. keyword name
//   3. whitespace between keyword name and value
//   4. value
//
// If matchWithKeyword does not match, then the input matches all
// valid values for the implicit keyword, ":name", that are prefixed
// by input.
//
// If group two is empty, then the last character is ':' and the input
// matches all keywords.
//
// If group three is empty, then the input matches all keywords beginning
// with group two.
//
// If group three is not empty and group four is empty, then the input
// matches all values that are valid for the keyword named by group
// two.
//
// If group four is not empty, then the input matches all valid values
// for the keywored named by group one that have group four as their
// prefix.
var matchWithKeyword = regexp.MustCompile(`(.*:)(\w*)([[:blank:]]*)(.*)`)

// wordCompleter returns a set of completions for word ending at line[pos].
func wordCompleter(line string, pos int) (head string, completions []string, tail string) {
	// 3/23/15: liner's Bash-style completion deletes the
	// to-be-completed word if completions has more than one item.
	// We fix that with fix by manually appending the completed
	// section of the token to "head".
	//
	// NOTE: check that completion works when len(completions) == 0.
	fix := func(head string, completions []string, tail string) (string, []string, string) {
		// When completions empty, the word completer is not
		// applied. The bug does not manifest when completions
		// has only one item.
		if len(completions) == 0 || len(completions) == 1 {
			return head, completions, tail
		}
		// invariant: prefix is the common case-insensitive
		// prefix among all of the completions seen so far
		// after each loop iteration.
		prefix := strings.ToLower(completions[1])
		for i := 2; i < len(completions); i++ {
			c := completions[i]
			l := len(prefix)
			if len(c) < l {
				l = len(c)
			}
			for ; l >= 0; l-- {
				if c[:l] == prefix[:l] {
					break
				}
			}
			prefix = c[:l]
			// Short-circuit if prefix is empty.
			if prefix == "" {
				break
			}
		}
		return head + prefix, completions, tail
	}
	seg := line[:pos]
	end := line[pos:]
	m := matchWithKeyword.FindStringSubmatch(seg)
	// If no match was found, the keyword is "name" implicitly.
	if m == nil {
		return fix("", valueCompleter(string(keyName), seg), end)
	}
	if m[2] == "" || m[3] == "" {
		return fix(m[1], keywordCompleter(m[2]), end)
	}
	return fix(m[1], valueCompleter(m[2], m[4]), end)
}

// keywordCompleter returns a set of keywords that complete token.
func keywordCompleter(token string) []string {
	var cs []string
	for _, k := range allKeywords {
		if strings.HasPrefix(string(k), token) {
			cs = append(cs, string(k))
		}
	}
	return cs
}

// valueCompleter returns a set of values that complete token for
// keyword. If keyword is not a valid tokKeyword, valueCompleter
// returns nil.
func valueCompleter(keyword, token string) []string {
	if !keywordValid(tokKeyword(keyword)) {
		return nil
	}
	switch tokKeyword(keyword) {
	case keyName:
		return nameCompleter(token)
	case keyHelp:
		return keywordCompleter(keyword)
	}
	return nil
}

// nameCompleter returns a set of values that complete token for name.
// Because nameCompleter blocks, only completes tokens larger than
// three characters.
// TODO: investigave blocking -- can the user break out of it?
func nameCompleter(token string) []string {
	if len(token) < 4 {
		return nil
	}
	// PERF: do we need to limit this call?
	c := &StoreDefsCmd{
		Query:    string(token),
		CommitID: activeContext.repo.CommitID,
	}
	defs, err := c.Get()
	if err != nil {
		// TODO: log this error.
		return nil
	}
	completions := make([]string, 0, len(defs))
	for _, d := range defs {
		completions = append(completions, d.Name)
	}
	return completions
}

// cleanOutput returns o with only one trailing newline.
func cleanOutput(o string) string {
	return strings.TrimSuffix(o, "\n") + "\n"
}

// inputValues holds the fully-parsed input.
type inputValues struct {
	m map[tokKeyword][]tokValue
}

func newInputValues() *inputValues {
	return &inputValues{m: make(map[tokKeyword][]tokValue)}
}

// TODO: turn inputValues into a map. There is no reason to
// implement a map with switch statements.

func (i inputValues) isEmpty() bool {
	return len(i.m) == 0
}

// get returns the tokValue slice that tokKeyword 'k' maps to. get
// panics if k is not a valid keyword.
func (i inputValues) get(k tokKeyword) []tokValue {
	if !keywordValid(k) {
		panic("get: " + tokKeywordError{k}.Error())
	}
	return i.m[k]
}

// append appends vs to the tokValue slice that tokKeyword 'k' maps to.
func (i *inputValues) append(k tokKeyword, vs ...tokValue) error {
	if !keywordValid(k) {
		return tokKeywordError{k}
	}
	i.m[k] = append(i.m[k], vs...)
	return nil
}

// init initializes the tokValue slice that k maps to. It is used to
// indicate that the keyword has been seen.
func (i *inputValues) init(k tokKeyword) error {
	if !keywordValid(k) {
		return tokKeywordError{k}
	}
	i.m[k] = []tokValue{}
	return nil
}

// validate returns nil if 'i' is valid for the tokKeyword 'k'.
// Otherwise, validate returns an error.
func (i inputValues) validate(k tokKeyword) error {
	info := keywordInfoMap[k]
	if info.validVals == nil {
		return nil
	}
	if !keywordValid(k) {
		return tokKeywordError{k}
	}
	for _, val := range i.get(k) {
		var valid bool
		for _, v := range info.validVals {
			if val == v {
				valid = true
				break
			}
		}
		if !valid {
			return fmt.Errorf("%s is not a valid value for :%s. The following are valid values for :%s: %v", val, k, k, info.validVals)
		}
	}
	return nil
}

// validateAll returns an error if i is invalid.
func (i inputValues) finalize() error {
	for _, keyword := range allKeywords {
		if err := i.validate(keyword); err != nil {
			return err
		}
	}
	return nil
}

// setDefaults sets default values for i's empty tokValue slices.
func (i *inputValues) setDefaults() *inputValues {
	for keyword, info := range keywordInfoMap {
		if len(i.get(keyword)) == 0 {
			i.append(keyword, info.defaultVals...)
		}
	}
	return i
}

// A token is a structure that a lexer can emit.
type token interface {
	isToken()
}

// tokError represents a lexing error.
type tokError struct {
	msg   string
	start int
	pos   int
}

func (e tokError) isToken() {}

func (e tokError) Error() string { return fmt.Sprintf("%d:%d: %s", e.start, e.pos-e.start, e.msg) }

// tokEOF is emitted by the lexer when it is out of input.
type tokEOF struct{}

func (e tokEOF) isToken() {}

// A tokKeyword is a keyword for the 'src i' lanugage. Keywords always
// start with ":" in the language, but tokKeywords do not contain the
// ":". Do not cast strings to tokKeywords. Always use 'toTokKeyword'.
type tokKeyword string

var (
	keyName   tokKeyword = "name"
	keyIn     tokKeyword = "in"
	keySelect tokKeyword = "select"
	keyFormat tokKeyword = "format"
	keyKind   tokKeyword = "kind"
	keyFile   tokKeyword = "file"
	keyLimit  tokKeyword = "limit"
	keyHelp   tokKeyword = "help"
)

// keywordInfo holds a keyword's meta information.
type keywordInfo struct {
	// If validVals is nil, then the keyword is not restricted to
	// any specific values. Otherwise, the keyword is only valid
	// if its values match one of validVals.
	validVals []tokValue
	// defaultVals are the default values for a keyword. They are
	// only set if the user does not specify values for the
	// keyword. If defaultVals is nil, then the keyword has no
	// default vals.
	defaultVals []tokValue
	// typeConstraint constrains the keyword to a type.
	// typeConstraint can be "int" or the empty string. If
	// typeConstraint is non-empty, then validVals must be nil.
	typeConstraint string
	// argName is the semantic name for the keyword's argument
	// list. The help text displays it as ":<keyword> <argName>".
	argName string
	// description is a description of the keyword's semantics.
	description string
}

// keywordInfoMap is a map from tokKeyword to keywordInfo for every
// keyword.
// Invariant: keywordInfoMap is not modified.
var keywordInfoMap = map[tokKeyword]keywordInfo{
	keyName: keywordInfo{
		argName: "name",
		description: `Narrow to all objects matching 'name'.
':name' is special because it is implicit if no keyword is used, and its input cannot be a list. For example, the input 'some word' is equivalent to ':name some word'.`,
	},
	keyIn: keywordInfo{
		argName:     "repo",
		description: "Change the repository to 'repo'.",
	},
	keySelect: keywordInfo{
		validVals:   []tokValue{"defs", "refs", "docs"},
		defaultVals: []tokValue{"defs"},
		argName:     "objects",
		description: "Narrow search to 'objects'.",
	},
	keyKind: keywordInfo{
		argName:     "kinds",
		description: `Narrow search to 'kinds', which are language-specific "kinds" of objects. Possible values include "type", "func", "var", etc.`,
	},
	keyFile: keywordInfo{
		argName:     "prefixes",
		description: `Narrow search to files that begin with any values in 'prefixes'`,
	},
	keyFormat: keywordInfo{
		validVals:   []tokValue{"decl", "methods", "body", "full"},
		defaultVals: []tokValue{"decl", "body"},
		argName:     "formats",
		description: "Show defs in the formats specificed by 'formats'.",
	},
	keyLimit: keywordInfo{
		typeConstraint: "int",
		argName:        "number",
		description:    "Only display 'number' results.",
	},
	keyHelp: keywordInfo{
		argName:     "topics",
		description: "Show the help for 'topics'. If 'topics' is empty, show general help.",
	},
}

// allKeywords is a sorted list of all available keywords.
var allKeywords = keywordInfoMapToList(keywordInfoMap)

// keywordInfoMapToList returns a sorted list of all of the keywords
// in m.
func keywordInfoMapToList(m map[tokKeyword]keywordInfo) []tokKeyword {
	// ks is a string map for easier sorting.
	ks := make([]string, 0, len(m))
	for key := range m {
		ks = append(ks, string(key))
	}
	sort.Strings(ks)
	keywords := make([]tokKeyword, 0, len(ks))
	for _, k := range ks {
		keywords = append(keywords, tokKeyword(k))
	}
	return keywords
}

func (k tokKeyword) isToken() {}

// tokKeywordError represents a validation error for a keyword.
type tokKeywordError struct {
	k tokKeyword
}

func (e tokKeywordError) Error() string {
	if e.k == "" {
		return "invalid keyword: keyword is empty"
	}
	return fmt.Sprintf("unknown keyword: %s", e.k)
}

func keywordValid(k tokKeyword) bool {
	_, ok := keywordInfoMap[k]
	return ok
}

// toTokKeyword returns a keyword for 's'. It does not check 's's
// validity.
func toTokKeyword(s string) tokKeyword {
	return tokKeyword(strings.ToLower(s))
}

// A tokValue is any non-keyword value.
type tokValue string

func (v tokValue) isToken() {}

// formatValues returns vs as comma-delimited string.
func formatValues(vs []tokValue) string {
	buf := &bytes.Buffer{}
	for i, v := range vs {
		buf.WriteString(string(v))
		if i != len(vs)-1 {
			buf.WriteString(", ")
		}
	}
	return buf.String()
}

// These character runs are used by the lexer to identify terms.
const (
	horizontalWhitespace = " \t"
	alpha                = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	num                  = "0123456789"
	symbol               = "-()*_\\"
	wordChar             = alpha + num + symbol
)

// lexer holds the state of the lexer.
type lexer struct {
	input  string     // String being scanned.
	start  int        // Start position of token.
	pos    int        // Current position of input.
	width  int        // Width of last rune read.
	tokens chan token // Channel of scanned tokens.
}

const eof = -1

// TODO(samer): This leaks channels/lexers if the parsing step errors out.
func (l *lexer) run() {
	l.input = strings.TrimSpace(l.input)
	for state := lexStart; state != nil; {
		state = state(l)
	}
	l.emitEOF()
}

// next returns the next rune and steps 'width' forward.
func (l *lexer) next() rune {
	if l.pos >= len(l.input) {
		l.width = 0
		return eof
	}
	r, s := utf8.DecodeRuneInString(l.input[l.pos:])
	if r == utf8.RuneError && s == 1 {
		log.Fatal("input error")
	}
	l.width = s
	l.pos += l.width
	return r
}

// backup can only be called once after each call to 'next'.
func (l *lexer) backup() {
	l.pos -= l.width
}

// accept returns true and moves forward if the next rune is in
// 'valid'.
func (l *lexer) accept(valid string) bool {
	if strings.IndexRune(valid, l.next()) != -1 {
		return true
	}
	l.backup()
	return false
}

// acceptRun moves forward for all runes that match 'valid'.
func (l *lexer) acceptRun(valid string) {
	for strings.IndexRune(valid, l.next()) != -1 {
	}
	l.backup()
}

// acceptRunAllBut moves forward for all runes that do not match
// 'invalid'.
func (l *lexer) acceptRunAllBut(invalid string) {
	n := l.next()
	for n != eof && strings.IndexRune(invalid, n) == -1 {
		n = l.next()
	}
	l.backup()
}

// peek returns the next rune without moving forward.
func (l *lexer) peek() rune {
	r := l.next()
	l.backup()
	return r
}

// ignore ignores the text that the lexer has eaten.
func (l *lexer) ignore() {
	l.start = l.pos
}

// emitKeyword emits the text the lexer has eaten as a tokKeyword.
func (l *lexer) emitKeyword() {
	v := l.input[l.start:l.pos]
	l.start = l.pos
	l.tokens <- tokKeyword(v)
}

// emitValue emits the text the lexer has eaten as a tokValue. If
// 'quoted' is true, then all backslash escape sequences are replaced
// with the literal of the escaped char. If 'quoted' is false, then
// the value's whitespace is trimmed on both sides.
func (l *lexer) emitValue(quoted bool) {
	v := l.input[l.start:l.pos]
	l.start = l.pos
	if quoted {
		for i := 0; i < len(v); i++ {
			if v[i] == '\\' {
				if i == len(v)-1 {
					log.Println("emitValue: '\\' cannot be last character in quoted string. Please file a bug report.")
				} else {
					v = v[:i] + v[i+1:]
				}
				i++ // Don't process the escaped char.
			}
		}
	} else {
		v = strings.TrimSpace(v)
	}
	l.tokens <- tokValue(v)
}

// emitErrorf emits a formatted tokError.
func (l *lexer) emitErrorf(format string, a ...interface{}) {
	l.tokens <- tokError{msg: fmt.Sprintf(format, a), start: l.start, pos: l.pos}
}

// emitError emits a tokError.
func (l *lexer) emitError(a ...interface{}) {
	l.tokens <- tokError{msg: fmt.Sprint(a), start: l.start, pos: l.pos}
}

// emitEOF emits a tokEOF.
func (l *lexer) emitEOF() {
	l.tokens <- tokEOF{}
}

// A stateFn is a function that represents one of the lexer's states.
type stateFn func(l *lexer) stateFn

// lexStart is the entrypoint for the lexer.
func lexStart(l *lexer) stateFn {
	// if GlobalOpt.Verbose {
	// 	log.Printf("lexStart: on %s", string(l.peek()))
	// }
	l.acceptRun(horizontalWhitespace)
	l.ignore()
	if l.peek() == eof {
		return nil
	}
	if l.accept(":") {
		l.ignore()
		return lexKeyword
	}
	if l.accept("\"") {
		l.ignore()
		return lexDoubleQuote
	}
	if l.accept("'") {
		l.ignore()
		return lexSingleQuote
	}
	if l.accept(wordChar) {
		return lexValue
	}
	l.emitErrorf("unexpected char: '%s'", l.next())
	return nil
}

func lexKeyword(l *lexer) stateFn {
	// if GlobalOpt.Verbose {
	// 	log.Printf("lexKeyword: on %s", string(l.peek()))
	// }
	l.acceptRun(alpha)
	l.emitKeyword()
	return lexStart
}

func lexDoubleQuote(l *lexer) stateFn {
	// if GlobalOpt.Verbose {
	// 	log.Printf("lexDoubleQuote: on %s", string(l.peek()))
	// }
	return lexQuote(l, lexDoubleQuote, '"')
}

func lexSingleQuote(l *lexer) stateFn {
	// if GlobalOpt.Verbose {
	// 	log.Printf("lexSingleQuote: on %s", string(l.peek()))
	// }
	return lexQuote(l, lexSingleQuote, '\'')
}

func lexQuote(l *lexer, fromFn stateFn, quote rune) stateFn {
	// if GlobalOpt.Verbose {
	// 	log.Printf("lexQuote: on %s", string(l.peek()))
	// }
	l.acceptRunAllBut(string(quote) + "\\")
	n := l.next()
	switch n {
	case eof:
		// TODO(samer): Continue input.
		l.emitError("unexpected eof")
		return nil
	case '\\':
		l.next() // eat next char
	case quote:
		l.backup()
		l.emitValue(true)
		l.next()
		l.ignore() // ignore quote
		return lexStart
	}
	l.emitErrorf("unexpected char: '%s'", string(n))
	return nil
}

func lexValue(l *lexer) stateFn {
	// if GlobalOpt.Verbose {
	// 	log.Printf("lexValue: on %s", string(l.peek()))
	// }
	l.acceptRunAllBut(":,")
	if l.accept(":") {
		l.backup()
		l.emitValue(false)
		return lexStart
	}
	if l.accept(",") {
		l.backup()
		l.emitValue(false)
		l.next()
		l.ignore() // ignore ','
		return lexValue
	}
	if l.peek() == eof {
		l.emitValue(false)
		return nil
	}
	panic("unreachable")
	return nil
}

type parseError struct{ error }

func (e *parseError) Error() string { return fmt.Sprintf("parse error: %s", e.error) }

// parse parses the user input and organizes it into inputValues.
// Invariant: if error is not nil, it is of type parseError.
func parse(input string) (*inputValues, error) {
	if GlobalOpt.Verbose {
		log.Printf("parsing: input %s\n", input)
	}
	l := &lexer{
		input:  input,
		tokens: make(chan token),
	}
	go l.run()
	// Create the inputValues from the input tokens.
	i := newInputValues()
	type parseState int

	// invariant:
	//  - on is empty before the loop starts, and is set on the
	//  first successful iteration of theloop.
	var on tokKeyword
loop:
	for t := range l.tokens {
		switch t := t.(type) {
		case tokEOF:
			if GlobalOpt.Verbose {
				log.Printf("parsing: got tokEOF\n")
			}
			break loop
		case tokError:
			if GlobalOpt.Verbose {
				log.Printf("parsing: got tokError %s\n", t)
			}
			return nil, &parseError{t}
		case tokKeyword:
			if GlobalOpt.Verbose {
				log.Printf("parsing: got tokKeyword %s\n", t)
			}
			// if we see a tokKeyword, check that the input
			// for the previously seen tokKeyword (stored as
			// 'on') is valid.
			if on != "" {
				if err := i.validate(on); err != nil {
					return nil, &parseError{err}
				}
			}
			if !keywordValid(t) {
				return nil, &parseError{tokKeywordError{t}}
			}
			if len(i.get(t)) > 0 {
				return nil, &parseError{fmt.Errorf("error: keyword :%s can only appear once.", t)}
			}
			// Set 'on' to the new tokKeyword.
			on = t
			i.init(on)
		case tokValue:
			if GlobalOpt.Verbose {
				log.Printf("parsing: got tokValue %s\n", t)
			}
			// If we haven't seen a tokKeyword ('on' is
			// empty), then we're implicitly on keyName.
			if on == "" {
				on = keyName
			}
			i.append(on, t)
		default:
			panic("unknown concrete type for token: " + reflect.TypeOf(t).Name())
		}
	}
	if err := i.finalize(); err != nil {
		return nil, &parseError{err}
	}

	if GlobalOpt.Verbose {
		log.Printf("parsed: %+v\n", i)
	}
	return i, nil
}

func inputToFormat(i *inputValues) format {
	var f format
	for _, s := range i.get(keySelect) {
		switch s {
		case "defs":
			f.showDefs = true
		case "refs":
			f.showRefs = true
		case "docs":
			f.showDocs = true
		}
	}
	for _, fmt := range i.get(keyFormat) {
		switch fmt {
		case "decl":
			f.showDefDecl = true
		case "methods":
			f.showDefMethods = true
		case "body":
			f.showDefBody = true
		case "full":
			f.showDefFull = true
		}
	}
	// TODO: make limit parsing more robust.
	if len(i.get(keyLimit)) == 1 {
		l, err := strconv.Atoi(string(i.get(keyLimit)[0]))
		if err != nil {
			log.Printf("Could not convert limit %s to an int, skipping.\n", i.get(keyLimit)[0])
		}
		f.limit = l
	}
	return f
}

type format struct {
	showDefs    bool
	showRefs    bool
	showDocs    bool
	showDefDecl bool
	showDefBody bool
	limit       int
	// The following are unimplemented:
	showDefMethods bool
	showDefFull    bool
}

// briefHelpText prints out the commands in brief, suitable for
// displaying after an error.
func briefHelpText() string {
	buf := &bytes.Buffer{}
	buf.WriteString("Available commands: -- \":help all\" for detailed help, \":help usage\" for usage\n")
	for i, k := range allKeywords {
		fmt.Fprintf(buf, "  :%s <%s>", k, keywordInfoMap[k].argName)
		if i != len(allKeywords)-1 {
			buf.WriteString("\n")
		}
	}
	return buf.String()
}

// helpText returns the concatenated help text for each topic, where
// topic is "all", "usage" or a keyword name (such as "format").
// TODO: explain "all" and "usage".
func helpText(topics []tokValue) (string, error) {
	// If no topics are specified, show the help for all topics.
	if len(topics) == 0 {
		return briefHelpText(), nil
	}
	// If the first value is "all", show help on every topic.
	if topics[0] == "all" {
		topics = make([]tokValue, 0, len(allKeywords)+1)
		topics = append(topics, tokValue("usage"))
		for _, k := range allKeywords {
			topics = append(topics, tokValue(k))
		}
	}
	buf := &bytes.Buffer{}
	indent := "  "
	for _, t := range topics {
		// "usage" is a special value. It prints out the
		// tutorial for help.
		if t == "usage" {
			fmt.Fprint(buf, `These are some example queries:
src> Hello
;; All definitions (funcs, vars, etc) that begin with "Hello".
src> :name Hello
;; Same as above. The ":name" keyword is assumed if left out.
src> Hello :select defs, refs
;; All defintions that begin with "Hello" and their references.
;; If ":select" is left out, "defs" is included by default.
src> Hello :kind func
;; All functions that begin with "Hello". ":kind" is language-defined.
src> Hello :format decl
;; All definition declarations -- ignore defintion bodies.
`)
			continue
		}
		k := tokKeyword(t)
		info, ok := keywordInfoMap[k]
		if !ok {
			return "", fmt.Errorf("Topic %s does not exist", t)
		}
		if buf.Len() == 0 {
			fmt.Fprint(buf, "Keywords:\n")
		}
		fmt.Fprintf(buf, "-- :%s <%s>\n%s%s\n", k, info.argName,
			indent, strings.Replace(info.description, "\n", "\n"+indent, -1))
		if len(info.validVals) == 0 &&
			len(info.defaultVals) == 0 &&
			info.typeConstraint == "" {
			continue
		}
		buf.WriteString("\n")
		if len(info.validVals) != 0 {
			fmt.Fprintf(buf, "%sValid values for '%s': %s\n",
				indent, info.argName, formatValues(info.validVals))
		}
		if len(info.defaultVals) != 0 {
			fmt.Fprintf(buf, "%sDefault values for '%s': %s\n",
				indent, info.argName, formatValues(info.defaultVals))
		}
		if info.typeConstraint == "" {
			fmt.Fprintf(buf, "%s'%s' must be of type %s\n",
				indent, info.argName, info.typeConstraint)
		}
	}
	return buf.String(), nil
}

// eval evaluates input and returns the results as output. If output
// is non-empty, it should be displayed even if err is non-nil.
func eval(input string) (output string, err error) {
	i, err := parse(input)
	if err != nil {
		return briefHelpText(), err
	}
	// There is some sense of order: if i is empty or ":help" is
	// set, do not evaluate the rest of 'i'.
	switch {
	case i.isEmpty():
		return briefHelpText(), nil
	case i.get(keyHelp) != nil:
		return helpText(i.get(keyHelp))
	}
	i.setDefaults()

	f := inputToFormat(i)
	var out []string
	// TODO: only deal with one name!
	for _, input := range i.get(keyName) {
		c := &StoreDefsCmd{
			Query:    string(input),
			CommitID: activeContext.repo.CommitID,
			Limit:    f.limit,
		}
		// TODO: make the following filters work with more
		// than one value.
		if len(i.get(keyKind)) != 0 {
			c.Filter = byDefKind{string(i.get(keyKind)[0])}
		}
		if len(i.get(keyFile)) != 0 {
			c.File = string(i.get(keyFile)[0])
		}
		defs, err := c.Get()
		if err != nil {
			return "", err
		}
		if f.showRefs {
			outDefRefs := make([]defRefs, 0, len(defs))
			for _, d := range defs {
				c := &StoreRefsCmd{
					DefRepo:     d.Repo,
					DefUnitType: d.UnitType,
					DefUnit:     d.Unit,
					DefPath:     d.Path,
				}
				refs, err := c.Get()
				if err != nil {
					return "", err
				}
				outDefRefs = append(outDefRefs, defRefs{d, refs})
			}
			out = append(out, formatObject(outDefRefs, f))
			continue
		}
		out = append(out, formatObject(defs, f))
	}
	return strings.Join(out, "\n"), nil
}

// TODO: move to store package.
type byDefKind struct {
	kind string
}

func (h byDefKind) SelectDef(def *graph.Def) bool {
	return def.Kind == "" || def.Kind == h.kind
}

type defRefs struct {
	def  *graph.Def
	refs []*graph.Ref
}

func getFileSegment(file string, start, end uint32, header bool) string {
	f, err := ioutil.ReadFile(file)
	if err != nil {
		return ""
	}
	if header {
		startLine := bytes.Count(f[:start], []byte{'\n'}) + 1
		// Roll 'start' back and 'end' forward to the nearest
		// newline.
		for ; start-1 > 0 && f[start-1] != '\n'; start-- {
		}
		for ; end < uint32(len(f)) && f[end] != '\n'; end++ {
		}
		var out []string
		onLine := startLine
		for _, line := range bytes.Split(f[start:end], []byte{'\n'}) {
			var marker string
			if startLine == onLine {
				marker = ":"
			} else {
				marker = "-"
			}
			out = append(out, fmt.Sprintf("%s:%d%s%s",
				file, onLine, marker, string(line),
			))
			onLine++
		}
		return strings.Join(out, "\n")
	}
	return string(f[start:end])
}

func formatObject(objs interface{}, f format) string {
	switch o := objs.(type) {
	case *graph.Def:
		if o == nil {
			return "def is nil"
		}
		var output []string
		if f.showDefs {
			output = append(output, "---------- def ----------")
			if f.showDefDecl {
				b, err := json.Marshal(o)
				if err != nil {
					return fmt.Sprintf("error unmarshalling: %s", err)
				}
				c := &FmtCmd{
					UnitType:   o.UnitType,
					ObjectType: "def",
					Format:     "decl",
					Object:     string(b),
				}
				out, err := c.Get()
				if err != nil {
					return fmt.Sprintf("error formatting def: %s", err)
				}
				output = append(output, out)
			}
			if f.showDefBody {
				output = append(output, getFileSegment(o.File, o.DefStart, o.DefEnd, true))
			}
		}
		if f.showDocs {
			var data string
			for _, doc := range o.Docs {
				if doc.Format == "text/plain" {
					data = doc.Data
					break
				}
			}
			if data != "" {
				output = append(output, "---------- doc ----------", data)
			}
		}
		return strings.Join(output, "\n")
	case []*graph.Def:
		var out []string
		for _, d := range o {
			out = append(out, formatObject(d, f))
		}
		return strings.Join(out, "\n")
	case *graph.Ref:
		if o == nil {
			return "ref is nil"
		}
		if f.showRefs {
			return getFileSegment(o.File, o.Start, o.End, true)
		}
		return ""
	case []*graph.Ref:
		var out []string
		for _, r := range o {
			if !r.Def {
				out = append(out, formatObject(r, f))
			}
		}
		return strings.Join(out, "\n")
	case []defRefs:
		var out []string
		for _, d := range o {
			out = append(out, formatObject(d, f))
		}
		return strings.Join(out, "\n")
	case defRefs:
		var out []string
		out = append(out,
			formatObject(o.def, f),
			"---------- refs ----------",
			formatObject(o.refs, f),
		)
		return strings.Join(out, "\n")
	default:
		log.Printf("formatObject: no output for %#v\n", o)
		return ""
	}
}

// from google/cayley
func terminal(path string) (*liner.State, error) {
	term := liner.NewLiner()
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, os.Kill)
		<-c
		err := persist(term, historyFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to properly clean up terminal: %v\n", err)
			os.Exit(1)
		}
		os.Exit(0)
	}()
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return term, nil
		}
		return term, err
	}
	defer f.Close()
	_, err = term.ReadHistory(f)
	return term, err
}

// from google/cayley
func persist(term *liner.State, path string) error {
	f, err := os.OpenFile(path, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0666)
	if err != nil {
		return fmt.Errorf("could not open %q to append history: %v", path, err)
	}
	defer f.Close()
	_, err = term.WriteHistory(f)
	if err != nil {
		return fmt.Errorf("could not write history to %q: %v", path, err)
	}
	return term.Close()
}
