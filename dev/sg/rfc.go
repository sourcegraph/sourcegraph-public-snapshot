package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	"github.com/cockroachdb/errors"
	"github.com/sourcegraph/sourcegraph/lib/output"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

const (
	tokenFileName = ".sg.token.json"
	credentials   = `{"installed":{"client_id":"1043390970557-1okrt0mo0qt2ogn2mkp217cfrirr1rfd.apps.googleusercontent.com","project_id":"sg-cli","auth_uri":"https://accounts.google.com/o/oauth2/auth","token_uri":"https://oauth2.googleapis.com/token","auth_provider_x509_cert_url":"https://www.googleapis.com/oauth2/v1/certs","client_secret":"gkQ2alKQZr2088IFGr55ET_I","redirect_uris":["urn:ietf:wg:oauth:2.0:oob","http://localhost"]}}` // CI:LOCALHOST_OK
)

func tokenFilePath() (string, error) {
	homedir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	tokenFilePath := filepath.Join(homedir, ".sourcegraph", tokenFileName)

	if err := os.MkdirAll(filepath.Dir(tokenFilePath), os.ModePerm); err != nil {
		return "", err
	}

	return tokenFilePath, nil
}

// Retrieve a token, saves the token, then returns the generated client.
func getClient(ctx context.Context, config *oauth2.Config) (*http.Client, error) {
	fp, err := tokenFilePath()
	if err != nil {
		return nil, err
	}

	// Try to read token from file...
	tok, err := readTokenFromFile(fp)
	if err != nil {
		// ...if it doesn't exist, open browser and ask user to give us
		// permissions
		tok, err = getTokenFromWeb(ctx, config)
		if err != nil {
			return nil, err
		}
		if err := saveToken(fp, tok); err != nil {
			return nil, err
		}
	}

	return config.Client(ctx, tok), nil
}

// Request a token from the web, then returns the retrieved token.
func getTokenFromWeb(ctx context.Context, config *oauth2.Config) (*oauth2.Token, error) {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)

	out.Writef("Opening %s ...", authURL)
	if err := openURL(authURL); err != nil {
		return nil, err
	}

	fmt.Printf("Paste the OAuth token here: ")
	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		return nil, err
	}

	return config.Exchange(ctx, authCode)
}

// Retrieves a token from a local file.
func readTokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	tok := &oauth2.Token{}
	return tok, json.NewDecoder(f).Decode(tok)
}

// Saves a token to a file path.
func saveToken(path string, token *oauth2.Token) error {
	fmt.Printf("Saving credential file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return errors.Wrap(err, "unable to cache oauth token")
	}
	defer f.Close()

	return json.NewEncoder(f).Encode(token)
}

func queryRFCs(ctx context.Context, query string, orderBy string, pager func(r *drive.FileList) error) error {
	// If modifying these scopes, delete your previously saved token.json.
	config, err := google.ConfigFromJSON([]byte(credentials), drive.DriveMetadataReadonlyScope)
	if err != nil {
		return errors.Wrap(err, "Unable to parse client secret file to config")
	}
	client, err := getClient(ctx, config)
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

func listRFCs(ctx context.Context) error {
	return queryRFCs(ctx, "", "createdTime,name", printRFCTitles)
}

func searchRFCs(ctx context.Context, query string) error {
	return queryRFCs(ctx, fmt.Sprintf("(name contains '%s' or fullText contains '%s')", query, query), "", printRFCTitles)
}

var rfcTitleRegex = regexp.MustCompile(`RFC\s(\d+):*\s(\w+):\s(.*)$`)

func printRFCTitles(r *drive.FileList) error {
	if len(r.Files) == 0 {
		return nil
	}

	for _, i := range r.Files {
		matches := rfcTitleRegex.FindStringSubmatch(i.Name)
		if len(matches) == 4 {
			number := matches[1]
			status := matches[2]
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

func openURL(url string) error {
	var err error
	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}
	return err
}

func openRFC(ctx context.Context, number string) error {
	return queryRFCs(ctx, fmt.Sprintf("name contains 'RFC %s'", number), "", func(r *drive.FileList) error {
		for _, f := range r.Files {
			openURL(fmt.Sprintf("https://docs.google.com/document/d/%s/edit", f.Id))
		}
		return nil
	})
}

func rfcExec(ctx context.Context, args []string) error {
	if len(args) == 0 {
		args = append(args, "list")
	}

	switch args[0] {
	case "list":
		return listRFCs(ctx)

	case "search":
		if len(args) != 2 {
			return errors.New("no search query given")
		}

		return searchRFCs(ctx, args[1])

	case "open":
		if len(args) != 2 {
			return errors.New("no number given")
		}

		return openRFC(ctx, args[1])
	}

	return nil
}
