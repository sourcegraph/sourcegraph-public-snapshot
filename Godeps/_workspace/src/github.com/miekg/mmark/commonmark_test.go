// Unit tests for commonmark parsing

package mmark

import (
	"testing"
)

func runMarkdownCommonMark(input string, extensions int) string {
	htmlFlags := 0

	renderer := HtmlRenderer(htmlFlags, "", "")

	return Parse([]byte(input), renderer, extensions).String()
}

func doTestsCommonMark(t *testing.T, tests []string, extensions int) {
	// catch and report panics
	var candidate string
	defer func() {
		if err := recover(); err != nil {
			t.Errorf("\npanic while processing [%#v]: %s\n", candidate, err)
		}
	}()

	for i := 0; i+1 < len(tests); i += 2 {
		input := tests[i]
		candidate = input
		expected := tests[i+1]
		actual := runMarkdownCommonMark(candidate, extensions)
		if actual != expected {
			t.Errorf("\nInput   [%#v]\nExpected[%#v]\nActual  [%#v]",
				candidate, expected, actual)
		}

		// now test every substring to stress test bounds checking
		if !testing.Short() {
			for start := 0; start < len(input); start++ {
				for end := start + 1; end <= len(input); end++ {
					candidate = input[start:end]
					_ = runMarkdownBlock(candidate, extensions)
				}
			}
		}
	}
}

func TestPrefixHeaderCommonMark_29(t *testing.T) {
	var tests = []string{
		"# hallo\n\n # hallo\n\n  # hallo\n\n   # hallo\n\n    # hallo\n",
		"<h1>hallo</h1>\n\n<h1>hallo</h1>\n\n<h1>hallo</h1>\n\n<h1>hallo</h1>\n\n<pre><code># hallo\n</code></pre>\n",
	}
	doTestsCommonMark(t, tests, 0)
}

func TestHRuleCommonMark_18_22(t *testing.T) {
	var tests = []string{
		"*   List\n    * Sublist\n    Not a header\n    ------\n",
		"<ul>\n<li>List\n\n<ul>\n<li>Sublist\nNot a header</li>\n</ul>\n\n<hr></li>\n</ul>\n",

		"- Foo\n- * * *\n- Bar\n",
		"<ul>\n<li>Foo</li>\n<li>* * *</li>\n<li>Bar</li>\n</ul>\n",

		"* Foo\n* * * *\n",
		"<ul>\n<li>Foo</li>\n</ul>\n\n<hr>\n",

		"* Foo\n  * - - -\n",
		"<ul>\n<li>Foo\n\n<ul>\n<li>- - -</li>\n</ul></li>\n</ul>\n",
	}
	doTestsCommonMark(t, tests, 0)
}

func TestFencedCodeBlockCommonMark_81(t *testing.T) {
	var tests = []string{
		" ```\n aaa\naaa\n```\n",
		"<pre><code>aaa\naaa\n</code></pre>\n",

		"```\n bbb\nbbb\n```\n",
		"<pre><code> bbb\nbbb\n</code></pre>\n",

		"```\nbbb\nbbb\n",
		"<pre><code>bbb\nbbb\n</code></pre>\n",

		"~~~~    ruby\ndef foo(x)\n return 3\nend\n~~~~\n",
		"<pre><code class=\"language-ruby\">def foo(x)\n return 3\nend\n</code></pre>\n",

		"```\n[foo]: /url\n```\n",
		"<pre><code>[foo]: /url\n</code></pre>\n",
	}

	doTestsCommonMark(t, tests, EXTENSION_FENCED_CODE)
}
