package syntaxhighlight

import (
	"bytes"
	"strings"
)

func init() {

	const stringPrefixes = "rRuU"

	var stringQuoteStyles = [][]byte{
		[]byte("'''"),
		[]byte(`"""`),
	}

	// consumes string prefix part if any, returns position of next character after string prefix
	var consumeStringPrefix = func(source []byte) int {
		r := rune(source[0])
		if !strings.ContainsRune(stringPrefixes, r) {
			return 0
		}
		l := len(source)
		if l == 1 || !strings.ContainsRune(stringPrefixes, rune(source[1])) {
			return 1
		}
		return 2
	}

	// matches strings [rRuU]("""|''')......("""|''')
	var tripleQMatcher = func(source []byte) []int {
		start := consumeStringPrefix(source)
		l := len(source)
		if start == l {
			return nil
		}
		for _, stringQuoteStyle := range stringQuoteStyles {
			if bytes.HasPrefix(source[start:], stringQuoteStyle) {
				prefixLen := len(stringQuoteStyle)
				end := bytes.Index(source[start+prefixLen:], stringQuoteStyle)
				if end < 0 {
					return nil
				}
				return []int{0, start + end + prefixLen*2}
			}
		}
		return nil
	}

	// matches backticks `...`
	var backticksMatcher = func(source []byte) []int {
		l := len(source)
		if l < 2 {
			// must have enough bytes for ``
			return nil
		}
		if source[0] == '`' {
			end := bytes.IndexByte(source[1:], '`')
			if end < 0 {
				return nil
			}
			return []int{0, end + 2}
		}
		return nil
	}

	// Matches binary numbers in form 0[bB][01]+
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
				if c == 'b' || c == 'B' {
					state = 2
				} else {
					return nil
				}
				break
			case 2:
				if c == '0' || c == '1' {
					// continue
				} else {
					return []int{0, i}
				}
				break
			}
		}
		return []int{0, len(source)}
	}

	NewRegexpLexer(
		[]string{`.py`, `.pyw`, `.sc`, `.SConstruct`, `SConscript`, `.tac`, `.sage`},
		[]string{`text/x-python`, `application/x-python`},
		map[string][]RegexpRule{
			`root`: {
				MS.MatcherToken(tripleQMatcher, String),
				MS.MatcherToken(StringMatcher('"'), String),
				MS.MatcherToken(StringMatcher('\''), String),
				MS.MatcherToken(SingleLineCommentMatcher("#"), Comment),
				MS.Token(`[]{}:(),;[]`, Punctuation),
				MS.Consume("\\\n", true),
				MS.MatcherToken(Words(`in`, `is`, `and`, `or`, `not`), Operator_Word),
				MS.MatcherToken(
					Words(`!=`, `==`, `<<`, `>>`, `-`, `~`, `+`, `/`, `*`, `%`, `=`, `<`, `>`, `&`, `^`, `|`, `.`),
					Operator),
				Include(`keywords`),
				MS.MatcherToken(Words(`def`), Keyword, `funcname`),
				MS.MatcherToken(Words(`class`), Keyword, `classname`),
				MS.MatcherToken(Words(`from`), Keyword_Namespace, `fromimport`),
				MS.MatcherToken(Words(`import`), Keyword_Namespace, `import`),
				Include(`builtins`),
				Include(`backtick`),
				Include(`name`),
				Include(`numbers`),
			},
			`keywords`: {
				MS.MatcherToken(Words(
					`assert`, `break`, `continue`, `del`, `elif`, `else`, `except`,
					`exec`, `finally`, `for`, `global`, `if`, `lambda`, `pass`,
					`print`, `raise`, `return`, `try`, `while`, `yield`,
					`yield from`, `as`, `with`),
					Keyword),
			},
			`builtins`: {
				MS.MatcherToken(Words(
					`__import__`, `abs`, `all`, `any`, `apply`, `basestring`, `bin`,
					`bool`, `buffer`, `bytearray`, `bytes`, `callable`, `chr`, `classmethod`,
					`cmp`, `coerce`, `compile`, `complex`, `delattr`, `dict`, `dir`, `divmod`,
					`enumerate`, `eval`, `execfile`, `exit`, `file`, `filter`, `float`,
					`frozenset`, `getattr`, `globals`, `hasattr`, `hash`, `hex`, `id`,
					`input`, `int`, `intern`, `isinstance`, `issubclass`, `iter`, `len`,
					`list`, `locals`, `long`, `map`, `max`, `min`, `next`, `object`,
					`oct`, `open`, `ord`, `pow`, `property`, `range`, `raw_input`, `reduce`,
					`reload`, `repr`, `reversed`, `round`, `set`, `setattr`, `slice`,
					`sorted`, `staticmethod`, `str`, `sum`, `super`, `tuple`, `type`,
					`unichr`, `unicode`, `vars`, `xrange`, `zip`), Name_Builtin),
				MS.MatcherToken(Words(`self`, `None`, `Ellipsis`, `NotImplemented`, `False`, `True`), Name_Builtin_Pseudo),
				MS.MatcherToken(Words(
					`ArithmeticError`, `AssertionError`, `AttributeError`,
					`BaseException`, `DeprecationWarning`, `EOFError`, `EnvironmentError`,
					`Exception`, `FloatingPointError`, `FutureWarning`, `GeneratorExit`,
					`IOError`, `ImportError`, `ImportWarning`, `IndentationError`,
					`IndexError`, `KeyError`, `KeyboardInterrupt`, `LookupError`,
					`MemoryError`, `NameError`, `NotImplemented`, `NotImplementedError`,
					`OSError`, `OverflowError`, `OverflowWarning`, `PendingDeprecationWarning`,
					`ReferenceError`, `RuntimeError`, `RuntimeWarning`, `StandardError`,
					`StopIteration`, `SyntaxError`, `SyntaxWarning`, `SystemError`,
					`SystemExit`, `TabError`, `TypeError`, `UnboundLocalError`,
					`UnicodeDecodeError`, `UnicodeEncodeError`, `UnicodeError`,
					`UnicodeTranslateError`, `UnicodeWarning`, `UserWarning`,
					`ValueError`, `VMSError`, `Warning`, `WindowsError`,
					`ZeroDivisionError`), Name_Exception),
			},
			`numbers`: {
				MS.MatcherToken(NumberMatcher("lLjJ"), Number),
				MS.MatcherToken(HexNumberMatcher, Number_Hex),
				MS.MatcherToken(binNumberMatcher, Number_Bin),
			},
			`backtick`: {
				MS.MatcherToken(backticksMatcher, String_Backtick),
			},
			`name`: {
				MS.Token(`@[\w.]+`, Name_Decorator),
				MS.Token(`[a-zA-Z_][\w.]*`, Name),
			},
			`funcname`: {
				MS.Token(`[a-zA-Z_][\w.]*`, Name_Function, `#pop`),
			},
			`classname`: {
				MS.Token(`[a-zA-Z_][\w.]*`, Name_Class, `#pop`),
			},
			`import`: {
				MS.Consume("\\\n", true),
				MS.MatcherToken(Words(`as`), Keyword_Namespace, `#pop`),
				MS.Token(`,`, Operator),
				MS.Token(`[a-zA-Z_][\w.]*`, Name_Namespace),
				Default(`#pop`),
			},
			`fromimport`: {
				MS.Consume("\\\n", true),
				MS.MatcherToken(Words(`import`), Keyword_Namespace, `#pop`),
				MS.MatcherToken(Words(`None`), Name_Builtin_Pseudo, `#pop`),
				MS.Token(`[a-zA-Z_][\w.]*`, Name_Namespace),
				Default(`#pop`),
			},
		})
}
