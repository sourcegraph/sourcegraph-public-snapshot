package doc

import "testing"

func TestChooseReadme(t *testing.T) {
	tests := []struct {
		filenames []string
		want      string
	}{
		{
			filenames: []string{"README.en.md", "README.md"},
			want:      "README.md",
		},
		{
			filenames: []string{"foo", "bar"},
			want:      "",
		},
		{
			filenames: []string{"README.txt"},
			want:      "README.txt",
		},
		{
			filenames: []string{"README"},
			want:      "README",
		},
		{
			filenames: []string{"readme"},
			want:      "readme",
		},
	}
	for _, test := range tests {
		got := ChooseReadme(test.filenames)
		if got != test.want {
			t.Errorf("got %q, want %q (filenames were: %v)", got, test.want, test.filenames)
		}
	}
}

func TestMarkdown(t *testing.T) {
	markdown := `
# Header

some paragraph text

![img](/img/foobar.png)

[click me](/yada/yada)
`
	want := `<h1>Header</h1>

<p>some paragraph text</p>

<p><img src="/img/foobar.png" alt="img"/></p>

<p><a href="/yada/yada" rel="nofollow">click me</a></p>
`
	got, err := ToHTML(Markdown, []byte(markdown))
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != want {
		t.Logf("got:\n%q\nwant:\n%q\n", string(got), want)
		t.Logf("")
		t.Fatalf("got:\n%s\nwant:\n%s\n", string(got), want)
	}
}

func TestMarkdownXSS(t *testing.T) {
	markdown := `
# Header

some paragraph text

[URL](javascript&#58window;document.write&#40;document.cookie&#41;)

[click me](javascript:alert("pwn"))

<script>
alert("pwned");
</script>
`
	want := `<h1>Header</h1>

<p>some paragraph text</p>

<p><a href="javascript&amp;#58window;document.write&amp;%2340;document.cookie&amp;%2341;" rel="nofollow">URL</a></p>

<p><a title="pwn">click me</a>)</p>


`
	got, err := ToHTML(Markdown, []byte(markdown))
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != want {
		t.Logf("got:\n%q\nwant:\n%q\n", string(got), want)
		t.Logf("")
		t.Fatalf("got:\n%s\nwant:\n%s\n", string(got), want)
	}
}
