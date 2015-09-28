package ace

import (
	"testing"
)

func TestLineCompile(t *testing.T) {
	for i, this := range []struct {
		template string
		expect   string
	}{
		{
			template: `a href=./foo foo`,
			expect:   `<a href="./foo">foo</a>`,
		},
		{
			template: `span.color-red red`,
			expect:   `<span class="color-red">red</span>`,
		},
		{
			template: `span#ref1 text`,
			expect:   `<span id="ref1">text</span>`,
		},
		{
			template: `span#ref1.color-red.text-big text`,
			expect:   `<span id="ref1" class="color-red text-big">text</span>`,
		},
		{
			template: `span.color-red#ref1.text-big text`,
			expect:   `<span id="ref1" class="color-red text-big">text</span>`,
		},
		{
			template: `#ref1 text`,
			expect:   `<div id="ref1">text</div>`,
		},
		{
			template: `#ref1.color-red.text-big text`,
			expect:   `<div id="ref1" class="color-red text-big">text</div>`,
		},
		{
			template: `.color-red#ref1.text-big text`,
			expect:   `<div id="ref1" class="color-red text-big">text</div>`,
		},
		{
			template: "div class=\"dialog {{ if eq .Attr `important` }}color-red{{end}}\" text",
			expect:   "<div class=\"dialog {{if eq .Attr `important`}}color-red{{end}}\">text</div>",
		},
		{
			template: "div class=\"dialog {{ if eq .Attr `important` }}color-red text-big{{end}}\" text",
			expect:   "<div class=\"dialog {{if eq .Attr `important`}}color-red text-big{{end}}\">text</div>",
		},
		{
			template: "div class=\"dialog {{ if eq .Attr \"important\" }}color-red{{end}}\" text",
			expect:   "<div class=\"dialog {{if eq .Attr \"important\"}}color-red{{end}}\">text</div>",
		},
		{
			template: "div class=\"dialog {{ if eq .Attr \"important\" }}color-red text-big{{end}}\" text",
			expect:   "<div class=\"dialog {{if eq .Attr \"important\"}}color-red text-big{{end}}\">text</div>",
		},
	} {
		name, filepath := "dummy", "dummy.ace"
		base := NewFile(filepath, []byte(this.template))
		inner := NewFile("", []byte{})

		src := NewSource(base, inner, []*File{})
		rslt, err := ParseSource(src, nil)
		if err != nil {
			t.Errorf("[%d] failed: %s", i, err)
			continue
		}

		tpl, err := CompileResult(name, rslt, nil)
		if err != nil {
			t.Errorf("[%d] failed: %s", i, err)
			continue
		}

		compiled := tpl.Lookup(name).Tree.Root.String()
		if compiled != this.expect {
			t.Errorf("[%d] Compiler didn't return an expected value, got %v but expected %v", i, compiled, this.expect)
		}
	}
}
