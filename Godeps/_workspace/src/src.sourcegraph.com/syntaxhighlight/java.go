package syntaxhighlight

// Matches string literals ("...")
func javaString(source []byte) []int {
	state := 0
	for i, c := range source {
		switch state {
		case 0:
			if c == '"' {
				state = 1
			} else {
				return nil
			}
			break
		case 1:
			if c == '\\' {
				state = 2
			} else if c == '"' {
				return []int{0, i + 1}
			}
			break
		case 2:
			state = 1
			break
		}
	}
	return nil
}

func init() {
	NewRegexpLexer(
		[]string{`.java`},
		[]string{`text\x-java`},
		map[string][]RegexpRule{
			`root`: {
				MS.MatcherToken(SingleLineCommentMatcher, Comment_Single),
				MS.MatcherToken(MultiLineCommentMatcher, Comment_Multiline),
				// keywords: go before method names to avoid lexing "throw new XYZ" as a method signature
				MS.MatcherToken(Words(`assert`, `break`, `case`, `catch`, `continue`, `default`, `do`, `else`,
					`finally`, `for`, `if`, `goto`, `instanceof`, `new`, `return`, `switch`, `this`, `throw`,
					`try`, `while`), Keyword),
				// method names
				MS.Action(`((?:(?:[^\W\d]|\$)[\w.\[\]$<>]*\s+)+?)((?:[^\W\d]|\$)[\w$]*)(?:\s*)(\()`,
					ByGroups(UsingThis(), Name_Function, Operator)),
				MS.Token(`@[^\W\d][\w.]*`, Name_Decorator),
				MS.MatcherToken(Words(`abstract`, `const`, `enum`, `extends`, `final`, `implements`, `native`,
					`private`, `protected`, `public`, `static`, `strictfp`, `super`, `synchronized`, `throws`,
					`transient`, `volatile`), Keyword_Declaration),
				MS.MatcherToken(Words(`boolean`, `byte`, `char`, `double`, `float`, `int`, `long`, `short`, `void`),
					Keyword_Type),
				MS.MatcherToken(Words(`package`, `import`), Keyword_Namespace, `import`),
				MS.MatcherToken(Words(`true`, `false`, `null`), Keyword_Constant),
				MS.MatcherToken(Words(`class`, `interface`), Keyword_Declaration, `class`),
				MS.MatcherToken(javaString, String),
				// MS.Token(`'\.'|'[^\]'|'\\u[0-9a-fA-F]{4}'`, String_Char),
				MS.MatcherToken(JavaCharMatcher, String_Char),
				MS.Action(`(\.)((?:[^\W\d]|\$)[\w$]*)`, ByGroups(Operator, Name_Attribute)),
				MS.Token(`^\s*([^\W\d]|\$)[\w$]*:`, Name_Label),
				MS.Token(`([^\W\d]|\$)[\w$]*`, Name),
				MS.Token(`[~^*!%&\[\](){}<>|+=:;,./?-]`, Operator),
				MS.Token(`[0-9][0-9]*\.[0-9]+([eE][0-9]+)?[fd]?`, Number_Float),
				MS.Token(`0x[0-9a-fA-F]+`, Number_Hex),
				MS.Token(`[0-9]+(_+[0-9]+)*L?`, Number_Integer),
			},
			`class`: {
				MS.Token(`([^\W\d]|\$)[\w$]*`, Name_Class, `#pop`),
			},
			`import`: {
				MS.Token(`[\w.]+\*?`, Name_Namespace, `#pop`),
			},
		})
}
