package ext

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"golang.org/x/net/context"
)

var accessTokenConfigFilename = filepath.Join("config", "access_tokens.json")

type TokenNotFoundError struct {
	msg string
}

func (e TokenNotFoundError) Error() string { return e.msg }

// AccessTokens contains methods for storing global access tokens needed
// to authenticate against external services.
type AccessTokens struct{}

type serviceTokens map[string]string // Key is a service name, value is its token.

func readAccessTokenFile(ctx context.Context) (serviceTokens, error) {
	tokenFile := filepath.Join(os.Getenv("SGPATH"), accessTokenConfigFilename)
	f, err := os.Open(tokenFile)
	if err != nil {
		if os.IsNotExist(err) {
			return serviceTokens{}, nil
		}
		return nil, err
	}
	defer f.Close()

	var tokens serviceTokens
	if err := json.NewDecoder(f).Decode(&tokens); err != nil {
		return nil, err
	}
	return tokens, nil
}

func writeAccessTokenFile(ctx context.Context, tokens serviceTokens) error {
	data, err := json.MarshalIndent(tokens, "", "  ")
	if err != nil {
		return err
	}

	tokenFile := filepath.Join(os.Getenv("SGPATH"), accessTokenConfigFilename)
	dir, _ := filepath.Split(tokenFile)
	err = os.MkdirAll(dir, 0700)
	if err != nil {
		return err
	}
	f, err := os.Create(tokenFile)
	if err != nil {
		return err
	}
	defer f.Close()
	if err := f.Chmod(0600); err != nil {
		return err
	}

	_, err = f.Write(data)
	return err
}

func (s *AccessTokens) Set(ctx context.Context, host, token string) error {
	tokens, err := readAccessTokenFile(ctx)
	if err != nil {
		return err
	}

	tokens[host] = token
	return writeAccessTokenFile(ctx, tokens)
}

func (s *AccessTokens) Get(ctx context.Context, host string) (string, error) {
	tokens, err := readAccessTokenFile(ctx)
	if err != nil {
		return "", err
	}

	token, present := tokens[host]
	if !present {
		return "", TokenNotFoundError{msg: fmt.Sprintf("no token found for %s", host)}
	}
	return token, nil
}
