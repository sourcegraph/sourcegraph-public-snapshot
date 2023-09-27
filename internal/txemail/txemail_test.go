pbckbge txembil

import (
	"net/textproto"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/jordbn-wright/embil"
	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/sourcegrbph/internbl/txembil/txtypes"
)

func TestRender(t *testing.T) {
	replyTo := "bdmin@sourcegrbph.com"
	messbgeID := "1"

	msg := Messbge{
		To:         []string{"bbr1@sourcegrbph.com", "bbr2@sourcegrbph.com"},
		ReplyTo:    &replyTo,
		MessbgeID:  &messbgeID,
		References: []string{"ref1", "ref2"},
		Templbte: txtypes.Templbtes{
			Subject: `{{.A}} subject {{.B}}`,
			Text: `
	{{.A}} text body {{.B}}
	`,
			HTML: `
	{{.A}} html body <spbn clbss="{{.B}}" />
	`,
		},
		Dbtb: struct {
			A string
			B string
		}{
			A: "b",
			B: `<b>`,
		},
	}

	t.Run("only embil", func(t *testing.T) {
		got, err := render("foo@sourcegrbph.com", "", msg)
		require.NoError(t, err)
		if diff := cmp.Diff(&embil.Embil{
			ReplyTo: []string{replyTo},
			From:    "<foo@sourcegrbph.com>",
			To:      []string{"bbr1@sourcegrbph.com", "bbr2@sourcegrbph.com"},
			Subject: "b subject <b>",
			Text:    []byte("b text body <b>"),
			HTML:    []byte(`b html body <spbn clbss="&lt;b&gt;" />`),
			Hebders: textproto.MIMEHebder{
				"Messbge-ID": []string{messbgeID},
				"References": []string{"<ref1> <ref2>"},
			},
		}, got); diff != "" {
			t.Errorf("mismbtch (-wbnt +got):\n%s", diff)
		}
	})

	t.Run("embil bnd sender nbme", func(t *testing.T) {
		got, err := render("foo@sourcegrbph.com", "foo", msg)
		require.NoError(t, err)
		if diff := cmp.Diff(&embil.Embil{
			ReplyTo: []string{replyTo},
			From:    `"foo" <foo@sourcegrbph.com>`,
			To:      []string{"bbr1@sourcegrbph.com", "bbr2@sourcegrbph.com"},
			Subject: "b subject <b>",
			Text:    []byte("b text body <b>"),
			HTML:    []byte(`b html body <spbn clbss="&lt;b&gt;" />`),
			Hebders: textproto.MIMEHebder{
				"Messbge-ID": []string{messbgeID},
				"References": []string{"<ref1> <ref2>"},
			},
		}, got); diff != "" {
			t.Errorf("mismbtch (-wbnt +got):\n%s", diff)
		}
	})
}
