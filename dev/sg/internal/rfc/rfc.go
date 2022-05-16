package rfc

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/grafana/regexp"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/open"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/secrets"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/output"
)

const (
	credentials = `{"installed":{"client_id":"1043390970557-1okrt0mo0qt2ogn2mkp217cfrirr1rfd.apps.googleusercontent.com","project_id":"sg-cli","auth_uri":"https://accounts.google.com/o/oauth2/auth","token_uri":"https://oauth2.googleapis.com/token","auth_provider_x509_cert_url":"https://www.googleapis.com/oauth2/v1/certs","client_secret":"gkQ2alKQZr2088IFGr55ET_I","redirect_uris":["urn:ietf:wg:oauth:2.0:oob","http://localhost"]}}` // CI:LOCALHOST_OK
)

// Retrieve a token, saves the token, then returns the generated client.
func getClient(ctx context.Context, config *oauth2.Config, out *output.Output) (*http.Client, error) {
	sec, err := secrets.FromContext(ctx)
	if err != nil {
		return nil, err
	}
	tok := &oauth2.Token{}
	if err := sec.Get("rfc", tok); err != nil {
		// ...if it doesn't exist, open browser and ask user to give us
		// permissions
		tok, err = getTokenFromWeb(ctx, config, out)
		if err != nil {
			return nil, err
		}
		err := sec.PutAndSave("rfc", tok)
		if err != nil {
			return nil, err
		}
	}

	return config.Client(ctx, tok), nil
}

// Request a token from the web, then returns the retrieved token.
func getTokenFromWeb(ctx context.Context, config *oauth2.Config, out *output.Output) (*oauth2.Token, error) {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)

	out.Writef("Opening %s ...", authURL)
	if err := open.URL(authURL); err != nil {
		return nil, err
	}

	fmt.Printf("Paste the OAuth token here: ")
	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		return nil, err
	}

	return config.Exchange(ctx, authCode)
}

func queryRFCs(ctx context.Context, query string, orderBy string, pager func(r *drive.FileList) error, out *output.Output) error {
	// If modifying these scopes, delete your previously saved token.json.
	config, err := google.ConfigFromJSON([]byte(credentials), drive.DriveMetadataReadonlyScope)
	if err != nil {
		return errors.Wrap(err, "Unable to parse client secret file to config")
	}
	client, err := getClient(ctx, config, out)
	if err != nil {
		return errors.Wrap(err, "Unable to build client")
	}

	srv, err := drive.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return errors.Wrap(err, "Unable to retrieve Drive client")
	}

	if query == "" {
		query = "name contains 'RFC'"
	}
	q := fmt.Sprintf("%s and parents in '1zP3FxdDlcSQGC1qvM9lHZRaHH4I9Jwwa' or %s and parents in '1KCq4tMLnVlC0a1rwGuU5OSCw6mdDxLuv'", query, query)

	list := srv.Files.List().
		Corpora("drive").SupportsAllDrives(true).
		DriveId("0AK4DcztHds_pUk9PVA").
		IncludeItemsFromAllDrives(true).
		SupportsAllDrives(true).
		PageSize(100).
		Q(q).
		Fields("nextPageToken, files(id, name, parents)")

	if orderBy != "" {
		list = list.OrderBy(orderBy)
	}

	return list.Pages(ctx, pager)
}

func List(ctx context.Context, out *output.Output) error {
	return queryRFCs(ctx, "", "createdTime,name", rfcTitlesPrinter(out), out)
}

func Search(ctx context.Context, query string, out *output.Output) error {
	return queryRFCs(ctx, fmt.Sprintf("(name contains '%s' or fullText contains '%s')", query, query), "", rfcTitlesPrinter(out), out)
}

func Open(ctx context.Context, number string, out *output.Output) error {
	return queryRFCs(ctx, fmt.Sprintf("name contains 'RFC %s'", number), "", func(r *drive.FileList) error {
		for _, f := range r.Files {
			open.URL(fmt.Sprintf("https://docs.google.com/document/d/%s/edit", f.Id))
		}
		return nil
	}, out)
}

var rfcTitleRegex = regexp.MustCompile(`RFC\s(\d+):*\s(\w+):\s(.*)$`)

func rfcTitlesPrinter(out *output.Output) func(r *drive.FileList) error {
	return func(r *drive.FileList) error {
		if len(r.Files) == 0 {
			return nil
		}

		for _, i := range r.Files {
			matches := rfcTitleRegex.FindStringSubmatch(i.Name)
			if len(matches) == 4 {
				number := matches[1]
				status := strings.ToUpper(matches[2])
				name := matches[3]

				var statusColor output.Style
				switch strings.ToUpper(status) {
				case "WIP":
					statusColor = output.StylePending
				case "REVIEW":
					statusColor = output.Fg256Color(208)
				case "IMPLEMENTED", "APPROVED":
					statusColor = output.StyleSuccess
				case "ABANDONED", "PAUSED":
					statusColor = output.StyleSearchAlertTitle
				}

				numberColor := output.Fg256Color(8)

				out.Writef("RFC %s%s %s%s%s %s", numberColor, number, statusColor, status, output.StyleReset, name)
			} else {
				out.Writef("%s%s", i.Name, output.StyleReset)
			}
		}

		return nil
	}

}
