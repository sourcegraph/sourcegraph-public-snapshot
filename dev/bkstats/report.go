package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

// Request a token from the web, then returns the retrieved token.
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		log.Fatalf("Unable to read authorization code: %v", err)
	}

	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web: %v", err)
	}
	return tok
}

// Retrieves a token from a local file.
func tokenFromString(token string) (*oauth2.Token, error) {
	tok := &oauth2.Token{}
	err := json.NewDecoder(strings.NewReader(token)).Decode(tok)
	return tok, err
}

func getOAuthConfig(config string) (*oauth2.Config, error) {
	return google.ConfigFromJSON([]byte(config), "https://www.googleapis.com/auth/spreadsheets")
}

func pushReport(ctx context.Context, oauth2config string, token string, t time.Time, downtime time.Duration) error {
	// If modifying these scopes, delete your previously saved token.json.
	config, err := google.ConfigFromJSON([]byte(oauth2config), "https://www.googleapis.com/auth/spreadsheets")
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}
	tok, err := tokenFromString(token)
	if err != nil {
		return err
	}

	client := config.Client(ctx, tok)
	srv, err := sheets.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		log.Fatalf("Unable to retrieve Sheets client: %v", err)
	}
	call := srv.Spreadsheets.Values.Append("1hykYXEa2emn4zoM5sfS42bT6pZk24RChjzvx8Oog8g8", "Daily Red Time Data!A:B", &sheets.ValueRange{
		Values: [][]interface{}{{t.Format("01/02/2006"), downtime.Minutes()}},
	})
	call.ValueInputOption("USER_ENTERED")
	_, err = call.Do()
	return err
}
