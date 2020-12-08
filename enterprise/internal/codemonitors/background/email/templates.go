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
{{.NumberOfResultsWithDetail}}

View search on Sourcegraph {{.SearchURL}}
	
__
You are receiving this notification because you are a recipient on a code monitor.

View code monitor: {{.CodeMonitorURL}}

Search results may contain confidential data. To protect your privacy and security,
Sourcegraph limits what information is contained in this notification.
`,
	HTML: `
<!DOCTYPE html>
<html>
  <body>
    <p style="font-size: 16px; line-height: 24px">
      Code monitoring triggered a new event:
    </p>
    <p style="font-size: 20px; line-height: 30px; font-weight: 700">
      {{.Description}}<br />
      <span style="font-size: 16px; line-height: 24px; font-weight: 400"
        >{{.NumberOfResultsWithDetail}}</span
      >
    </p>
    <p style="font-size: 16px; line-height: 24px">
      <a href="{{.SearchURL}}">View search on Sourcegraph</a>
    </p>
    <br />
    <br />
    __
    <p style="font-size: 14px; line-height: 24px">
      You are receiving this notification because you are a recipient on a code
      monitor.
    </p>
    <p style="font-size: 14px; line-height: 24px">
      <a href="{{.CodeMonitorURL}}">View code monitor</a>
    </p>
    <p style="font-size: 12px; line-height: 24px">
      Search results may contain confidential data. To protect your privacy and
      security, Sourcegraph limits what information is contained in this
      notification.
    </p>
  </body>
</html>
`,
})
