package fs

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"golang.org/x/net/context"
	"sourcegraph.com/sourcegraph/rwvfs"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/store"
)

const regClientDBFilename = "registered-clients.json"

// readRegClientDB reads the regClient/account database from disk. If no such
// file exists, an empty slice is returned (and no error).
func readRegClientDB(ctx context.Context) ([]*sourcegraph.RegisteredClient, error) {
	f, err := dbVFS(ctx).Open(regClientDBFilename)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	defer f.Close()

	var regClients []*sourcegraph.RegisteredClient
	if err := json.NewDecoder(f).Decode(&regClients); err != nil {
		return nil, err
	}
	return regClients, nil
}

// writeRegClientDB writes the regClient/account database to disk.
func writeRegClientDB(ctx context.Context, regClients []*sourcegraph.RegisteredClient) (err error) {
	data, err := json.MarshalIndent(regClients, "", "  ")
	if err != nil {
		return err
	}

	if err := rwvfs.MkdirAll(dbVFS(ctx), "."); err != nil {
		return err
	}
	f, err := dbVFS(ctx).Create(regClientDBFilename)
	if err != nil {
		return err
	}
	defer func() {
		if err2 := f.Close(); err2 != nil {
			if err == nil {
				err = err2
			} else {
				log.Printf("Warning: closing registered clients DB after error (%s) failed: %s.", err, err2)
			}
		}
	}()

	_, err = f.Write(data)
	return err
}

// registeredClients is a FS-backed implementation of the RegisteredClients store.
type registeredClients struct{}

var _ store.RegisteredClients = (*registeredClients)(nil)

func (s *registeredClients) Get(ctx context.Context, client sourcegraph.RegisteredClientSpec) (*sourcegraph.RegisteredClient, error) {
	clients, err := readRegClientDB(ctx)
	if err != nil {
		return nil, err
	}

	for _, c := range clients {
		if client.ID == c.ID {
			c.ClientSecret = "" // avoid leaking
			return c, nil
		}
	}

	return nil, &store.RegisteredClientNotFoundError{ID: client.ID}
}

func (s *registeredClients) GetByCredentials(ctx context.Context, cred sourcegraph.RegisteredClientCredentials) (*sourcegraph.RegisteredClient, error) {
	clients, err := readRegClientDB(ctx)
	if err != nil {
		return nil, err
	}

	cred.Secret = hashSecret(cred.Secret)
	for _, c := range clients {
		if cred.ID == c.ID && cred.Secret == c.ClientSecret {
			c.ClientSecret = "" // avoid leaking
			return c, nil
		}
	}

	return nil, &store.RegisteredClientNotFoundError{ID: cred.ID, Secret: cred.Secret}
}

func (s *registeredClients) Create(ctx context.Context, client sourcegraph.RegisteredClient) error {
	if client.ID == "" {
		return fmt.Errorf("registered client ID must be set")
	}
	if client.ClientSecret == "" && client.JWKS == "" {
		return fmt.Errorf("registered client secret or JWKS must be set")
	}

	clients, err := readRegClientDB(ctx)
	if err != nil {
		return err
	}

	// Check ID uniqueness.
	for _, c := range clients {
		if client.ID == c.ID {
			return store.ErrRegisteredClientIDExists
		}
	}

	if client.ClientSecret != "" {
		client.ClientSecret = hashSecret(client.ClientSecret)
	}

	clients = append(clients, &client)
	return writeRegClientDB(ctx, clients)
}

func (s *registeredClients) Update(ctx context.Context, client sourcegraph.RegisteredClient) error {
	if client.ID == "" {
		return fmt.Errorf("registered client ID must be set")
	}
	if client.ClientSecret != "" {
		return fmt.Errorf("registered client secret must not be set")
	}

	clients, err := readRegClientDB(ctx)
	if err != nil {
		return err
	}

	for i, c := range clients {
		if client.ID == c.ID {
			clients[i] = &client
			return writeRegClientDB(ctx, clients)
		}
	}

	return &store.RegisteredClientNotFoundError{ID: client.ID}
}

func (s *registeredClients) Delete(ctx context.Context, client sourcegraph.RegisteredClientSpec) error {
	clients, err := readRegClientDB(ctx)
	if err != nil {
		return err
	}

	var keep []*sourcegraph.RegisteredClient
	for _, c := range clients {
		if c.ID != client.ID {
			keep = append(keep, c)
		}
	}

	if len(keep) == len(clients) {
		return &store.RegisteredClientNotFoundError{ID: client.ID}
	}

	return writeRegClientDB(ctx, keep)
}

func (s *registeredClients) List(ctx context.Context, opt sourcegraph.RegisteredClientListOptions) (*sourcegraph.RegisteredClientList, error) {
	clients, err := readRegClientDB(ctx)
	if err != nil {
		return nil, err
	}

	return &sourcegraph.RegisteredClientList{
		Clients: clients,
		StreamResponse: sourcegraph.StreamResponse{
			HasMore: false,
		},
	}, nil
}

// hashSecret hashes the registered client secret value.
func hashSecret(secret string) string {
	h := sha256.Sum256([]byte(secret))
	return base64.StdEncoding.EncodeToString(h[:])
}
