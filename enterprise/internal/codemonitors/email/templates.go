package email

import (
	"github.com/sourcegraph/sourcegraph/internal/txemail"
	"github.com/sourcegraph/sourcegraph/internal/txemail/txtypes"
)

var newSearchResultsEmailTemplates = txemail.MustValidate(txtypes.Templates{
	Subject: `{{ if .IsTest }}Test: {{ end }}[{{.Priority}} event] {{.Description}}`,
	Text: `
{{ if .IsTest }}This email is a preview. Links are disabled.{{ end }}

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
	{{ if .IsTest }}
	<p style="color: #523704; padding: 16px; background-color: #FDECCC; font-size: 14px; line-height: 21px; border-radius: 4px; margin-bottom: 50px">
		<span style="font-weight: 700">This email is a preview.</span>&nbsp;<span style="font-weight: 400">Links are disabled.</span>
	</p>
	{{ end }}

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
	  <a href="{{.SearchURL}}" {{ if .IsTest }}style="color: #9C9FA6; font-weight: 400; text-decoration: underline; cursor: default"{{ end }}>
        View search on Sourcegraph
	  </a>
    </p>
    <br />
    <br />
    __
    <p style="font-size: 14px; line-height: 24px">
      You are receiving this notification because you are a recipient on a code
      monitor.
    </p>
    <p style="font-size: 14px; line-height: 24px">
	  <a href="{{.CodeMonitorURL}}" {{ if .IsTest }}style="color: #9C9FA6; font-weight: 400; text-decoration: underline; cursor: default"{{ end }}>
	    View code monitor
	  </a>
    </p>
    <p style="font-size: 12px; line-height: 24px; margin-bottom: 24px">
      Search results may contain confidential data. To protect your privacy and
      security, Sourcegraph limits what information is contained in this
      notification.
	</p>
	<img src="https://about.sourcegraph.com/sourcegraph-logo-small.png" width="106" height="20" alt="Sourcegraph logo" />
  </body>
</html>
`,
})
