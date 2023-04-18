package store

import (
	"context"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/github_apps/types"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	encryption "github.com/sourcegraph/sourcegraph/internal/encryption"
	"github.com/sourcegraph/sourcegraph/internal/encryption/keyring"
)

// Store handles storing and retrieving GitHub Apps from the database.
type GithubAppsStore interface {
	// Create inserts a new GitHub App into the database.
	Create(ctx context.Context, app *types.GitHubApp) error

	// Delete removes a GitHub App from the database by ID.
	Delete(ctx context.Context, id int) error

	// Update updates a GitHub App in the database.
	Update(ctx context.Context, id int, app *types.GitHubApp) error

	// GetByID retrieves a GitHub App from the database by ID.
	GetByID(ctx context.Context, id int) (*types.GitHubApp, error)

	//  WithEncryptionKey sets encryption key on store. Returns a new GithubAppStore
	WithEncryptionKey(key encryption.Key) GithubAppsStore
}

// githubAppStore handles storing and retrieving GitHub Apps from the database.
type githubAppsStore struct {
	*basestore.Store

	key encryption.Key
}

func GithubAppsWith(other *basestore.Store) GithubAppsStore {
	return &githubAppsStore{
		Store: basestore.NewWithHandle(other.Handle()),
	}
}

// WithEncryptionKey sets encryption key on store. Returns a new GithubAppStore
func (s *githubAppsStore) WithEncryptionKey(key encryption.Key) GithubAppsStore {
	return &githubAppsStore{Store: s.Store, key: key}
}

func (s *githubAppsStore) getEncryptionKey() encryption.Key {
	if s.key != nil {
		return s.key
	}
	return keyring.Default().GithubAppKey
}

// Create inserts a new GitHub App into the database.
func (s *githubAppsStore) Create(ctx context.Context, app *types.GitHubApp) error {
	key := s.getEncryptionKey()
	clientSecret, keyID, err := encryption.MaybeEncrypt(ctx, key, app.ClientSecret)
	if err != nil {
		return err
	}
	privateKey, keyID, err := encryption.MaybeEncrypt(ctx, key, app.PrivateKey)
	if err != nil {
		return err
	}

	query := sqlf.Sprintf(`INSERT INTO
	    github_apps (app_id, name, slug, client_id, client_secret, private_key, encryption_key_id, logo)
    	VALUES (%s, %s, %s, %s, %s, %s, %s, %s)`,
		app.AppID, app.Name, app.Slug, app.ClientID, clientSecret, privateKey, keyID, app.Logo)
	return s.Exec(ctx, query)
}

// Delete removes a GitHub App from the database by ID.
func (s *githubAppsStore) Delete(ctx context.Context, id int) error {
	query := sqlf.Sprintf(`DELETE FROM github_apps WHERE id = %s`, id)
	return s.Exec(ctx, query)
}

// Update updates a GitHub App in the database.
func (s *githubAppsStore) Update(ctx context.Context, id int, app *types.GitHubApp) error {
	key := s.getEncryptionKey()
	clientSecret, keyID, err := encryption.MaybeEncrypt(ctx, key, app.ClientSecret)
	if err != nil {
		return err
	}
	privateKey, keyID, err := encryption.MaybeEncrypt(ctx, key, app.PrivateKey)
	if err != nil {
		return err
	}

	query := sqlf.Sprintf(`UPDATE github_apps
             SET name = %s, slug = %s, client_id = %s, client_secret = %s, private_key = %s, encryption_key_id = %s, logo = %s, updated_at = NOW()
             WHERE id = %s`, app.ID, app.Name, app.Slug, app.ClientID, clientSecret, privateKey, keyID, app.Logo)
	return s.Exec(ctx, query)
}

// GetByID retrieves a GitHub App from the database by ID.
func (s *githubAppsStore) GetByID(ctx context.Context, id int) (*types.GitHubApp, error) {
	var app types.GitHubApp
	var clientSecret, privateKey, keyID string

	query := sqlf.Sprintf(`SELECT
		id,
		name,
		slug,
		client_id,
		client_secret,
		private_key,
		encryption_key_id,
		logo,
		created_at,
		updated_at
	FROM github_apps
	WHERE id = %s`, id)
	err := s.QueryRow(ctx, query).Scan(
		&app.ID,
		&app.Name,
		&app.Slug,
		&app.ClientID,
		&clientSecret,
		&privateKey,
		&keyID,
		&app.Logo,
		&app.CreatedAt,
		&app.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	key := s.getEncryptionKey()
	app.ClientSecret, err = encryption.MaybeDecrypt(ctx, key, clientSecret, keyID)
	if err != nil {
		return nil, err
	}
	app.PrivateKey, err = encryption.MaybeDecrypt(ctx, key, privateKey, keyID)
	if err != nil {
		return nil, err
	}
	return &app, nil
}
