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

func (r Reference) ResolutionGuess() *codeowners.Person {
	if r.Handle == "" && r.Email == "" {
		return nil
	}
	return &codeowners.Person{
		Handle: r.Handle,
		Email:  r.Email,
	}
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

// ReferenceOrder is to allow sorting references in tests for comparison.
// It is not meant to be used in production or to be efficient.
type ReferenceOrder []Reference

func (o *ReferenceOrder) Len() int           { return len(*o) }
func (o *ReferenceOrder) Less(i, j int) bool { return (*o)[i].String() < (*o)[j].String() }
func (o *ReferenceOrder) Swap(i, j int)      { (*o)[i], (*o)[j] = (*o)[j], (*o)[i] }

// Bag is a collection of platonic forms or identities of owners (currently supports
// only users - teams coming). The purpose of this object is to group references
// that refer to the same user, so that the user can be found by each of the references.
//
// It can either be created from test references using `ByTextReference`,
// or `EmptyBag` if no initial references are know for the unification use case.
type Bag interface {
	// Contains answers true if given bag contains any of the references
	// in given References object. This includes references that are in the bag
	// (via `Add` or `ByTextReference`) and were not resolved to a user.
	Contains(Reference) bool

	// FindResolved returns the owner that was resolved from a reference
	// that was added through Add or the ByTextReference constructor (and true flag).
	// In case the reference was not resolved to any user, the result
	// is nil, and false boolean is returned.
	FindResolved(Reference) (codeowners.ResolvedOwner, bool)

	// Add puts a Reference in the Bag so that later a user associated
	// with that Reference can be pulled from the database by calling Resolve.
	Add(Reference)

	// Resolve retrieves all users it can find that are associated
	// with references that are in the Bag by means of Add or ByTextReference
	// constructor.
	Resolve(context.Context, edb.EnterpriseDB) error
}

func EmptyBag() *bag {
	return &bag{
		resolvedUsers: map[int32]*userReferences{},
		references:    map[refKey]*refContext{},
	}
}

// ByTextReference returns a Bag of all the forms (users, persons, teams)
// that can be referred to by given text (name or email alike).
// This can be used in search to find relevant owners by different identifiers
// that the database reveals.
// TODO(#52141): Search by code host handle.
// TODO(#52246): ByTextReference uses fewer queries.
func ByTextReference(ctx context.Context, db edb.EnterpriseDB, text ...string) (Bag, error) {
	b := EmptyBag()
	for _, t := range text {
		// Empty text does not resolve at all.
		if t == "" {
			continue
		}
		if _, err := mail.ParseAddress(t); err == nil {
			b.add(refKey{email: t})
		} else {
			b.add(refKey{handle: strings.TrimPrefix(t, "@")})
		}
	}
	if err := b.Resolve(ctx, db); err != nil {
		return nil, err
	}
	return b, nil
}

// bag is implemented as a map of resolved users and map of references.
type bag struct {
	// resolvedUsers map from user id to `userReferences` which contain
	// all the references found in the database for a given user.
	// These references are linked back to the `references` via `resolve`
	// call.
	resolvedUsers map[int32]*userReferences
	// references map a user reference to a refContext which can be either:
	// - resolved to a user, in which case it has non-0 `resolvedUserID`,
	//   and an entry with that user id exists in `resolvedUsers`.
	// - unresolved which means that either resolution was not attempted,
	//   so `resolve` was not called after adding given reference,
	//   or no user was able to be resolved (indicated by `resolutionDone` being `true`).
	references map[refKey]*refContext
}

// Contains returns true if given reference can be found in the bag,
// irrespective of whether the reference was resolved or not.
// This means that any reference that was added or passed
// to the `ByTextReference` should be in the bag. Moreover,
// for every user that was resolved by added reference,
// all references for that user are also in the bag.
func (b bag) Contains(ref Reference) bool {
	var ks []refKey
	if id := ref.UserID; id != 0 {
		ks = append(ks, refKey{userID: id})
	}
	if h := ref.Handle; h != "" {
		ks = append(ks, refKey{handle: strings.TrimPrefix(h, "@")})
	}
	if e := ref.Email; e != "" {
		ks = append(ks, refKey{email: e})
	}
	for _, k := range ks {
		if _, ok := b.references[k]; ok {
			return true
		}
	}
	return false
}

func (b bag) FindResolved(ref Reference) (codeowners.ResolvedOwner, bool) {
	var ks []refKey
	if id := ref.UserID; id != 0 {
		ks = append(ks, refKey{userID: id})
	}
	if h := ref.Handle; h != "" {
		ks = append(ks, refKey{handle: strings.TrimPrefix(h, "@")})
	}
	if e := ref.Email; e != "" {
		ks = append(ks, refKey{email: e})
	}
	// Attempt to find user by any reference:
	for _, k := range ks {
		if refCtx, ok := b.references[k]; ok {
			if id := refCtx.resolvedUserID; id != 0 {
				userRefs := b.resolvedUsers[id]
				// TODO: Email resolution here is best effort,
				// we do not know if this is primary email.
				var email *string
				if len(userRefs.verifiedEmails) > 0 {
					e := userRefs.verifiedEmails[0]
					email = &e
				}
				return &codeowners.Person{
					User:         userRefs.user,
					PrimaryEmail: email,
					Handle:       userRefs.user.Username,
					// TODO: How to set email?
				}, true
			}
		}
	}
	// Best effort, return the default asssumed person based
	// on the input:
	return nil, false
}

func (b bag) String() string {
	var mapping []string
	for k, refCtx := range b.references {
		mapping = append(mapping, fmt.Sprintf("%s->%d", k, refCtx.resolvedUserID))
	}
	// TODO add unresolved references to the bag
	return fmt.Sprintf("[%s]", strings.Join(mapping, ", "))
}

// Add inserts all given references individually to the Bag.
// Next time Resolve is called, Bag will attempt to evaluate these
// against the database.
func (b *bag) Add(ref Reference) {
	if e := ref.Email; e != "" {
		b.add(refKey{email: e})
	}
	if h := ref.Handle; h != "" {
		b.add(refKey{handle: h})
	}
	if id := ref.UserID; id != 0 {
		b.add(refKey{userID: id})
	}
}

// add inserts given reference key (one of: user ID, email, handle)
// to the bag, so that it can be resolved later in batch.
func (b *bag) add(k refKey) {
	if _, ok := b.references[k]; !ok {
		b.references[k] = &refContext{}
	}
}

// Resolve takes all references that were added but not resolved
// before and queries the database to find corresponding users.
// Fetched users are augmented with all the other references that
// can point to them (also from the database), and the newly fetched
// references are then linked back to the bag.
func (b *bag) Resolve(ctx context.Context, db edb.EnterpriseDB) error {
	for k, refCtx := range b.references {
		if !refCtx.resolutionDone {
			userRefs, err := k.fetch(ctx, db)
			refCtx.resolutionDone = true
			if err != nil {
				return err
			}
			// Failed to resolve user.
			if userRefs == nil {
				continue
			}
			// User resolved successfully:
			refCtx.resolvedUserID = userRefs.id
			if _, ok := b.resolvedUsers[userRefs.id]; !ok {
				if err := userRefs.augment(ctx, db); err != nil {
					return err
				}
				b.linkBack(userRefs)
				b.resolvedUsers[userRefs.id] = userRefs
			}
		}
	}
	return nil
}

// linkBack adds all the extra references that were fetched for a user
// from the database (via `augment`) so that `Contains` can be valid
// for all known references to a user that is in the bag.
//
// For example: bag{refKey{email: alice@example.com}} is resolved.
// User with id=42 is fetched, that has second verified email: alice2@example.com,
// and a github handle aliceCodes. In that case calling linkBack on userReferences
// like above will result in bag with the following refKeys:
// {email:alice@example.com} -> 42
// {email:alice2@example.com} -> 42
// {handle:aliceCodes} -> 42
//
// TODO(#52441): For now the first handle or email assigned points to a user.
// This needs to be refined so that the same handle text can be considered
// in different contexts properly.
func (b *bag) linkBack(userRefs *userReferences) {
	ks := []refKey{{userID: userRefs.id}}
	if u := userRefs.user; u != nil {
		ks = append(ks, refKey{handle: u.Username})
	}
	for _, e := range userRefs.verifiedEmails {
		if _, ok := b.references[refKey{email: e}]; !ok {
			ks = append(ks, refKey{email: e})
		}
	}
	for _, h := range userRefs.codeHostHandles {
		if _, ok := b.references[refKey{handle: h}]; !ok {
			ks = append(ks, refKey{handle: h})
		}
	}
	for _, k := range ks {
		b.references[k] = &refContext{
			resolvedUserID: userRefs.id,
			resolutionDone: true,
		}
	}
}

// userReferences represents all the references found for a given user in the database.
// Every valid `userReferences` object has an `id`
type userReferences struct {
	// id must point at the ID of an actual user for userReferences to be valid.
	id   int32
	user *types.User
	// codeHostHandles are handles on the code-host that are linked with the user
	codeHostHandles []string
	verifiedEmails  []string
}

// augment fetches all the references for this user that are missing.
// These can then be linked back into the bag using `linkBack`.
// In order to call augment, `id`
func (r *userReferences) augment(ctx context.Context, db edb.EnterpriseDB) error {
	if r.id == 0 {
		return errors.New("userReferences needs id set for augmenting")
	}
	var err error
	if r.user == nil {
		r.user, err = db.Users().GetByID(ctx, r.id)
		if err != nil {
			return errors.Wrap(err, "augmenting user")
		}
	}
	if len(r.codeHostHandles) == 0 {
		r.codeHostHandles, err = fetchCodeHostHandles(ctx, db, r.id)
		if err != nil {
			return errors.Wrap(err, "augmenting code host handles")
		}
	}
	if len(r.verifiedEmails) == 0 {
		r.verifiedEmails, err = fetchVerifiedEmails(ctx, db, r.id)
		if err != nil {
			return errors.Wrap(err, "augmenting verified emails")
		}
	}
	return nil
}

func fetchVerifiedEmails(ctx context.Context, db edb.EnterpriseDB, userID int32) ([]string, error) {
	ves, err := db.UserEmails().ListByUser(ctx, database.UserEmailsListOptions{UserID: userID, OnlyVerified: true})
	if err != nil {
		return nil, errors.Wrap(err, "UserEmails.ListByUser")
	}
	var ms []string
	for _, email := range ves {
		ms = append(ms, email.Email)
	}
	return ms, nil
}

func fetchCodeHostHandles(ctx context.Context, db edb.EnterpriseDB, userID int32) ([]string, error) {
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
	return codeHostHandles, nil
}

// refKey is how the bag keys the references. Only one of the fields is filled.
type refKey struct {
	userID int32
	handle string
	email  string
}

func (k refKey) String() string {
	if id := k.userID; id != 0 {
		return fmt.Sprintf("#%d", id)
	}
	if h := k.handle; h != "" {
		return fmt.Sprintf("@%s", h)
	}
	if e := k.email; e != "" {
		return e
	}
	return "<empty refKey>"
}

// fetch pulls userReferences for given key from the database.
// It queries by email, userID or username based on what information is available.
func (k refKey) fetch(ctx context.Context, db edb.EnterpriseDB) (*userReferences, error) {
	if k.userID != 0 {
		return &userReferences{id: k.userID}, nil
	}
	if k.handle != "" {
		u, err := findUserByUsername(ctx, db, k.handle)
		if err != nil {
			return nil, err
		}
		if u == nil {
			return nil, nil
		}
		return &userReferences{id: u.ID, user: u}, nil
	}
	if k.email != "" {
		id, err := findUserIDByEmail(ctx, db, k.email)
		if err != nil {
			return nil, err
		}
		if id == 0 {
			return nil, nil
		}
		return &userReferences{id: id}, nil
	}
	return nil, errors.New("empty refKey is not valid")
}

func findUserByUsername(ctx context.Context, db edb.EnterpriseDB, handle string) (*types.User, error) {
	user, err := db.Users().GetByUsername(ctx, handle)
	if err != nil && !errcode.IsNotFound(err) {
		return nil, errors.Wrap(err, "Users.GetByUsername")
	}
	return user, nil
}

// TODO(#52246): GetVerifiedEmails accepts var-args, can batch
func findUserIDByEmail(ctx context.Context, db edb.EnterpriseDB, email string) (int32, error) {
	// Checking that provided email is verified.
	verifiedEmails, err := db.UserEmails().GetVerifiedEmails(ctx, email)
	if err != nil {
		return 0, errors.Wrap(err, "findUserIDByEmail")
	}
	if len(verifiedEmails) == 0 {
		return 0, nil
	}
	return verifiedEmails[0].UserID, nil
}

// refContext contains information about resolving a reference to a user.
type refContext struct {
	// resolvedUserID is not 0 if this reference has been recognized as a user.
	resolvedUserID int32
	// resolutionDone is set to true after the reference pointing at this refContext
	// has been attempted to be resolved.
	resolutionDone bool
}
