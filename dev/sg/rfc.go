package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/cockroachdb/errors"
	"github.com/sourcegraph/sourcegraph/dev/sg/root"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

// Retrieve a token, saves the token, then returns the generated client.
func getClient(ctx context.Context, config *oauth2.Config) (*http.Client, error) {
	root, err := root.RepositoryRoot()
	if err != nil {
		return nil, err
	}

	tokFile := filepath.Join(root, ".sg.token.json")
	tok, err := tokenFromFile(tokFile)
	if err != nil {
		tok, err = getTokenFromWeb(ctx, config)
		if err != nil {
			return nil, err
		}
		saveToken(tokFile, tok)
	}
	return config.Client(ctx, tok), nil
}

// Request a token from the web, then returns the retrieved token.
func getTokenFromWeb(ctx context.Context, config *oauth2.Config) (*oauth2.Token, error) {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)

	if err := exec.Command("xdg-open", authURL).Run(); err != nil {
		return nil, err
	}

	fmt.Printf("Paste the OAuth token here: ")
	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		return nil, err
	}

	tok, err := config.Exchange(ctx, authCode)
	if err != nil {
		return nil, err
	}
	return tok, nil
}

// Retrieves a token from a local file.
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

// Saves a token to a file path.
func saveToken(path string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}

const credentials = `{"installed":{"client_id":"1043390970557-1okrt0mo0qt2ogn2mkp217cfrirr1rfd.apps.googleusercontent.com","project_id":"sg-cli","auth_uri":"https://accounts.google.com/o/oauth2/auth","token_uri":"https://oauth2.googleapis.com/token","auth_provider_x509_cert_url":"https://www.googleapis.com/oauth2/v1/certs","client_secret":"gkQ2alKQZr2088IFGr55ET_I","redirect_uris":["urn:ietf:wg:oauth:2.0:oob","http://localhost"]}}`

func rfcExec(ctx context.Context, args []string) error {
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

	q := "name contains 'RFC' and parents in '1zP3FxdDlcSQGC1qvM9lHZRaHH4I9Jwwa' or name contains 'RFC' and parents in '1KCq4tMLnVlC0a1rwGuU5OSCw6mdDxLuv'"

	r, err := srv.Files.List().
		Corpora("drive").SupportsAllDrives(true).
		DriveId("0AK4DcztHds_pUk9PVA").
		IncludeItemsFromAllDrives(true).
		SupportsAllDrives(true).
		PageSize(100).
		Q(q).
		Fields("nextPageToken, files(id, name, parents)").
		OrderBy("createdTime,name").
		Do()
	if err != nil {
		log.Fatalf("Unable to retrieve files: %v", err)
	}
	fmt.Println("RFCs:")
	if len(r.Files) == 0 {
		fmt.Println("No files found.")
	} else {
		for _, i := range r.Files {
			fmt.Printf("%s (%s)\n", i.Name, i.Id)
		}
	}

	return nil
}
