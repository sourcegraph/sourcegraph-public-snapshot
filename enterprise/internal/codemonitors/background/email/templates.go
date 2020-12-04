package email

import (
	"github.com/sourcegraph/sourcegraph/internal/txemail"
	"github.com/sourcegraph/sourcegraph/internal/txemail/txtypes"
)

var newSearchResultsEmailTemplates = txemail.MustValidate(txtypes.Templates{
	Subject: `[{{.Priority}} event] {{.Description}}`,
	Text: `
Code monitoring triggered a new event:

{{.Description}}
{{.NumberOfResults}}

View search on Sourcegraph {{.SearchURL}}
	
__
You are receiving this notification because you are a recipient on a code monitor.

View code monitor: {{.CodeMonitorURL}}

Search results may contain confidential data. To protect your privacy and security,
Sourcegraph limits what information is contained in this notification.
`,
	HTML: `
<html>
<body>
<span style="font-size:16px;line-height:24px;font-weight:400">Code monitoring triggered a new event:</span>
<br>
<br>
<span style="font-size:20px;line-height:30px;font-weight:700">{{.Description}}</span>
<br>
<span style="font-size:16px;line-height:24px;font-weight:400">{{.NumberOfResults}}</span>
<br>
<br>
<span style="font-size:16px;line-height:24px;font-weight:400"><a href="{{.SearchURL}}">View search on Sourcegraph</a></span><br>
<br>
<br>
__
<br>
<span style="font-size:14px;line-height:24px;font-weight:400">You are receiving this notification because you are a recipient on a code monitor.</span>
<br>
<br>
<span style="font-size:14px;line-height:24px;font-weight:400"><a href="{{.CodeMonitorURL}}">View code monitor</a>
<br>
<br>
<span style="font-size:12px;line-height:24px;font-weight:400">Search results may contain confidential data. To protect your privacy and security,
<br>
Sourcegraph limits what information is contained in this notification.<span>
</body>
</html>
`,
})
