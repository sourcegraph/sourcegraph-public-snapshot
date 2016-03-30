package syntaxhighlight

import (
	"bytes"
	"strings"
)

func init() {

	// Matches identifiers
	// According to the PHP manual identifier should match against [a-zA-Z_\x7f-\xff][a-zA-Z0-9_\x7f-\xff]*
	// but we'll add backslash to support \foo or foo\bar
	var identSequence = func(source []byte) int {
		state := 0
		for i, c := range source {
			switch state {
			case 0:
				if c >= 'a' && c <= 'z' ||
					c >= 'A' && c <= 'Z' ||
					c == '_' ||
					c >= 0x7f && c <= 0xff ||
					c == '\\' {
					state = 1
				} else {
					return 0
				}
				break
			case 1:
				if c >= 'a' && c <= 'z' ||
					c >= 'A' && c <= 'Z' ||
					c == '_' ||
					c >= 0x7f && c <= 0xff ||
					c >= '0' && c <= '9' ||
					c == '\\' {
					// continue
				} else {
					return i
				}
				break
			}
		}
		return len(source)
	}

	var identMatcher = func(source []byte) []int {
		l := identSequence(source)
		if l > 0 {
			return []int{0, l}
		}
		return nil
	}

	var varMatcher = func(source []byte) []int {
		if source[0] != '$' {
			return nil
		}
		sLen := len(source)
		if sLen < 2 {
			return nil
		}

		var offset int
		curly := false

		if source[1] == '{' {
			// ${....}
			if sLen < 3 {
				return nil
			}
			offset = 2
			curly = true
		} else {
			offset = 1
		}

		l := identSequence(source[offset:])
		if l > 0 {
			l = l + offset
			if curly {
				if l == sLen {
					return nil // ${.... w/o "}"
				}
				if source[l] != '}' {
					return nil // ${.... w/o "}"
				}
				l += 1
			}
			return []int{0, l}
		}
		return nil
	}

	// Matches octal numbers
	var octNumberMatcher = func(source []byte) []int {
		state := 0
		for i, c := range source {
			switch state {
			case 0:
				if c == '0' {
					state = 1
				} else {
					return nil
				}
				break
			case 1:
				if c >= '0' && c <= '7' {
					// continue
				} else {
					return []int{0, i + 1}
				}
				break
			}
		}
		return []int{0, len(source)}
	}

	// Matches binary numbers
	var binNumberMatcher = func(source []byte) []int {
		state := 0
		for i, c := range source {
			switch state {
			case 0:
				if c == '0' {
					state = 1
				} else {
					return nil
				}
				break
			case 1:
				if c == 'b' {
					state = 2
				} else {
					return nil
				}
				break
			case 2:
				if c == '0' || c == '1' {
					// continue
				} else {
					return []int{0, i + 1}
				}
				break
			}
		}
		return []int{0, len(source)}
	}

	var hereDoc = []byte{'<', '<', '<'}

	// Matches heredoc and nowdoc literals
	var heredocMatcher = func(source []byte) []int {
		l := len(source)
		// <<<A\nA;
		if l < 6 {
			return nil
		}
		// starts with <<<
		if !bytes.HasPrefix(source, hereDoc) {
			return nil
		}
		eol := bytes.IndexAny(source, "\r\n")
		if eol < 0 {
			return nil
		}
		// "'?...'?" or '"?..."?'
		ident := bytes.TrimSpace(source[3:eol])
		il := len(ident)
		if ident[0] == '\'' || ident[0] == '"' {
			if il < 3 {
				return nil
			}
			ident = ident[1 : il-1]
			il = il - 2
		}
		// ...;
		tail := make([]byte, il)
		copy(tail, ident)
		tail = append(tail, ';')
		end := bytes.Index(source, tail)
		if end < 0 {
			return nil
		}
		return []int{0, end + len(tail)}
	}

	// Matches operations
	var opMatcher = func(source []byte) []int {

		const ops = "~!%^&*+=|:.<>/@-"

		for i, c := range source {
			if strings.IndexByte(ops, c) == -1 {
				if i == 0 {
					return nil
				}
				return []int{0, i}
			}
		}
		return []int{0, len(source)}
	}

	// Matches punctuation
	var punctMatcher = func(source []byte) []int {

		const ops = "[]{}();,"

		for i, c := range source {
			if strings.IndexByte(ops, c) == -1 {
				if i == 0 {
					return nil
				}
				return []int{0, i}
			}
		}
		return []int{0, len(source)}
	}

	var initCallback = func(self *RegexpLexer, source []byte) {
		// if we found no "<?", assuming that we are working with fragment
		var isFragment = bytes.Index(source, []byte{'<', '?'}) == -1
		if isFragment {
			self.statestack = append(self.statestack, `php`)
		}
	}

	NewRegexpLexerWithCallback(
		[]string{`.php`, `.php3`, `.php4`, `.php5`, `.inc`},
		[]string{`text/x-php`},
		map[string][]RegexpRule{
			`root`: {
				MSI.MatcherToken(WordWithBoundary(`<?php`, false), Comment_Preproc, `php`),
				MSI.MatcherToken(WordWithBoundary(`<?`, false), Comment_Preproc, `php`),
			},
			`php`: {
				MSI.MatcherToken(WordWithBoundary(`?>`, false), Comment_Preproc, `#pop`),
				MS.MatcherToken(heredocMatcher, String),
				MSI.MatcherToken(SingleLineCommentMatcher("#"), Comment_Single),
				MSI.MatcherToken(SingleLineCommentMatcher("//"), Comment_Single),
				// put the empty comment here, it is otherwise seen as
				// the start of a docstring
				MSI.MatcherToken(MultiLineCommentMatcher, Comment_Multiline),
				// (r'/\*\*.*?\*/', String.Doc), TODO?
				MSI.MatcherToken(WordsWithBoundary(false, `->`, `::`), Operator, `attribute`),
				MSI.MatcherToken(opMatcher, Operator),
				MSI.MatcherToken(WordWithBoundary(`?`, false), Operator), // don't add to the charclass above!
				MSI.MatcherToken(punctMatcher, Punctuation),
				MSI.MatcherToken(Word(`class`), Keyword, `classname`),
				MSI.MatcherToken(Word(`constant`), Keyword, `constant`),
				MSI.MatcherToken(Words(`and`, `E_PARSE`, `old_function`, `E_ERROR`, `or`, `as`, `E_WARNING`, `parent`,
					`eval`, `PHP_OS`, `break`, `exit`, `case`, `extends`, `PHP_VERSION`, `cfunction`, `function`,
					`FALSE`, `print`, `for`, `require`, `continue`,
					`declare`, `return`, `default`, `static`, `do`, `switch`, `die`, `stdClass`,
					`echo`, `else`, `TRUE`, `var`, `empty`, `if`, `xor`, `enddeclare`, `include`,
					`virtual`, `endfor`, `while`, `global`, `__FILE__`,
					`endif`, `list`, `__LINE__`, `endswitch`, `new`, `__sleep`, `endwhile`, `not`,
					`array`, `__wakeup`, `E_ALL`, `NULL`, `final`, `php_user_filter`, `interface`,
					`implements`, `public`, `private`, `protected`, `abstract`, `clone`, `try`,
					`catch`, `throw`, `this`, `use`, `namespace`, `trait`, `yield`), Keyword),
				MSI.MatcherToken(Words(`foreach`, `require_once`, `include_once`, `elseif`, `endforeach`, `finally`), Keyword),
				MSI.MatcherToken(Words(`true`, `false`, `null`), Keyword_Constant),
				MSI.MatcherToken(varMatcher, Name_Variable),
				MSI.MatcherToken(identMatcher, Name_Other),
				MSI.MatcherToken(NumberMatcher(""), Number),
				MSI.MatcherToken(HexNumberMatcher, Number_Hex),
				MSI.MatcherToken(octNumberMatcher, Number_Oct),
				MSI.MatcherToken(binNumberMatcher, Number_Bin),
				MSI.MatcherToken(StringMatcher('\''), String_Single),
				MSI.MatcherToken(StringMatcher('`'), String_Backtick),
				MSI.MatcherToken(StringMatcher('"'), String_Double),
			},
			`classname`: {
				MSI.MatcherToken(identMatcher, Name_Class, `#pop`),
			},
			`constant`: {
				MSI.MatcherToken(identMatcher, Name_Constant, `#pop`),
			},
			`attribute`: {
				MSI.MatcherToken(identMatcher, Name_Attribute, `#pop`),
			},
		},
		initCallback)
}
