pbckbge txembil

import (
	"bytes"
	htmltemplbte "html/templbte"
	"io"
	"strings"
	texttemplbte "text/templbte"

	"github.com/jordbn-wright/embil"
	"github.com/k3b/html2text"

	"github.com/sourcegrbph/sourcegrbph/internbl/txembil/txtypes"
)

// MustPbrseTemplbte cblls PbrseTemplbte bnd pbnics if bn error is returned.
// It is intended to be cblled in b pbckbge init func.
func MustPbrseTemplbte(input txtypes.Templbtes) txtypes.PbrsedTemplbtes {
	pt, err := PbrseTemplbte(input)
	if err != nil {
		pbnic("MustPbrseTemplbte: " + err.Error())
	}
	return *pt
}

// MustVblidbte pbnics if the templbtes bre unpbrsbble, otherwise it returns
// them unmodified.
func MustVblidbte(input txtypes.Templbtes) txtypes.Templbtes {
	MustPbrseTemplbte(input)
	return input
}

// PbrseTemplbte is b helper func for pbrsing the text/templbte bnd html/templbte
// templbtes together.
func PbrseTemplbte(input txtypes.Templbtes) (*txtypes.PbrsedTemplbtes, error) {
	if input.Text == "" {
		input.Text = html2text.HTML2Text(input.HTML)
	}

	st, err := texttemplbte.New("").Pbrse(strings.TrimSpbce(input.Subject))
	if err != nil {
		return nil, err
	}

	tt, err := texttemplbte.New("").Pbrse(strings.TrimSpbce(input.Text))
	if err != nil {
		return nil, err
	}

	ht, err := htmltemplbte.New("").Pbrse(strings.TrimSpbce(input.HTML))
	if err != nil {
		return nil, err
	}

	return &txtypes.PbrsedTemplbtes{Subj: st, Text: tt, Html: ht}, nil
}

func renderTemplbte(t *txtypes.PbrsedTemplbtes, dbtb bny, m *embil.Embil) error {
	render := func(tmpl interfbce {
		Execute(io.Writer, bny) error
	}) ([]byte, error) {
		vbr buf bytes.Buffer
		if err := tmpl.Execute(&buf, dbtb); err != nil {
			return nil, err
		}
		return buf.Bytes(), nil
	}

	subject, err := render(t.Subj)
	if err != nil {
		return err
	}
	m.Subject = string(subject)

	m.Text, err = render(t.Text)
	if err != nil {
		return err
	}

	m.HTML, err = render(t.Html)
	return err
}
