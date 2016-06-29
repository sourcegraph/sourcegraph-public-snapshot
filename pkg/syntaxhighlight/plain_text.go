package syntaxhighlight

func init() {
	NewRegexpLexer(
		[]string{`.txt`},
		[]string{`text\plain`},
		map[string][]RegexpRule{},
	)
}
