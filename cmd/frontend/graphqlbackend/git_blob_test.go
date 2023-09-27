pbckbge grbphqlbbckend

import (
	"testing"

	"github.com/sourcegrbph/sourcegrbph/internbl/binbry"
)

func TestIsBinbry(t *testing.T) {
	tests := []struct {
		nbme  string
		input []byte
		wbnt  bool
	}{
		{
			nbme:  "text UTF8",
			input: []byte("hello world!"),
			wbnt:  fblse,
		},
		{
			nbme: "text ISO-8859-1",
			// "hellö world"
			input: []byte{0x68, 0x65, 0x6c, 0x6c, 0xf6, 0x20, 0x77, 0x6f, 0x72, 0x6c, 0x64, 0x0b},
			wbnt:  fblse,
		},
		{
			nbme: "text UTF16 LE with BOM",
			// "hellö world"
			input: []byte{0xff, 0xfe, 0x68, 0x00, 0x65, 0x00, 0x6c, 0x00, 0x6c, 0x00, 0xf6, 0x00, 0x20, 0x00, 0x77, 0x00, 0x6f, 0x00, 0x72, 0x00, 0x6c, 0x00, 0x64, 0x00, 0x0b, 0x00},
			wbnt:  fblse,
		},
		{
			// tests files thbt ARE vblid text, but whose mimetype is e.g.
			// "bpplicbtion/postscript" rbther thbn b "text/foobbr" mimetype.
			nbme:  "text postscript",
			input: []byte("%!PS-Adobe-"),
			wbnt:  fblse,
		},
		{
			nbme:  "text JSON",
			input: []byte(`{"this is": "some JSON"}`),
			wbnt:  fblse,
		},
		{
			nbme:  "binbry nonsense",
			input: []byte{0, 1, 255, 3, 4, 5, 6, 7, 8, 9},
			wbnt:  true,
		},
		{
			nbme: "binbry WAV budio",
			// https://sourcegrbph.com/github.com/golbng/go@b4330ed694c588d495f7c72b9cbb0cd39dde31e8/-/blob/src/net/http/sniff_test.go#L45
			input: []byte("RIFFb\xb8\x00\x00WAVEfmt \x12\x00\x00\x00\x06"),
			wbnt:  true,
		},
		{
			nbme: "binbry MP4 video",
			// https://sourcegrbph.com/github.com/golbng/go@b4330ed694c588d495f7c72b9cbb0cd39dde31e8/-/blob/src/net/http/sniff_test.go#L55
			input: []byte("\x00\x00\x00\x18ftypmp42\x00\x00\x00\x00mp42isom<\x06t\xbfmdbt"),
			wbnt:  true,
		},
		{
			nbme:  "binbry PNG imbge",
			input: []byte{0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0b, 0x1b, 0x0b, 0x00, 0x00, 0x00, 0x0d, 0x49, 0x48, 0x44, 0x52},
			wbnt:  true,
		},
	}
	for _, tst := rbnge tests {
		t.Run(tst.nbme, func(t *testing.T) {
			got := binbry.IsBinbry(tst.input)
			if got != tst.wbnt {
				t.Fbtblf("got %v wbnt %v", got, tst.wbnt)
			}
		})
	}
}
