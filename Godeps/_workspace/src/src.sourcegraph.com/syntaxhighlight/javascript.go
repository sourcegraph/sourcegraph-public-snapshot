package syntaxhighlight

// TODO: Unicode support
var jsIdentStart = `(?:[$_` + `A-Za-z0-9` + /*uni.combine('Lu', 'Ll', 'Lt', 'Lm', 'Lo', 'Nl') + */
	`]|\\\\u[a-fA-F0-9]{4})`
var jsIdentPart = `(?:[$` + `A-Za-z0-9` +
	/*uni.combine('Lu', 'Ll', 'Lt', 'Lm', 'Lo', 'Nl', 'Mn', 'Mc', 'Nd', 'Pc') + u'\u200c\u200d + */
	`]|\\\\u[a-fA-F0-9]{4})`
var jsIdent = jsIdentStart + `(?:` + jsIdentPart + `)*`

// Matches division operand
func jsDivision(source []byte) []int {
	if source[0] == '/' {
		if len(source) < 2 {
			return nil
		}
		if source[1] == ' ' || source[1] == '\t' || source[1] == '\n' || source[1] == '\v' ||
			source[1] == '\f' || source[1] == '\r' ||
			source[1] == 0x85 || source[1] == 0xA0 {
			return []int{0, 1}
		}
	}
	return nil
}

func init() {
	NewRegexpLexer(
		[]string{`.js`},
		[]string{`application/javascript`, `application/x-javascript`, `text/x-javascript`, `text/javascript`},
		map[string][]RegexpRule{
			`commentsandwhitespace`: {
				MS.Token(`<!--`, Comment),
				MS.MatcherToken(SingleLineCommentMatcher("//"), Comment_Single),
				MS.MatcherToken(MultiLineCommentMatcher, Comment_Multiline),
			},
			`slashstartsregex`: {
				Include(`commentsandwhitespace`),
				MS.MatcherToken(jsDivision, Operator),
				MS.Token(`/(\\.|[^[/\\\n]|\[(\\.|[^\]\\\n])*])+/([gim]+\b|\B)`, String_Regex, `#pop`),
				MS.Lookahead(`/`, `#pop`),
				Default(`#pop`),
			},
			`root`: {
				MS.MatcherToken(SingleLineCommentMatcher("#"), Comment),
				MS.MatcherToken(jsDivision, Operator),
				MS.Lookahead(`(?:/[^\s]|<!--)`, `slashstartsregex`),
				Include(`commentsandwhitespace`),
				MS.MatcherToken(WordsWithBoundary(false, `++`, `--`, `~`, `&&`, `?`, `:`, `||`, `\`,
					`<<`, `>>`, `>>>`, `=`, `==`, `!`, `!=`, `-`, `<`, `>`, `+`, `*`, `%`, `&`, `|`, `^`,
					`<<=`, `>>=`, `>>>=`, `===`, `!==`, `-=`, `<=`, `>=`, `+=`, `*=`, `%=`, `&=`, `|=`, `/=`, `^=`),
					Operator, `slashstartsregex`),
				//            MS.Token(`(?:\+\+|--|~|&&|\?|:|\|\||\\(?:\n)|(<<|>>>?|==?|!=?|[-<>+*%&|^/])=?)`, Operator, `slashstartsregex`),
				MS.Token(`[{(\[;,]`, Punctuation, `slashstartsregex`),
				MS.Token(`[})\].]`, Punctuation),
				MS.MatcherToken(Words(`for`, `in`, `while`, `do`, `break`, `return`, `continue`, `switch`, `case`,
					`default`, `if`, `else`, `throw`, `try`, `catch`, `finally`, `new`, `delete`, `typeof`,
					`instanceof`, `void`, `yield`, `this`), Keyword, `slashstartsregex`),
				MS.MatcherToken(Words(`var`, `let`, `with`, `function`), Keyword_Declaration, `slashstartsregex`),
				MS.MatcherToken(Words(`abstract`, `boolean`, `byte`, `char`, `class`, `const`, `debugger`, `double`,
					`enum`, `export`, `extends`, `final`, `float`, `goto`, `implements`, `import`, `int`, `interface`,
					`long`, `native`, `package`, `private`, `protected`, `public`, `short`, `static`, `super`,
					`synchronized`, `throws`, `transient`, `volatile`), Keyword_Reserved),
				MS.MatcherToken(Words(`true`, `false`, `null`, `NaN`, `Infinity`, `undefined`), Keyword_Constant),
				MS.MatcherToken(Words(`Array`, `Boolean`, `Date`, `Error`, `Function`, `Math`, `netscape`, `Number`,
					`Object`, `Packages`, `RegExp`, `String`, `sun`, `decodeURI`, `decodeURIComponent`, `encodeURI`,
					`encodeURIComponent`, `Error`, `eval`, `isFinite`, `isNaN`, `parseFloat`, `parseInt`, `document`,
					`this`, `window`), Name_Builtin),
				MS.Token(jsIdent, Name_Other),
				MS.Token(`[0-9][0-9]*\.[0-9]+([eE][0-9]+)?[fd]?`, Number_Float),
				MS.Token(`0x[0-9a-fA-F]+`, Number_Hex),
				MS.Token(`[0-9]+`, Number_Integer),
				MS.MatcherToken(StringMatcher('"'), String_Double),
				MS.MatcherToken(StringMatcher('\''), String_Single),
			},
		})
}
