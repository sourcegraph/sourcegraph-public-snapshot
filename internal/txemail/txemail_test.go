package txemail

import (
	"net/textproto"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/jordan-wright/email"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/txemail/txtypes"
)

func TestRender(t *testing.T) {
	replyTo := "admin@sourcegraph.com"
	messageID := "1"

	msg := Message{
		To:         []string{"bar1@sourcegraph.com", "bar2@sourcegraph.com"},
		ReplyTo:    &replyTo,
		MessageID:  &messageID,
		References: []string{"ref1", "ref2"},
		Template: txtypes.Templates{
			Subject: `{{.A}} subject {{.B}}`,
			Text: `
	{{.A}} text body {{.B}}
	`,
			HTML: `
	{{.A}} html body <span class="{{.B}}" />
	`,
		},
		Data: struct {
			A string
			B string
		}{
			A: "a",
			B: `<b>`,
		},
	}

	t.Run("only email", func(t *testing.T) {
		got, err := render("foo@sourcegraph.com", "", msg)
		require.NoError(t, err)
		if diff := cmp.Diff(&email.Email{
			ReplyTo: []string{replyTo},
			From:    "<foo@sourcegraph.com>",
			To:      []string{"bar1@sourcegraph.com", "bar2@sourcegraph.com"},
			Subject: "a subject <b>",
			Text:    []byte("a text body <b>"),
			HTML:    []byte(`a html body <span class="&lt;b&gt;" />`),
			Headers: textproto.MIMEHeader{
				"Message-ID": []string{messageID},
				"References": []string{"<ref1> <ref2>"},
			},
		}, got); diff != "" {
			t.Errorf("mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("email and sender name", func(t *testing.T) {
		got, err := render("foo@sourcegraph.com", "foo", msg)
		require.NoError(t, err)
		if diff := cmp.Diff(&email.Email{
			ReplyTo: []string{replyTo},
			From:    `"foo" <foo@sourcegraph.com>`,
			To:      []string{"bar1@sourcegraph.com", "bar2@sourcegraph.com"},
			Subject: "a subject <b>",
			Text:    []byte("a text body <b>"),
			HTML:    []byte(`a html body <span class="&lt;b&gt;" />`),
			Headers: textproto.MIMEHeader{
				"Message-ID": []string{messageID},
				"References": []string{"<ref1> <ref2>"},
			},
		}, got); diff != "" {
			t.Errorf("mismatch (-want +got):\n%s", diff)
		}
	})
}
