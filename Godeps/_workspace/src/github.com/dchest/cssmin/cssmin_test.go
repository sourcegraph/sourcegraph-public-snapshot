package cssmin

import (
	"testing"
)

func TestMinify(t *testing.T) {
	for i, v := range tests {
		min := Minify([]byte(v.input))
		if string(min) != v.output {
			t.Fatalf("%d: expected:\n%s\ngot:\n%s\n", i, v.output, min)

		}
	}
}

var tests = []struct {
	input, output string
}{
	{`
		body {
			padding: 0 0 0 0;
			margin-left: 0px;
			margin-right: 1px;
			color: #cceeff;
			background: #aabbc0;
		}

		.hello,
		div.welcome {
			background-color: red; /* or green? */
			color: rgb(0, 16, 255);
		}`,

		`body{padding:0;margin-left:0;margin-right:1px;color:#cef;background:#aabbc0}.hello,div.welcome{background-color:red;color:#0010ff}`,
	},

	// comments.css
	{`/*****
  Multi-line comment
  before a new class name
*****/
.classname {
    /* comment in declaration block */
    font-weight: normal;
}`,
		`.classname{font-weight:normal}`,
	},

	// media_queries.css
	{`@media screen and (-webkit-min-device-pixel-ratio:0) {
  some-css : here
}`,
		`@media screen and (-webkit-min-device-pixel-ratio:0){some-css:here}`,
	},
}
