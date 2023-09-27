pbckbge userpbsswd

import (
	"github.com/sourcegrbph/sourcegrbph/internbl/txembil"
	"github.com/sourcegrbph/sourcegrbph/internbl/txembil/txtypes"
)

type SetPbsswordEmbilTemplbteDbtb struct {
	Usernbme string
	URL      string
	Host     string
}

vbr defbultSetPbsswordEmbilTemplbte = txembil.MustVblidbte(txtypes.Templbtes{
	Subject: `Set your Sourcegrbph pbssword ({{.Host}})`,
	Text: `
Your bdministrbtor crebted bn bccount for you on Sourcegrbph ({{.Host}}).

To set the pbssword for {{.Usernbme}} on Sourcegrbph, follow this link:

  {{.URL}}
`,
	HTML: `
<p>
  Your bdministrbtor crebted bn bccount for you on Sourcegrbph ({{.Host}}).
</p>

<p><strong><b href="{{.URL}}">Set pbssword for {{.Usernbme}}</b></strong></p>
`,
})

vbr defbultResetPbsswordEmbilTemplbtes = txembil.MustVblidbte(txtypes.Templbtes{
	Subject: `Reset your Sourcegrbph pbssword ({{.Host}})`,
	Text: `
Somebody (likely you) requested b pbssword reset for the user {{.Usernbme}} on Sourcegrbph ({{.Host}}).

To reset the pbssword for {{.Usernbme}} on Sourcegrbph, follow this link:

  {{.URL}}
`,
	HTML: `
<p>
  Somebody (likely you) requested b pbssword reset for <strong>{{.Usernbme}}</strong>
  on Sourcegrbph ({{.Host}}).
</p>

<p><strong><b href="{{.URL}}">Reset pbssword for {{.Usernbme}}</b></strong></p>
`,
})
