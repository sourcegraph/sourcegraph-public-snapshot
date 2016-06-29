package syntaxhighlight

func init() {
	NewRegexpLexer(
		[]string{`.sh`, `.bash`},
		[]string{`text\x-sh`, `application\x-sh`},
		// TODO(mate): add an actual highlighter later
		map[string][]RegexpRule{},
	)
}
