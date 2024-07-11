package release

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/pkg/browser"
	"github.com/urfave/cli/v2"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
	"github.com/sourcegraph/sourcegraph/internal/jsonc"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/output"
)

type Event struct {
	Name               string      `json:"name"`
	MonthlyReleaseDate time.Time   `json:"monthlyReleaseDate"`
	PatchReleaseDates  []time.Time `json:"patchReleaseDates"`
}

type CalendarConfig struct {
	TeamEmail string  `json:"teamEmail"`
	Events    []Event `json:"events"`
}

func generateCalendarEvents(cctx *cli.Context) error {
	// config path: ${SOURCEGRAPH_REPO_ROOT}/tools/release/config/calendar.jsonc
	configPath := cctx.String("config")
	if configPath == "" {
		return errors.New("config path is required")
	}

	f, err := os.ReadFile(configPath)
	if err != nil {
		return err
	}

	var cc CalendarConfig
	if err := jsonc.Unmarshal(string(f), &cc); err != nil {
		return err
	}

	client, err := getGoogleCalClient(cctx.Context)
	if err != nil {
		return err
	}

	now := time.Now()
	for _, e := range cc.Events {
		p := std.Out.Pending(output.Styledf(output.StylePending, "Processing Calendar event: %q", e.Name))

		patchEventsToCreate := make([]time.Time, len(e.PatchReleaseDates))

		if e.MonthlyReleaseDate.Before(now) {
			p.Complete(output.Linef(output.EmojiWarning, output.StyleWarning, "Skipping event: %q because the monthly release date is in the past.", e.Name))
			continue
		}

		for i, patchReleaseDate := range e.PatchReleaseDates {
			if patchReleaseDate.Before(now) {
				continue
			}
			patchEventsToCreate[i] = patchReleaseDate
		}

		if len(patchEventsToCreate) == 0 {
			p.Complete(output.Linef(output.EmojiWarning, output.StyleWarning, "Skipping event: %q because there are no patch release dates for the month.", e.Name))
			continue
		}

		p.Updatef("Creating Monthly Release event for %q", e.Name)
		monthlyReleaseEvt := createReleaseEvent(cc.TeamEmail, fmt.Sprintf("Monthly Release: (%s)", e.Name), e.MonthlyReleaseDate)
		_, err = client.Events.Insert("primary", monthlyReleaseEvt).Context(cctx.Context).Do()
		if err != nil {
			p.Destroy()
			return errors.Wrapf(err, "Failed to create monthly release event for %q", e.Name)
		}

		p.Updatef("Creating Patch Release events for %q", e.Name)
		for _, prd := range patchEventsToCreate {
			patchReleaseEvt := createReleaseEvent(cc.TeamEmail, fmt.Sprintf("Patch Release: (%s)", e.Name), prd)
			_, err = client.Events.Insert("primary", patchReleaseEvt).Context(cctx.Context).Do()
			if err != nil {
				p.Destroy()
				return errors.Wrapf(err, "Failed to create patch release event for %q. Patch release date: %s", e.Name, prd)
			}
		}

		p.Complete(output.Linef(output.EmojiSuccess, output.StyleSuccess, "Sent and created event: %q", e.Name))
	}

	return nil
}

func createReleaseEvent(email, title string, eventTime time.Time) *calendar.Event {
	startTime := eventTime.Format("2006-01-02T15:04:05.000-07:00")
	endTime := eventTime.Add(time.Minute).Format("2006-01-02T15:04:05.000-07:00")
	return &calendar.Event{
		AnyoneCanAddSelf: true,
		Summary:          title,
		Description:      "(This is not an actual event to attend, just a calendar marker.)",
		Location:         "Sourcegraph (Remote)",
		Start: &calendar.EventDateTime{
			DateTime: startTime,
			TimeZone: "America/Los_Angeles",
		},
		End: &calendar.EventDateTime{
			DateTime: endTime,
			TimeZone: "America/Los_Angeles",
		},
		Attendees: []*calendar.EventAttendee{
			{Email: email, Optional: true},
		},
	}
}

type GoogleCalendarCred struct {
	Installed struct {
		ClientID     string   `json:"client_id"`
		ProjectID    string   `json:"project_id"`
		AuthURI      string   `json:"auth_uri"`
		TokenURI     string   `json:"token_uri"`
		AuthProvider string   `json:"auth_provider_x509_cert_url"`
		ClientSecret string   `json:"client_secret"`
		RedirectURIs []string `json:"redirect_uris"`
	}
}

func getGoogleCalClient(ctx context.Context) (*calendar.Service, error) {
	credentialsJSON, err := std.Out.
		PromptPasswordf(
			os.Stdin,
			`Paste Google Calendar credentials (1Password "Release automation Google Calendar API App credentials"): `,
		)
	if err != nil {
		return nil, err
	}

	var creds GoogleCalendarCred
	if err := json.Unmarshal([]byte(credentialsJSON), &creds); err != nil {
		return nil, err
	}

	srv, err := createServer()
	if err != nil {
		return nil, err
	}
	defer srv.Close()

	cfg := oauth2.Config{
		ClientID:     creds.Installed.ClientID,
		ClientSecret: creds.Installed.ClientSecret,
		RedirectURL:  fmt.Sprintf("http://localhost%s", srv.Addr), // CI:LOCALHOST_OK
		Endpoint:     google.Endpoint,
		Scopes:       []string{"https://www.googleapis.com/auth/calendar.events"},
	}
	authUrl := cfg.AuthCodeURL("state", oauth2.AccessTypeOffline)

	if err := browser.OpenURL(authUrl); err != nil {
		return nil, err
	}

	// Wait for the authCode from the redirect
	authCode, err := waitForAuthCode()
	if err != nil {
		return nil, err
	}

	token, err := cfg.Exchange(ctx, authCode)
	if err != nil {
		return nil, err
	}

	oauth2Client := cfg.Client(ctx, token)
	return calendar.NewService(ctx, option.WithHTTPClient(oauth2Client))
}

// authCodeChan is used to communicate the auth code from the server handler to the waitForAuthCode function
var authCodeChan = make(chan string)

func createServer() (*http.Server, error) {
	// Listen on a random port
	listener, err := net.Listen("tcp", "")
	if err != nil {
		return nil, err
	}

	port := listener.Addr().(*net.TCPAddr).Port
	server := &http.Server{
		Addr: fmt.Sprintf(":%d", port),
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/plain")

			q := r.URL.Query()
			code := q.Get("code")
			err := q.Get("error")

			if err != "" {
				fmt.Fprintf(w, "Authentication failed: %s", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			authCodeChan <- code
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, "Authentication successful! Please return to the console")
		}),
	}

	go func() {
		defer listener.Close()
		if err := server.Serve(listener); err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	return server, nil
}

func waitForAuthCode() (string, error) {
	select {
	case authCode := <-authCodeChan:
		return authCode, nil
	case <-time.After(5 * time.Minute):
		return "", errors.Newf("timeout waiting for auth code")
	}
}
