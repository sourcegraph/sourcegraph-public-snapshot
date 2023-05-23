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
	"github.com/sourcegraph/sourcegraph/enterprise/internal/own/codeowners"
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
	Contains(Reference) bool

	Add(Reference)

	Resolve(context.Context, edb.EnterpriseDB) error

	Get(Reference) codeowners.ResolvedOwner
}

func EmptyBag() Bag {
	return &bag{}
}

func (b *bag) Add(ref Reference) {
	manyResolved := b.find(ref)
	if len(manyResolved) == 0 {
		*b = append(*b, &userReferences{unresolvedReferences: []Reference{ref}})
		return
	}
	var oneResolved *userReferences
	if len(manyResolved) == 1 {
		oneResolved = manyResolved[0]
	} else {
		// TODO: Unification case where we learn many references are one
		// For now we just pick the first one.
		oneResolved = manyResolved[0]
	}
	oneResolved.add(ref)
}

func (b *bag) Resolve(ctx context.Context, db edb.EnterpriseDB) error {
	// TODO: We should keep track of which user references were resolved
	// For now we try to resolve all unresolved references.
	for _, userRef := range *b {
		if userRef.id != 0 {
			// TODO: Probably in that case try to resolve references to see
			// if they don't point to some other user?
			continue
		}
		rs := userRef.unresolvedReferences
		for _, r := range rs {
			if r.Email != "" {
				bb, err := ByEmailReference(ctx, db, r.Email)
				if err != nil {
					return err
				}
				// TODO: Properly handle what if search yields more than one userRef.
				if len(bb) > 0 {
					// TODO: Properly merge data rather than just override
					*userRef = *bb[0]
				}
			}
			if r.Handle != "" {
				bb, err := ByUserHandleReference(ctx, db, r.Handle)
				if err != nil {
					return err
				}
				// TODO: Properly handle what if search yields more than one userRef.
				if len(bb) > 0 {
					// TODO: Properly merge data rather than just override
					*userRef = *bb[0]
				}
			}
			if r.UserID != 0 {
				userRef.id = r.UserID
				var err error
				userRef.user, err = db.Users().GetByID(ctx, r.UserID)
				if err != nil {
					return err
				}
			}
			// TODO if we're here, either user is still not resolved
			// or the user ref is updated. Anywho we can iterate further.
		}
		userRef.unresolvedReferences = nil
	}
	return nil
}

func (b *bag) Get(ref Reference) codeowners.ResolvedOwner {
	refs := b.find(ref)
	if len(refs) > 0 {
		// TODO: The object should be crafted so that only one ref
		// is returned here.
		r := refs[0]
		var primaryEmail *string
		// TODO: Assumption primary email is the first verified
		if len(r.verifiedEmails) > 0 {
			e := r.verifiedEmails[0]
			primaryEmail = &e
		}
		var handle string
		if r.user != nil {
			handle = r.user.Username
		} else if r.handle != "" {
			handle = r.handle
		} else {
			handle = ref.Handle
		}
		var email string
		if primaryEmail != nil {
			email = *primaryEmail
		} else {
			email = ref.Email
		}
		return &codeowners.Person{
			User:         r.user,
			PrimaryEmail: primaryEmail,
			Handle:       handle,
			Email:        email,
		}
	}
	// TODO: We need to be able to return an owner with stable identifier here
	return &codeowners.Person{
		Handle: ref.Handle,
		Email:  ref.Email,
	}
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
		if _, err := mail.ParseAddress(t); err == nil {
			b, err := ByEmailReference(ctx, db, t)
			if err != nil {
				return nil, err
			}
			multiBag = append(multiBag, b...)
		}
		b, err := ByUserHandleReference(ctx, db, strings.TrimPrefix(t, "@"))
		if err != nil {
			return nil, err
		}
		multiBag = append(multiBag, b...)
	}
	return &multiBag, nil
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
	codeHostHandles      []string
	verifiedEmails       []string
	unresolvedReferences []Reference
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

func (refs userReferences) add(ref Reference) {
	if refs.user != nil {
		// Merge in identifiers from given reference if any missing
		// For now do nothing
		return
	}
	refs.unresolvedReferences = append(refs.unresolvedReferences, ref)
}

func (refs userReferences) String() string {
	var buf bytes.Buffer
	fmt.Fprint(&buf, "{")
	var needsComma bool
	comma := func() {
		if needsComma {
			fmt.Fprint(&buf, ", ")
		}
		needsComma = true
	}
	if refs.id != 0 {
		comma()
		fmt.Fprintf(&buf, "#%d", refs.id)
	}
	if refs.user != nil {
		comma()
		fmt.Fprint(&buf, refs.user.Username)
	}
	if refs.handle != "" {
		comma()
		fmt.Fprintf(&buf, "@%s", refs.handle)
	}
	for _, e := range refs.verifiedEmails {
		comma()
		fmt.Fprint(&buf, e)
	}
	fmt.Fprint(&buf, "}")
	return buf.String()
}

// Contains at this point returns true
//   - if email reference matches any user's verified email,
//   - if handle reference matches the user handle,
//   - if user ID matches the ID if the user in the bag,
func (b *bag) Contains(ref Reference) bool {
	return len(b.find(ref)) > 0
}

func (b bag) find(ref Reference) []*userReferences {
	var found []*userReferences
	for _, userRefs := range b {
		if userRefs.containsEmail(ref.Email) {
			found = append(found, userRefs)
			continue
		}
		if userRefs.containsHandle(ref.Handle) {
			found = append(found, userRefs)
			continue
		}
		if userRefs.containsUserID(ref.UserID) {
			found = append(found, userRefs)
			continue
		}
	}
	return found
}

func (b bag) String() string {
	var buf bytes.Buffer
	fmt.Fprint(&buf, "[")
	var needsComma bool
	comma := func() {
		if needsComma {
			fmt.Fprint(&buf, ", ")
		}
		needsComma = true
	}
	for _, r := range b {
		comma()
		fmt.Fprint(&buf, r.String())
	}
	fmt.Fprint(&buf, "]")
	return buf.String()
}
