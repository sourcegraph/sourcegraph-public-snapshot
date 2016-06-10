package syntaxhighlight

// Callback invoked after lexer's initialization
type InitCallback func(self *RegexpLexer, source []byte)

// Lexer that uses mostly RE to produce tokens
type RegexpLexer struct {
	// map of rules to produce tokens in form state => []rule
	rules map[string][]RegexpRule
	// source code to analyse
	source []byte
	// state stack
	statestack []string
	// pos holds current position in the source code. When applying RE rules
	// we should make sure we are working in anchored mode (PCRE) or prefixing RE with \A.
	// We should only accept matches that starts at the given position of source text
	pos int
	// tokens cache
	cache []Token
	// initialization callback
	initCallback InitCallback
}

// Registers new RE-based lexer
// - extensions - list of extensions supported by a given lexer (in form .java, .exe and so on)
// - mimeTypes - list of supported MIME types (for example, text/x-java)
// - rules - map(state => array) of lexer rules
func NewRegexpLexer(extensions []string, mimeTypes []string, rules map[string][]RegexpRule) {
	NewRegexpLexerWithCallback(extensions, mimeTypes, rules, nil)
}

// Registers new RE-based lexer with init callback
// - extensions - list of extensions supported by a given lexer (in form .java, .exe and so on)
// - mimeTypes - list of supported MIME types (for example, text/x-java)
// - rules - map(state => array) of lexer rules
// - initCallback - function to be called at init phase (if not nil)
func NewRegexpLexerWithCallback(extensions []string, mimeTypes []string, rules map[string][]RegexpRule, initCallback InitCallback) {
	defs := processTokenDefs(rules)
	register(extensions, mimeTypes, func() Lexer {
		var ret Lexer
		ret = &RegexpLexer{rules: defs, initCallback: initCallback}
		return ret
	})
}

// Initializes lexer
func (self *RegexpLexer) Init(source []byte) {
	self.source = source
	self.statestack = []string{`root`}
	self.pos = 0
	self.cache = make([]Token, 0, 10)
	if self.initCallback != nil {
		self.initCallback(self, source)
	}
}

// Produces tokens using RE
func (self *RegexpLexer) NextToken() *Token {

	for {

		if self.pos >= len(self.source) {
			return self.consumeCache()
		}

		rules := self.rules[self.statestack[len(self.statestack)-1]]
		slice := self.source[self.pos:]
		found := false
		produced := true
		for _, rule := range rules {
			matcher := rule.matcher(slice)
			if matcher == nil {
				continue
			}
			if rule.ttype != nil {
				tok := NewToken(slice[matcher[0]:matcher[1]], rule.ttype, self.pos+matcher[0])
				self.cache = append(self.cache, tok)
			} else if rule.action != nil {
				tokens := rule.action(self, slice, self.pos, matcher)
				if tokens != nil {
					self.cache = append(self.cache, tokens...)
				} else {
					produced = false
				}
			} else {
				produced = false
			}
			found = true
			self.pos += matcher[1]
			self.statestack = updateStack(self.statestack[0:], rule)
			rules = self.rules[self.statestack[len(self.statestack)-1]]
			break
		}
		if !found {
			self.pos++
			continue
		}
		if produced {
			return self.consumeCache()
		}
	}
}

// Emits one token from cache if any
func (self *RegexpLexer) consumeCache() *Token {
	l := len(self.cache)
	if l > 0 {
		ret := self.cache[0]
		self.cache = self.cache[1:]
		return &ret
	}
	return nil
}

// Pre-processes lexer rules, replaces meta definitions where needed (for example Include(STATE) => append(...))
func processTokenDefs(rules map[string][]RegexpRule) map[string][]RegexpRule {
	processed := map[string][]RegexpRule{}
	for state := range rules {
		processState(rules, processed, state)
	}
	return processed
}

// Processes lexer state, replaces meta definitions where needed
// unprocessed - rules to process
// processed - rules that were already processed
// state - state to process
func processState(unprocessed map[string][]RegexpRule, processed map[string][]RegexpRule, state string) []RegexpRule {
	ret := processed[state]
	if ret != nil {
		return ret
	}
	tokens := []RegexpRule{}

	for _, tdef := range unprocessed[state] {
		if tdef.include != "" {
			processed[state] = []RegexpRule{}
			tokens = append(tokens, processState(unprocessed, processed, tdef.include)...)
			processed[state] = tokens
			continue
		} else {
			tokens = append(tokens,
				RegexpRule{matcher: tdef.matcher, ttype: tdef.ttype, action: tdef.action, states: tdef.states})
			processed[state] = tokens
		}
	}
	return tokens
}

// updates stack state using given rule
// rule may define zero or more transitions to be applied, such as
// - #pop - pops stack
// - #push - pushed head stack item on the top of stack
// - <state> - put state to stack
func updateStack(stack []string, rule RegexpRule) []string {
	for _, state := range rule.states {
		if state == `#pop` {
			stack = stack[:len(stack)-1]
		} else if state == `#push` {
			stack = append(stack, stack[len(stack)-1])
		} else {
			stack = append(stack, state)
		}
	}
	return stack
}
