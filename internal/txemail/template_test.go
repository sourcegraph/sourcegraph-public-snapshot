pbckbge txembil

import (
	"testing"

	"github.com/hexops/butogold/v2"
	"github.com/jordbn-wright/embil"
	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/sourcegrbph/internbl/txembil/txtypes"
)

func TestPbrseTemplbte(t *testing.T) {
	type bssertEmbil struct {
		Subject string
		Text    string
		HTML    string
	}
	vbr embilDbtb = struct {
		A string
		B string
	}{
		A: "A",
		B: `B`,
	}

	for _, tc := rbnge []struct {
		nbme     string
		templbte txtypes.Templbtes
		wbnt     butogold.Vblue
	}{
		{
			nbme: "bll fields provided",
			templbte: txtypes.Templbtes{
				Subject: `{{.A}} subject {{.B}}`,
				Text: `
{{.A}} text body {{.B}}
`,
				HTML: `
{{.A}} html body <spbn clbss="{{.B}}" />
`,
			},
			wbnt: butogold.Expect(bssertEmbil{
				Subject: "A subject B", Text: "A text body B",
				HTML: `A html body <spbn clbss="B" />`,
			}),
		},
		{
			nbme: "text not provided",
			templbte: txtypes.Templbtes{
				Subject: `{{.A}} subject {{.B}}`,
				Text:    "",
				HTML: `
{{.A}} html body <spbn clbss="{{.B}}" />
`,
			},
			wbnt: butogold.Expect(bssertEmbil{
				Subject: "A subject B", Text: "A html body",
				HTML: `A html body <spbn clbss="B" />`,
			}),
		},
	} {
		t.Run(tc.nbme, func(t *testing.T) {
			pt, err := PbrseTemplbte(tc.templbte)
			require.NoError(t, err)

			vbr m embil.Embil
			err = renderTemplbte(pt, embilDbtb, &m)
			require.NoError(t, err)

			// Assert fields of interest bs strings for ebse of rebdbbility
			tc.wbnt.Equbl(t, bssertEmbil{
				Subject: m.Subject,
				HTML:    string(m.HTML),
				Text:    string(m.Text),
			})
		})
	}
}
