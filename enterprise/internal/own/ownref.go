package own

import (
	"bytes"
	"context"
	"fmt"
	"net/mail"
	"strings"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/auth/providers"
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
// TODO(#52141): Search by code host handle.
// TODO(#52246): ByTextReference uses fewer queries.
func ByTextReference(ctx context.Context, db edb.EnterpriseDB, text ...string) (Bag, error) {
	var multiBag bag
	for _, t := range text {
		t = strings.TrimPrefix(t, "@")
		if _, err := mail.ParseAddress(t); err == nil {
			b, err := ByEmailReference(ctx, db, t)
			if err != nil {
				return nil, err
			}
			multiBag = append(multiBag, b...)
		}
		b, err := ByUserHandleReference(ctx, db, t)
		if err != nil {
			return nil, err
		}
		multiBag = append(multiBag, b...)
	}
	return multiBag, nil
}

func ByUserHandleReference(ctx context.Context, db edb.EnterpriseDB, handle string) (bag, error) {
	var b bag
	// Try to find a user by username.
	user, err := db.Users().GetByUsername(ctx, handle)
	if err != nil && !errcode.IsNotFound(err) {
		return nil, errors.Wrap(err, "Users.GetByUsername")
	}
	if user != nil {
		b = append(b, &userReferences{
			id:   user.ID,
			user: user,
		})
		b, err = hydrateWithCodeHostHandles(ctx, user.ID, db, b)
		if err != nil {
			return nil, err
		}
	} else {
		b = append(b, &userReferences{
			handle: handle,
		})
	}
	return hydrateWithVerifiedEmails(ctx, db, b)
}

func ByEmailReference(ctx context.Context, db edb.EnterpriseDB, email string) (bag, error) {
	var b bag
	// Checking that provided email is verified.
	verifiedEmails, err := db.UserEmails().GetVerifiedEmails(ctx, email)
	if err != nil {
		return nil, err
	}
	// Email is not verified, including an input email as is and returning the bag.
	if len(verifiedEmails) != 1 {
		b = append(b, &userReferences{
			verifiedEmails: []string{email},
		})
		return b, nil
	}
	verifiedEmail := verifiedEmails[0]
	user, err := db.Users().GetByID(ctx, verifiedEmail.UserID)
	if err != nil && !errcode.IsNotFound(err) {
		return nil, errors.Wrap(err, "Users.GetByID")
	}
	if user != nil {
		// Not adding an email here, because we will add it in hydrateWithVerifiedEmails.
		b = append(b, &userReferences{
			id:   user.ID,
			user: user,
		})
		b, err = hydrateWithCodeHostHandles(ctx, user.ID, db, b)
		if err != nil {
			return b, err
		}
	} else {
		// In fact, user emails are deleted as soon as a user is soft/hard deleted, we
		// shouldn't be here, but in some weird race condition, we still populate the ID
		// and verified emails.
		b = append(b, &userReferences{
			id:             verifiedEmail.UserID,
			verifiedEmails: []string{verifiedEmail.Email},
		})
	}
	return hydrateWithVerifiedEmails(ctx, db, b)
}

func hydrateWithVerifiedEmails(ctx context.Context, db edb.EnterpriseDB, b bag) (bag, error) {
	for _, userRefs := range b {
		if userRefs.id == 0 {
			continue
		}
		// Getting verified emails of this user.
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

func hydrateWithCodeHostHandles(ctx context.Context, userID int32, db edb.EnterpriseDB, b bag) (bag, error) {
	accounts, err := db.UserExternalAccounts().List(ctx, database.ExternalAccountsListOptions{UserID: userID})
	if err != nil {
		return nil, errors.Wrap(err, "UserExternalAccounts.List")
	}
	codeHostHandles := make([]string, 0, len(accounts))
	for _, account := range accounts {
		p := providers.GetProviderbyServiceType(account.ServiceType)
		if p == nil {
			return nil, errors.Errorf("cannot find authorization provider for the external account, service type: %s", account.ServiceType)
		}
		data, err := p.ExternalAccountInfo(ctx, *account)
		if err != nil || data == nil {
			return nil, errors.Wrap(err, "ExternalAccountInfo")
		}
		if data.Login != nil && len(*data.Login) > 0 {
			codeHostHandles = append(codeHostHandles, *data.Login)
		}
	}
	// Finding the user reference with given userID and adding code host handles to that.
	for _, bg := range b {
		if bg.id == userID {
			bg.codeHostHandles = codeHostHandles
			break
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
	handle string
	// codeHostHandles are user handles of connected code hosts of a resolved user.
	codeHostHandles []string
	verifiedEmails  []string
}

func (refs userReferences) containsEmail(email string) bool {
	for _, vEmail := range refs.verifiedEmails {
		if vEmail == email {
			return true
		}
	}
	return false
}

func (refs userReferences) containsHandle(handle string) bool {
	handle = strings.TrimPrefix(handle, "@")
	if u := refs.user; u != nil && u.Username == handle {
		return true
	}
	if refs.handle != "" && refs.handle == handle {
		return true
	}
	for _, codeHostHandle := range refs.codeHostHandles {
		if handle == codeHostHandle {
			return true
		}
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
