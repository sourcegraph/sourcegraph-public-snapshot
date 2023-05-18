package own

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	"github.com/sourcegraph/sourcegraph/lib/errors"

	edb "github.com/sourcegraph/sourcegraph/enterprise/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/errcode"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

// RepoContext allows us to anchor an author reference to a repo where it stems from.
// For instance a handle from a CODEOWNERS file comes from github.com/sourcegraph/sourcegraph.
// This is important for resolving namespaced owner names
// (like CODEOWNERS file can refer to team handle "own"), while the name in the database is "sourcegraph/own"
// because it was pulled from github, and by convention organization name is prepended.
type RepoContext struct {
	Name         api.RepoName
	CodeHostKind string
}

// Reference is whatever we get from a data source, like a commit,
// CODEOWNERS entry or file view.
type Reference struct {
	// RepoContext is present if given owner reference is associated
	// with specific repository.
	RepoContext *RepoContext
	// UserID indicates identifying a specific user.
	UserID int32
	// Handle is either a sourcegraph or code-host handle,
	// and can be considered within or outside of the repo context.
	Handle string
	// Email can be found in a CODEOWNERS entry, but can also
	// be a commit author email, which means it can be a code-host specific
	// email generated for the purpose of merging a pull-request.
	Email string
}

func (r Reference) String() string {
	var b bytes.Buffer
	fmt.Fprint(&b, "{")
	var needsComma bool
	nextPart := func() {
		if needsComma {
			fmt.Fprint(&b, ", ")
		}
		needsComma = true
	}
	if r.UserID != 0 {
		nextPart()
		fmt.Fprintf(&b, "userID: %d", r.UserID)
	}
	if r.Handle != "" {
		nextPart()
		fmt.Fprintf(&b, "handle: %s", r.Handle)
	}
	if r.Email != "" {
		nextPart()
		fmt.Fprintf(&b, "email: %s", r.Email)
	}
	if c := r.RepoContext; c != nil {
		nextPart()
		fmt.Fprintf(&b, "context.%s: %s", c.CodeHostKind, c.Name)
	}
	fmt.Fprint(&b, "}")
	return b.String()
}

// Bag is a collection of platonic forms or identities of owners (teams, people
// or otherwise). It is pre-seeded with database information based on the search
// query using `ByTextReference`, and used to match found owner references.
type Bag interface {
	// Contains answers true if given bag contains an owner form
	// that the given reference points at in some way.
	Contains(ref Reference) bool
}

// ByTextReference returns a Bag of all the forms (users, persons, teams)
// that can be referred to by given text (name or email alike).
// This can be used in search to find relevant owners by different identifiers
// that the database reveals.
// TODO(#52140): Search by verified email.
// TODO(#52141): Search by code host handle.
func ByTextReference(ctx context.Context, db edb.EnterpriseDB, text string) (Bag, error) {
	text = strings.TrimPrefix(text, "@")
	var b bag
	// Try to find a user by username (the only lookup for now):
	user, err := db.Users().GetByUsername(ctx, text)
	if err != nil && !errcode.IsNotFound(err) {
		return nil, errors.Wrap(err, "Users.GetByUsername")
	}
	if user != nil {
		b = append(b, &userReferences{
			id:   user.ID,
			user: user,
		})
	} else {
		b = append(b, &userReferences{
			handle: text,
		})
	}
	for _, userRefs := range b {
		if userRefs.id == 0 {
			continue
		}
		verifiedEmails, err := db.UserEmails().ListByUser(ctx, database.UserEmailsListOptions{UserID: userRefs.id, OnlyVerified: true})
		if err != nil {
			return nil, errors.Wrap(err, "UserEmails.ListByUser")
		}
		for _, email := range verifiedEmails {
			userRefs.verifiedEmails = append(userRefs.verifiedEmails, email.Email)
		}
	}
	return b, nil
}

// bag is implemented as a slice of references for different users.
type bag []*userReferences

type userReferences struct {
	id   int32
	user *types.User
	// handle text input used to search that reference,
	// it is present if user was not found in the database.
	handle         string
	verifiedEmails []string
}

func (refs userReferences) containsEmail(email string) bool {
	for _, vEmail := range refs.verifiedEmails {
		if vEmail == email {
			return true
		}
	}
	return false
}

// TODO(#52142): Introduce matching on linked code host handles.
func (refs userReferences) containsHandle(handle string) bool {
	handle = strings.TrimPrefix(handle, "@")
	if u := refs.user; u != nil && u.Username == handle {
		return true
	}
	if refs.handle != "" && refs.handle == handle {
		return true
	}
	return false
}

func (refs userReferences) containsUserID(userID int32) bool {
	return refs.id != 0 && refs.id == userID
}

// Contains at this point returns true
//   - if email reference matches any user's verified email,
//   - if handle reference matches the user handle,
//   - if user ID matches the ID if the user in the bag,
func (b bag) Contains(ref Reference) bool {
	for _, userRefs := range b {
		if userRefs.containsEmail(ref.Email) {
			return true
		}
		if userRefs.containsHandle(ref.Handle) {
			return true
		}
		if userRefs.containsUserID(ref.UserID) {
			return true
		}
	}
	return false
}
