package mailreply

import (
	"io/ioutil"
	"net/textproto"
	"strings"
	"testing"
)

func TestMessagePartTextContent(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		header    textproto.MIMEHeader
		want      string
		wantError string
	}{
		{
			name:  "content_type/UTF8",
			input: "Hello world!",
			header: textproto.MIMEHeader{
				"Content-Type": []string{`text/plain; charset="uTf-8"`},
			},
			want: "Hello world!",
		},
		{
			name:  "content_type/ASCII",
			input: "Hello world!",
			header: textproto.MIMEHeader{
				"Content-Type": []string{`text/plain; charset="us-ascii"`},
			},
			want: "Hello world!",
		},
		{
			name:  "content_type/ISO_8859_1",
			input: "Hello world!",
			header: textproto.MIMEHeader{
				"Content-Type": []string{`text/plain; charset=ISO-8859-1`},
			},
			want: "Hello world!",
		},
		{
			name:  "content_type/windows_1252",
			input: "Hello world!",
			header: textproto.MIMEHeader{
				"Content-Type": []string{`text/plain; charset="Windows-1252"`},
			},
			want: "Hello world!",
		},
		{
			name:   "content_type/none",
			input:  "Hello world!",
			header: textproto.MIMEHeader{},
			want:   "Hello world!",
		},
		{
			name:  "content_type/invalid",
			input: "Hello world!",
			header: textproto.MIMEHeader{
				"Content-Type": []string{`text/invalid; charset=utf-8`},
			},
			wantError: `content type "text/invalid; charset=utf-8" not supported`,
		},
		{
			name:  "content_type/charset_invalid",
			input: "Hello world!",
			header: textproto.MIMEHeader{
				"Content-Type": []string{`text/plain; charset=invalid`},
			},
			wantError: `unsupported charset: "invalid"`,
		},
		{
			// https://en.wikipedia.org/wiki/Quoted-printable#Example
			name: "encoding/quotedprintable",
			input: `J'interdis aux marchands de vanter trop leur marchandises. Car ils se font =
vite p=C3=A9dagogues et t'enseignent comme but ce qui n'est par essence qu'=
un moyen, et te trompant ainsi sur la route =C3=A0 suivre les voil=C3=A0 bi=
ent=C3=B4t qui te d=C3=A9gradent, car si leur musique est vulgaire ils te f=
abriquent pour te la vendre une =C3=A2me vulgaire.`,
			header: textproto.MIMEHeader{
				"Content-Type":              []string{`text/plain; charset=UTF-8`},
				"Content-Transfer-Encoding": []string{`quoted-printable`},
			},
			want: "J'interdis aux marchands de vanter trop leur marchandises. Car ils se font vite pédagogues et t'enseignent comme but ce qui n'est par essence qu'un moyen, et te trompant ainsi sur la route à suivre les voilà bientôt qui te dégradent, car si leur musique est vulgaire ils te fabriquent pour te la vendre une âme vulgaire.",
		},
		{
			name:  "encoding/base64",
			input: `8J+YjvCfmI7wn5iO8J+YjvCfmI7wn5iO8J+YjvCfmI7wn5iO8J+YjvCfmI7wn5iO8J+YjvCfmI7wn5iO8J+YjvCfmI7wn5iO8J+YjvCfmI7wn5iODQo=`,
			header: textproto.MIMEHeader{
				"Content-Type":              []string{`text/plain; charset=UTF-8`},
				"Content-Transfer-Encoding": []string{`base64`},
			},
			want: "😎😎😎😎😎😎😎😎😎😎😎😎😎😎😎😎😎😎😎😎😎\r\n",
		},
		{
			name:  "encoding/7bit",
			input: "Test message sent in iOS mail",
			header: textproto.MIMEHeader{
				"Content-Type":              []string{`text/plain; charset=us-ascii`},
				"Content-Transfer-Encoding": []string{`7bit`},
			},
			want: "Test message sent in iOS mail",
		},
		{
			name:  "encoding/8bit",
			input: "Test message sent in ??? mail",
			header: textproto.MIMEHeader{
				"Content-Type":              []string{`text/plain; charset=us-ascii`},
				"Content-Transfer-Encoding": []string{`8bit`},
			},
			want: "Test message sent in ??? mail",
		},
		{
			name:  "encoding/invalid",
			input: "Hello world!",
			header: textproto.MIMEHeader{
				"Content-Type":              []string{`text/plain; charset=us-ascii`},
				"Content-Transfer-Encoding": []string{`invalid`},
			},
			wantError: "", // should silently fail so that we 'skip' this part
			want:      "",
		},
	}
	for _, tst := range tests {
		t.Run(tst.name, func(t *testing.T) {
			r, err := messagePartTextContent(
				ioutil.NopCloser(strings.NewReader(tst.input)),
				tst.header,
			)
			if tst.wantError != "" {
				if err == nil {
					t.Fatalf("got no error, want %q\n", tst.wantError)
				}
				if err.Error() != tst.wantError {
					t.Fatalf("got %q want %q\n", err.Error(), tst.wantError)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			got := ""
			if r != nil {
				gotData, err := ioutil.ReadAll(r)
				if err != nil {
					t.Fatal(err)
				}
				got = string(gotData)
			}
			if got != tst.want {
				t.Fatalf("got %q want %q\n", got, tst.want)
			}
		})
	}
}
