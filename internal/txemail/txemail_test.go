package txemail

import (
	"testing"

	"github.com/hexops/autogold/v2"
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
		b, err := got.Bytes()
		require.NoError(t, err)
		autogold.Expect(`References: <ref1> <ref2>
Message-ID: 1
Reply-To: admin@sourcegraph.com
Date: Thu, 23 Mar 2023 22:04:44 +0900
Message-Id: <1679576684882150000.76848.6468314242858748196@bobbook-sourcegraph.local>
From: <foo@sourcegraph.com>
Mime-Version: 1.0
Content-Type: multipart/alternative;
 boundary=8a2ba19880a8635f7a824113c5c07af4dec92ea6b5a2a856a0adfa2a5d31
To: <bar1@sourcegraph.com>, <bar2@sourcegraph.com>
Subject: a subject <b>

--8a2ba19880a8635f7a824113c5c07af4dec92ea6b5a2a856a0adfa2a5d31
Content-Transfer-Encoding: quoted-printable
Content-Type: text/plain; charset=UTF-8

a text body <b>
--8a2ba19880a8635f7a824113c5c07af4dec92ea6b5a2a856a0adfa2a5d31
Content-Transfer-Encoding: quoted-printable
Content-Type: text/html; charset=UTF-8

a html body <span class=3D"&lt;b&gt;" />
--8a2ba19880a8635f7a824113c5c07af4dec92ea6b5a2a856a0adfa2a5d31--
`).Equal(t, string(b))
	})

	t.Run("email and sender name", func(t *testing.T) {
		got, err := render("foo@sourcegraph.com", "foo", msg)
		require.NoError(t, err)
		b, err := got.Bytes()
		require.NoError(t, err)
		autogold.Expect(`Content-Type: multipart/alternative;
 boundary=9120bd2820650184967d3cbb88e56ae66d34b0f8905cc194f0c517004b06
Date: Thu, 23 Mar 2023 22:04:45 +0900
Message-ID: 1
References: <ref1> <ref2>
Message-Id: <1679576685006321000.76848.5675638794021296688@bobbook-sourcegraph.local>
From: "foo" <foo@sourcegraph.com>
Mime-Version: 1.0
Reply-To: admin@sourcegraph.com
To: <bar1@sourcegraph.com>, <bar2@sourcegraph.com>
Subject: a subject <b>

--9120bd2820650184967d3cbb88e56ae66d34b0f8905cc194f0c517004b06
Content-Transfer-Encoding: quoted-printable
Content-Type: text/plain; charset=UTF-8

a text body <b>
--9120bd2820650184967d3cbb88e56ae66d34b0f8905cc194f0c517004b06
Content-Transfer-Encoding: quoted-printable
Content-Type: text/html; charset=UTF-8

a html body <span class=3D"&lt;b&gt;" />
--9120bd2820650184967d3cbb88e56ae66d34b0f8905cc194f0c517004b06--
`).Equal(t, string(b))
	})
}
