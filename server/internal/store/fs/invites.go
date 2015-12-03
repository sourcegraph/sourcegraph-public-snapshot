package fs

import (
	"encoding/json"
	"errors"
	"log"
	"os"
	"time"

	"golang.org/x/net/context"
	"sourcegraph.com/sourcegraph/rwvfs"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/store"
	"src.sourcegraph.com/sourcegraph/util/randstring"
)

const invitesDBFilename = "authorization_codes.json"

// dbInvites stores account invites and metadata.
type dbInvites struct {
	Email     string
	Token     string
	Write     bool
	Admin     bool
	InUse     bool
	CreatedAt time.Time
}

func toInvite(d *dbInvites) *sourcegraph.AccountInvite {
	return &sourcegraph.AccountInvite{
		Email: d.Email,
		Write: d.Write,
		Admin: d.Admin,
	}
}

// readInvitesDB reads the invite database from disk. If no such
// file exists, an empty slice is returned (and no error).
func readInvitesDB(ctx context.Context) ([]*dbInvites, error) {
	f, err := dbVFS(ctx).Open(invitesDBFilename)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	defer f.Close()

	var invitesList []*dbInvites
	if err := json.NewDecoder(f).Decode(&invitesList); err != nil {
		return nil, err
	}
	return invitesList, nil
}

// writeInvitesDB writes the regClient/account database to disk.
func writeInvitesDB(ctx context.Context, invitesList []*dbInvites) (err error) {
	data, err := json.MarshalIndent(invitesList, "", "  ")
	if err != nil {
		return err
	}

	if err := rwvfs.MkdirAll(dbVFS(ctx), "."); err != nil {
		return err
	}
	f, err := dbVFS(ctx).Create(invitesDBFilename)
	if err != nil {
		return err
	}
	defer func() {
		if err2 := f.Close(); err2 != nil {
			if err == nil {
				err = err2
			} else {
				log.Printf("Warning: closing auth codes DB after error (%s) failed: %s.", err, err2)
			}
		}
	}()

	_, err = f.Write(data)
	return err
}

// Invites is a FS-backed implementation of the Invites store.
type Invites struct{}

var _ store.Invites = (*Invites)(nil)

func (s *Invites) CreateOrUpdate(ctx context.Context, invite *sourcegraph.AccountInvite) (string, error) {
	invitesList, err := readInvitesDB(ctx)
	if err != nil {
		return "", err
	}

	newInvite := &dbInvites{
		Email:     invite.Email,
		Token:     randstring.NewLen(20),
		Write:     invite.Write,
		Admin:     invite.Admin,
		CreatedAt: time.Now(),
	}

	var update bool
	for i := range invitesList {
		if invite.Email == invitesList[i].Email {
			invitesList[i] = newInvite
			update = true
			break
		}
	}
	if !update {
		invitesList = append(invitesList, newInvite)
	}

	// Save to disk.
	if err := writeInvitesDB(ctx, invitesList); err != nil {
		return "", err
	}

	return newInvite.Token, nil
}

func (s *Invites) Retrieve(ctx context.Context, token string) (*sourcegraph.AccountInvite, error) {
	invitesList, err := readInvitesDB(ctx)
	if err != nil {
		return nil, err
	}

	for i := range invitesList {
		if token == invitesList[i].Token {
			if invitesList[i].InUse {
				return nil, errors.New("already used")
			}
			return toInvite(invitesList[i]), nil
		}
	}

	return nil, errors.New("not found")
}

func (s *Invites) MarkUnused(ctx context.Context, token string) error {
	invitesList, err := readInvitesDB(ctx)
	if err != nil {
		return err
	}

	for i := range invitesList {
		if token == invitesList[i].Token {
			invitesList[i].InUse = false

			// Save to disk.
			if err := writeInvitesDB(ctx, invitesList); err != nil {
				return err
			}
			return nil
		}
	}

	return errors.New("not found")
}

func (s *Invites) Delete(ctx context.Context, token string) error {
	invitesList, err := readInvitesDB(ctx)
	if err != nil {
		return err
	}

	for i := range invitesList {
		if token == invitesList[i].Token {
			invitesList[i] = invitesList[len(invitesList)-1]

			// Save to disk.
			if err := writeInvitesDB(ctx, invitesList[:len(invitesList)-1]); err != nil {
				return err
			}
			return nil
		}
	}

	return errors.New("not found")
}

func (s *Invites) List(ctx context.Context) ([]*sourcegraph.AccountInvite, error) {
	accountInvites := make([]*sourcegraph.AccountInvite, 0)
	invitesList, err := readInvitesDB(ctx)
	if err != nil {
		return accountInvites, err
	}

	for _, invite := range invitesList {
		accountInvites = append(accountInvites, toInvite(invite))
	}
	return accountInvites, nil
}
