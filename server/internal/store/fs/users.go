package fs

import (
	"encoding/json"
	"log"
	"os"

	"strings"

	"golang.org/x/net/context"
	"sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"sourcegraph.com/sourcegraph/rwvfs"
	"sourcegraph.com/sourcegraph/sourcegraph/store"
)

const userDBFilename = "users.json"

type userDBEntry struct {
	sourcegraph.User
	EmailAddrs []*sourcegraph.EmailAddr
}

func usersFromUserDBEntries(entries []*userDBEntry) []*sourcegraph.User {
	users := make([]*sourcegraph.User, len(entries))
	for i, e := range entries {
		users[i] = &e.User
	}
	return users
}

// readUserDB reads the user/account database from disk. If no such
// file exists, an empty slice is returned (and no error).
func readUserDB(ctx context.Context) ([]*userDBEntry, error) {
	f, err := dbVFS(ctx).Open(userDBFilename)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	defer f.Close()

	var users []*userDBEntry
	if err := json.NewDecoder(f).Decode(&users); err != nil {
		return nil, err
	}
	return users, nil
}

// writeUserDB writes the user/account database to disk.
func writeUserDB(ctx context.Context, users []*userDBEntry) (err error) {
	data, err := json.MarshalIndent(users, "", "  ")
	if err != nil {
		return err
	}

	if err := rwvfs.MkdirAll(dbVFS(ctx), "."); err != nil {
		return err
	}
	f, err := dbVFS(ctx).Create(userDBFilename)
	if err != nil {
		return err
	}
	defer func() {
		if err2 := f.Close(); err2 != nil {
			if err == nil {
				err = err2
			} else {
				log.Printf("Warning: closing user DB after error (%s) failed: %s.", err, err2)
			}
		}
	}()

	_, err = f.Write(data)
	return err
}

// Users is an FS-backed implementation of the Users store.
type Users struct{}

var _ store.Users = (*Users)(nil)

func (s *Users) Get(ctx context.Context, userSpec sourcegraph.UserSpec) (*sourcegraph.User, error) {
	e, err := s.getDBEntry(ctx, userSpec)
	if err != nil {
		return nil, err
	}
	return &e.User, nil
}

func (s *Users) getDBEntry(ctx context.Context, userSpec sourcegraph.UserSpec) (*userDBEntry, error) {
	users, err := readUserDB(ctx)
	if err != nil {
		return nil, err
	}

	for _, user := range users {
		match := (userSpec.UID == 0 || userSpec.UID == user.UID) && (userSpec.Login == "" || userSpec.Login == user.Login) && !(userSpec.UID == 0 && userSpec.Login == "")
		if match {
			return user, nil
		}
	}

	return nil, &store.UserNotFoundError{Login: userSpec.Login, UID: int(userSpec.UID)}
}

func (s *Users) List(ctx context.Context, opt *sourcegraph.UsersListOptions) ([]*sourcegraph.User, error) {
	entries, err := readUserDB(ctx)
	if err != nil {
		return nil, err
	}

	var users []*sourcegraph.User

	if opt != nil && opt.Query != "" {
		users = []*sourcegraph.User{} // non-nil sentinel value
		for _, e := range entries {
			if userMatchesQuery(&e.User, opt.Query) {
				users = append(users, &e.User)
			}
		}
	}

	if users == nil {
		users = usersFromUserDBEntries(entries)
	}

	// TODO(sqs): respect opt.

	low := opt.Offset()
	if low >= len(users) {
		return []*sourcegraph.User{}, nil
	}
	high := low + opt.Limit()
	if high > len(users) {
		high = len(users)
	}
	return users[low:high], nil
}

func userMatchesQuery(user *sourcegraph.User, query string) bool {
	return strings.HasPrefix(strings.ToLower(user.Login), strings.ToLower(query))
}

func (s *Users) ListEmails(ctx context.Context, user sourcegraph.UserSpec) ([]*sourcegraph.EmailAddr, error) {
	e, err := s.getDBEntry(ctx, user)
	if err != nil {
		return nil, err
	}
	return e.EmailAddrs, nil
}
