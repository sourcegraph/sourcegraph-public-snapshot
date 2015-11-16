package syntaxhighlight

import (
	"bytes"
)

func init() {

	ident_char := `[\\\w]|[^\x00-\x7f]`
	ident_begin := `(?:[\\_a-z]|[^\x00-\x7f])`
	ident_end := `(?:` + ident_char + `)*`
	ident_inner := ident_begin + ident_end

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
		tail := make([]byte, 0, il+1)
		copy(tail, ident)
		tail = append(tail, ';')
		end := bytes.Index(source, tail)
		if end < 0 {
			return nil
		}
		return []int{0, end + len(tail)}
	}

	NewRegexpLexer(
		[]string{`.php`, `.php3`, `.php4`, `.php5`, `.inc`},
		[]string{`text/x-php`},
		map[string][]RegexpRule{
			`root`: {
				MSI.Token(`<\?(php)?`, Comment_Preproc, `php`),
			},
			`php`: {
				MSI.Token(`\?>`, Comment_Preproc, `#pop`),
				MS.MatcherToken(heredocMatcher, String),
				MSI.MatcherToken(SingleLineCommentMatcher("#"), Comment_Single),
				MSI.MatcherToken(SingleLineCommentMatcher("//"), Comment_Single),
				// put the empty comment here, it is otherwise seen as
				// the start of a docstring
				MSI.MatcherToken(MultiLineCommentMatcher, Comment_Multiline),
				// (r'/\*\*.*?\*/', String.Doc), TODO?
				MSI.Action(`(->|::)(?:\s*)(`+ident_inner+`)`, ByGroups(Operator, Name_Attribute)),
				MSI.Token(`[~!%^&*+=|:.<>/@-]+`, Operator),
				MSI.Token(`\?`, Operator), // don't add to the charclass above!
				MSI.Token(`[\[\]{}();,]+`, Punctuation),
				MSI.MatcherToken(Words(`class`), Keyword, `classname`),
				MSI.Action(`(const)(?:\s+)(`+ident_inner+`)`, ByGroups(Keyword, Name_Constant)),
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
				MSI.Token(`\$\{\$+`+ident_inner+`\}`, Name_Variable),
				MSI.Token(`\$+`+ident_inner, Name_Variable),
				MSI.Token(ident_inner, Name_Other),
				MS.MatcherToken(NumberMatcher(""), Number),
				MS.MatcherToken(HexNumberMatcher, Number_Hex),
				MS.MatcherToken(octNumberMatcher, Number_Oct),
				MS.MatcherToken(binNumberMatcher, Number_Bin),
				MS.MatcherToken(StringMatcher('\\'), String_Single),
				MS.MatcherToken(StringMatcher('`'), String_Backtick),
				MS.MatcherToken(StringMatcher('"'), String_Double),
			},
			`classname`: {
				MS.Token(ident_inner, Name_Class, `#pop`),
			},
		})
}
