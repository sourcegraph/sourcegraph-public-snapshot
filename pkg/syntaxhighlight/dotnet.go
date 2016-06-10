package syntaxhighlight

import (
	"unicode/utf8"
)

func init() {

	var charMatcher = func(source []byte) []int {
		state := 0
		ucounter := 0
		buf := make([]byte, 0, 3)
		for i, c := range source {
			switch state {
			case 0:
				if c == '\'' {
					state = 1
				} else {
					return nil
				}
				break
			case 1:
				if c == '\\' {
					state = 2 // '\
				} else {
					buf = append(buf, c)
					r, _ := utf8.DecodeLastRune(buf)
					if r != utf8.RuneError {
						state = 3 // '...
					} else if len(buf) == 3 {
						return nil
					}
				}
				break
			case 2:
				// '\
				if c == 'u' || c == 'U' || c == 'x' {
					ucounter = 0
					state = 4 // '\u.... or '\x....
				} else {
					state = 3 // '\C
				}
				break
			case 3:
				if c == '\'' {
					return []int{0, i + 1}
				}
				return nil
			case 4:
				if c >= '0' && c <= '9' || c >= 'a' && c <= 'f' || c >= 'A' && c <= 'F' {
					ucounter++
					if ucounter == 4 {
						state = 3
					}
				} else if c == '\'' {
					return []int{0, i + 1}
				} else {
					return nil
				}
				break
			}
		}
		return nil
	}

	csIdent := `@?[_` + UnicodeClasses(`Lu`, `Ll`, `Lt`, `Lm`, `Nl`) + `][` + UnicodeClasses(`Lu`, `Ll`, `Lt`, `Lm`, `Nl`, `Nd`, `Pc`, `Cf`, `Mn`, `Mc`) + `]*`

	NewRegexpLexer(
		[]string{`.cs`},
		[]string{`text/x-csharp`},
		map[string][]RegexpRule{
			`root`: {
				MS.Action(`^([ \t]*(?:`+csIdent+`(?:\[\])?\s+)+?)(`+csIdent+`)(?:\s*)(\()`, ByGroups(UsingThis(), Name_Function, Punctuation)),
				MS.Token(`\[.*?\]`, Name_Attribute),
				MS.Consume("\\\n", false),
				MS.MatcherToken(SingleLineCommentMatcher("//"), Comment_Single),
				MS.MatcherToken(MultiLineCommentMatcher, Comment_Multiline),
				MS.Token(`[~!%^&*()+=|\[\]:;,.<>/?-]`, Punctuation),
				MS.Token(`[{}]`, Punctuation),
				MS.Token(`@"(""|[^"])*"`, String),
				MS.Token(`"(\\\\|\\"|[^"\n])*["\n]`, String),
				MS.MatcherToken(charMatcher, String_Char),
				MS.MatcherToken(NumberMatcher("fldFLD"), Number),
				MS.MatcherToken(HexNumberMatcher, Number),
				MS.Token(`#[ \t]*(if|endif|else|elif|define|undef|line|error|warning|region|endregion|pragma)\b.*?\n`, Comment_Preproc),
				MS.MatcherToken(Words(`extern`, `alias`), Keyword),
				MS.MatcherToken(Words(`abstract`, `async`, `await`, `base`, `break`, `case`, `catch`, `checked`, `const`, `continue`, `default`, `delegate`, `do`, `else`, `enum`, `event`, `explicit`, `extern`, `false`, `finally`, `fixed`, `foreach`, `goto`, `if`, `implicit`, `interface`, `internal`, `is`, `lock`, `new`, `null`, `operator`, `out`, `override`, `params`, `private`, `protected`, `public`, `readonly`, `ref`, `return`, `sealed`, `sizeof`, `stackalloc`, `static`, `switch`, `this`, `throw`, `true`, `try`, `typeof`, `unchecked`, `unsafe`, `virtual`, `void`, `while`, `get`, `set`, `partial`, `yield`, `add`, `remove`, `value`, `alias`, `ascending`, `descending`, `from`, `group`, `into`, `orderby`, `select`, `where`, `join`, `equals`), Keyword),
				MS.MatcherToken(Words(`as`, `for`, `in`), Keyword),
				MS.Action(`(global)(::)`, ByGroups(Keyword, Punctuation)),
				MS.Token(`(bool|byte|char|decimal|double|dynamic|float|int|long|object|sbyte|short|string|uint|ulong|ushort|var)\b\??`, Keyword_Type),
				MS.MatcherToken(Words(`class`, `struct`), Keyword, `class`),
				MS.MatcherToken(Words(`namespace`, `using`), Keyword, `namespace`),
				MS.Token(csIdent, Name),
			},
			`class`: {
				MS.Token(csIdent, Name_Class, `#pop`),
				Default(`#pop`),
			},
			`namespace`: {
				MS.Consume(``, true),
				MS.Lookahead(`\(`, `#pop`),
				MS.Token(csIdent, Name_Namespace),
				MS.MatcherToken(Words(`;`), Punctuation, `#pop`),
				MS.MatcherToken(Words(`.`), Punctuation),
				Default(`#pop`),
			},
		})
}
