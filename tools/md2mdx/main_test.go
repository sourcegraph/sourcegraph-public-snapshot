package main

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConversion(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{
			name:  "html comments",
			input: "foobar\n<!-- comment -->\nbazbar",
			want:  "foobar\n{/* comment */}\nbazbar",
		},
		{
			name:  "less than",
			input: "foobar\n<!-- comment -->\nbazbar",
			want:  "foobar\n{/* comment */}\nbazbar",
		},
		{
			name:  "curlies inside single quotes",
			input: "foobar\n'query { foo { bar } }'\nbazbar",
			want:  "foobar\n" + `'query \{ foo \{ bar \} \}'` + "\nbazbar",
		},
		{
			name:  "curlies inside double quotes",
			input: "foobar\n\"query { foo { bar } }\"\nbazbar",
			want:  "foobar\n" + `"query \{ foo \{ bar \} \}"` + "\nbazbar",
		},
		{
			name:  "curlies inside are fine inside backticks",
			input: "foobar\n`query { foo { bar } }`\nbazbar",
			want:  "foobar\n`query { foo { bar } }`\nbazbar",
		},
		{
			name:  "but pipes are not (wtf) inside backticks",
			input: "foobar\n`query {.|jsonIndent}`\nbazbar",
			want:  "foobar\n`query {.\\|jsonIndent}`\nbazbar",
		},
		{
			name:  "real curlies example",
			input: `GraphQL query variables to include as JSON string, e.g. '{"var": "val", "var2": "val2"}' |`,
			want:  `GraphQL query variables to include as JSON string, e.g. '\{"var": "val", "var2": "val2"\}' |`,
		},
		{
			name:  "another real curlies example",
			input: "| `-f` | Format for the output, using the syntax of Go package text/template. (e.g. \"{{.|json}}\") | `{{.|jsonIndent}}` |",
			want:  "| `-f` | Format for the output, using the syntax of Go package text/template. (e.g. \"\\{\\{.\\|json\\}\\}\") | `{{.\\|jsonIndent}}` |",
		},
		{
			name:  "ignore triple backticks",
			input: "foobar\n```\n<!-- comment -->\n```\nbazbar",
			want:  "foobar\n```\n<!-- comment -->\n```\nbazbar",
		},
		{
			name:  "details tag",
			input: `<p class="subtitle">Hard error search API responses every 5m</p>`,
			want:  `<p class="subtitle">Hard error search API responses every 5m</p>`,
		},
		{
			name:  "latency is >50ms",
			input: `latency is >50ms but <50s`,
			want:  `latency is &gt;50ms but &lt;50s`,
		},
		// {
		// 	name:  "real tags example",
		// 	input: "| `-extension-id` | The <extID> in https://sourcegraph.com/extensions/<extID> (e.g. sourcegraph/java) |  |",
		// 	want:  "| `-extension-id` | The &lt;extID&gt; in https://sourcegraph.com/extensions/&lt;extID&gt; (e.g. sourcegraph/java) |  |",
		// },
		{
			name:  "relative links to .md should point to mdx",
			input: "* [`admin`](admin.md)\n* [`api`](api.md)\n* [`batch`](batch/index.md)",
			want:  "* [`admin`](references/admin.mdx)\n* [`api`](references/api.mdx)\n* [`batch`](references/batch)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := strings.NewReader(tt.input)
			out, err := convert(r, "foo/bar/references/index.md")

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.want, out)
		})
	}
}
