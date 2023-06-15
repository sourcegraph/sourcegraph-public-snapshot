package rfc

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/grafana/regexp"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/open"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/secrets"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/output"
)

var PublicDrive = DriveSpec{
	DisplayName: "Public",
	DriveID:     "0AIPqhxqhpBETUk9PVA", // EXT - Sourcegraph RFC drive
	FolderID:    "1zP3FxdDlcSQGC1qvM9lHZRaHH4I9Jwwa",
	OrderBy:     "createdTime,name",
}

var PrivateDrive = DriveSpec{
	DisplayName: "Private",
	DriveID:     "0AK4DcztHds_pUk9PVA", // Sourcegraph DriveID
	FolderID:    "1KCq4tMLnVlC0a1rwGuU5OSCw6mdDxLuv",
	OrderBy:     "createdTime,name",
}

type DriveSpec struct {
	DisplayName string
	DriveID     string
	FolderID    string
	OrderBy     string
}

const AuthEndpoint = "/oauth2/callback"

func (d *DriveSpec) Query(q string) string {
	return fmt.Sprintf("%s and parents in '%s'", q, d.FolderID)
}

// Retrieve a token, saves the token, then returns the generated client.
func getClient(ctx context.Context, config *oauth2.Config, out *std.Output) (*http.Client, error) {
	sec, err := secrets.FromContext(ctx)
	if err != nil {
		return nil, err
	}
	tok := &oauth2.Token{}
	if err := sec.Get("rfc", tok); err != nil {
		// ...if it doesn't exist, open browser and ask user to give us
		// permissions
		tok, err = getTokenFromWeb(handleAuthResponse, ctx, NewTokenHandler(config), out)
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

// allocateRandomPort ... allocates a random port
func allocateRandomPort() (net.Listener, error) {
	socket, err := net.Listen("tcp", ":0")
	if err != nil {
		return nil, errors.Wrap(err, "cannot allocate port for Google Authentication handler")
	}
	return socket, nil
}

// authResponseHandler returns a handler for the OAuth redirect response from Google.
// It sends the authentication code received from the redirect to the sendCode channel.
//
// sendCode: A channel to send the authentication code received from the redirect to.
// gracefulShutdown: Whether the server should shutdown gracefully after handling the request.
func authResponseHandler(sendCode chan string, sendError chan error, gracefulShutdown *bool) func(
	rw http.ResponseWriter, r *http.Request) {
	return func(rw http.ResponseWriter, r *http.Request) {
		authCode := r.URL.Query().Get("code")
		if authCode == "" {
			sendError <- errors.Errorf("Did not get authentication code from Google")
			return
		}
		rw.Header().Add("Content-Type", "text/plain")
		_, _ = rw.Write([]byte(`'sg' authentication complete. You may close this window.`))
		sendError <- nil
		sendCode <- authCode
		*gracefulShutdown = true
	}
}

// startAuthHandlerServer starts a local HTTP server to handle the OAuth redirect
// response from Google.
//
// socket: The listener for the server.
// authEndpoint: The endpoint which will handle the OAuth redirect response.
// sendCode: A channel to send the authentication code received from the redirect to.
// server: The HTTP server.
// gracefulShutdown: Whether the server shutdown gracefully after handling a request.
// handler: The request handler for the server, containing the authEndpoint.
func startAuthHandlerServer(socket net.Listener, authEndpoint string,
	codeReceiver chan string, errorReceiver chan error) {
	var server http.Server
	var gracefulShutdown bool = false

	// Creates a handler to handle response
	handler := http.NewServeMux()
	handler.Handle(authEndpoint,
		http.HandlerFunc(authResponseHandler(codeReceiver, errorReceiver,
			&gracefulShutdown)))

	server.Handler = handler

	go func() {
		defer socket.Close()
		if err := server.Serve(socket); err != nil {
			if !gracefulShutdown {
				log.Fatal("failure to handle", err)
			}
		}
	}()
}

// handleAuthResponse sets up a local HTTP server to handle the OAuth redirect
// response from Google. It returns the redirect URL to provide to Google, and a
// channel which will receive the authentication code from the redirect.
//
// sendCode: A channel which will receive the authentication code from the redirect.
// socket: A listener for the local HTTP server.
// redirectUrl: The URL to provide to Google for the OAuth redirect.
// err: Any error encountered setting up the server.
func handleAuthResponse() (*url.URL, chan string, chan error, error) {
	codeReceiver := make(chan string, 1)
	errorReceiver := make(chan error, 1)

	socket, err := allocateRandomPort()
	if err != nil {
		return nil, nil, nil, err
	}

	startAuthHandlerServer(socket, AuthEndpoint, codeReceiver, errorReceiver)

	redirectUrl := url.URL{
		Host:   net.JoinHostPort("localhost", strconv.Itoa(socket.Addr().(*net.TCPAddr).Port)),
		Path:   AuthEndpoint,
		Scheme: "http",
	}

	return &redirectUrl, codeReceiver, errorReceiver, nil
}

type authResponseHandlerFactory func() (*url.URL, chan string, chan error, error)

// tokenHandler implements a minimal surface required to retrieve a token.
//
// It wraps the OAuth2 token acquisition, so we can mock it and
// test it without hitting Google servers.
type tokenHandler interface {
	AuthCodeURL(state string, opts ...oauth2.AuthCodeOption) string
	Exchange(ctx context.Context, code string,
		opts ...oauth2.AuthCodeOption) (*oauth2.Token, error)
	SetRedirectURL(*url.URL)
}

type tokenHandlerImpl struct {
	config *oauth2.Config
}

func (th *tokenHandlerImpl) SetRedirectURL(url *url.URL) {
	th.config.RedirectURL = url.String()
}

func (th *tokenHandlerImpl) AuthCodeURL(state string, opts ...oauth2.AuthCodeOption) string {
	return th.config.AuthCodeURL(state, opts...)
}

func (th *tokenHandlerImpl) Exchange(ctx context.Context, code string,
	opts ...oauth2.AuthCodeOption) (*oauth2.Token, error) {
	return th.config.Exchange(ctx, code, opts...)
}

func NewTokenHandler(config *oauth2.Config) *tokenHandlerImpl {
	return &tokenHandlerImpl{
		config: config,
	}
}

// Request a token from the web, then returns the retrieved token.
func getTokenFromWeb(responseFactory authResponseHandlerFactory, ctx context.Context,
	config tokenHandler, out *std.Output) (*oauth2.Token, error) {
	out.WriteNoticef("Setting up Google token via oAuth - follow the prompts to get set up!")

	var err error

	var redirectUrl *url.URL
	var waitForCode chan string
	var waitForError chan error

	if redirectUrl, waitForCode, waitForError, err = responseFactory(); err == nil {
		config.SetRedirectURL(redirectUrl)
	} else {
		log.Fatal("Cannot create a response handler for Google authorization: ", err)
	}

	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)

	out.Writef("Opening %s ...", authURL)
	if err := open.URL(authURL); err != nil {
		return nil, err
	}

	out.WriteWarningf(
		" Your action is required:\n" +
			"   1. Your computer may ask to receive incoming connections.\n" +
			"      Please allow so the browser and sg can communicate.\n" +
			"   2. Please accept the browser access request.\n\n" +
			"   This process will resume automatically.")

	authError := <-waitForError
	if authError != nil {
		return nil, authError
	}

	authCode := <-waitForCode
	out.WriteSuccessf("Received confirmation. Continuing.")

	return config.Exchange(ctx, authCode)
}

func queryRFCs(ctx context.Context, query string, driveSpec DriveSpec, pager func(r *drive.FileList) error, out *std.Output) error {
	// If modifying these scopes, delete your previously saved token.json.
	sec, err := secrets.FromContext(ctx)
	if err != nil {
		return err
	}
	clientCredentials, err := sec.GetExternal(ctx, secrets.ExternalSecret{
		Project: "sourcegraph-local-dev",
		// sg Google client credentials
		Name: "SG_GOOGLE_CREDS",
	})
	if err != nil {
		return errors.Wrap(err, "failed to get google client credentials")
	}

	config, err := google.ConfigFromJSON([]byte(clientCredentials), drive.DriveMetadataReadonlyScope)
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
	q := driveSpec.Query(query)

	list := srv.Files.List().
		Corpora("drive").SupportsAllDrives(true).
		DriveId(driveSpec.DriveID).
		IncludeItemsFromAllDrives(true).
		SupportsAllDrives(true).
		PageSize(100).
		Q(q).
		Fields("nextPageToken, files(id, name, parents, description, modifiedTime)")

	if driveSpec.OrderBy != "" {
		list = list.OrderBy(driveSpec.OrderBy)
	}

	return list.Pages(ctx, pager)
}

func List(ctx context.Context, driveSpec DriveSpec, out *std.Output) error {
	return queryRFCs(ctx, "", driveSpec, rfcTitlesPrinter(out), out)
}

func Search(ctx context.Context, query string, driveSpec DriveSpec, out *std.Output) error {
	return queryRFCs(ctx, fmt.Sprintf("(name contains '%[1]s' or fullText contains '%[1]s')", query), driveSpec, rfcTitlesPrinter(out), out)
}

func Open(ctx context.Context, number string, driveSpec DriveSpec, out *std.Output) error {
	return queryRFCs(ctx, fmt.Sprintf("name contains 'RFC %s'", number), driveSpec, func(r *drive.FileList) error {
		for _, f := range r.Files {
			if err := open.URL(fmt.Sprintf("https://docs.google.com/document/d/%s/edit", f.Id)); err != nil {
				out.WriteFailuref("failed to open browser ", err)
			}
		}
		return nil
	}, out)
}

// RFCs should have the following format:
//
//	RFC 123: WIP: Foobar
//	    ^^^  ^^^  ^^^^^^
//	     |    |       |
//	     | matches[2] |
//	 matches[1]     matches[3]
//
// Variations supported:
//
//	RFC 123 WIP: Foobar
//	RFC 123 PRIVATE WIP: Foobar
var rfcTitleRegex = regexp.MustCompile(`RFC\s(\d+):*\s([\w\s]+):\s(.*)$`)

func rfcTitlesPrinter(out *std.Output) func(r *drive.FileList) error {
	return func(r *drive.FileList) error {
		if len(r.Files) == 0 {
			return nil
		}

		for _, f := range r.Files {
			modified, err := time.Parse("2006-01-02T15:04:05.000Z", f.ModifiedTime)
			if err != nil {
				// if this errors then we are handling the Google API wrong, return an error
				return errors.Wrap(err, "ModifiedTime")
			}

			matches := rfcTitleRegex.FindStringSubmatch(f.Name)
			if len(matches) == 4 {
				number := matches[1]
				statuses := strings.Split(strings.ToUpper(matches[2]), " ")
				name := matches[3]

				var statusColor output.Style = output.StyleItalic
				for _, s := range statuses {
					switch strings.ToUpper(s) {
					case "WIP":
						statusColor = output.StylePending
					case "REVIEW":
						statusColor = output.Fg256Color(208)
					case "IMPLEMENTED", "APPROVED", "DONE":
						statusColor = output.StyleSuccess
					case "ABANDONED", "PAUSED":
						statusColor = output.StyleSearchAlertTitle
					}
				}

				// Modifiers should combine existing styles, applied after the first iteration
				for _, s := range statuses {
					switch strings.ToUpper(s) {
					case "PRIVATE":
						statusColor = output.CombineStyles(statusColor, output.StyleUnderline)
					}
				}

				numberColor := output.Fg256Color(8)

				out.Writef("RFC %s%s %s%s%s %s %s%s %s%s",
					numberColor, number,
					statusColor, strings.Join(statuses, " "),
					output.StyleReset, name,
					output.StyleSuggestion, modified.Format("2006-01-02"), f.Description,
					output.StyleReset)
			} else {
				out.Writef("%s%s", f.Name, output.StyleReset)
			}
		}

		return nil
	}

}
