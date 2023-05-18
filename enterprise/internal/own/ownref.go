package own

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	"github.com/sourcegraph/sourcegraph/lib/errors"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/database"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/own/codeowners"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/errcode"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

// repoContext allows us to anchor an author reference to a repo where it stems from.
// For instance a handle from a CODEOWNERS file comes from github.com/sourcegraph/sourcegraph.
// This is important for resolving namespaced owner names
// (like CODEOWNERS file can refer to team handle "own"), while the name in the database is "sourcegraph/own"
// because it was pulled from github, and by convention organiziation name is prepended.
type RepoContext struct {
	Name         api.RepoName
	CodehostKind string
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
	// and can be considered within or outside of
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
		fmt.Fprintf(&b, "context.%s: %s", c.CodehostKind, c.Name)
	}
	fmt.Fprint(&b, "}")
	return b.String()
}

// Bag is a collection of platonic forms or identities of owners
// (teams, people or otherwise). It is pre-seeded with database information
// based on the search query using `ByTextReference`, and used to match owner
// references found.
type Bag interface {

	// Contains answers true if given bag contains an owner form
	// that the given reference points at in some way.
	Contains(ref Reference) bool
}

// ByTextReference returns a Bag of all the forms (users, persons, teams)
// that can be referred to by given text (name or email alike).
// This can be used in search to find relevant owners by different identifiers
// that the database reveals.
// TODO: Search by verified email.
// TODO: Search by code host handle.
func ByTextReference(ctx context.Context, db database.EnterpriseDB, text string) (Bag, error) {
	if strings.HasPrefix(text, "@") {
		text = text[1:]
	}
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
		email, verified, err := db.UserEmails().GetPrimaryEmail(ctx, userRefs.id)
		if err != nil && !errcode.IsNotFound(err) {
			return nil, errors.Wrap(err, "UserEmails.GetPrimaryEmail")
		}
		if !verified {
			continue
		}
		userRefs.verifiedEmails = append(userRefs.verifiedEmails, email)
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

// TODO: Introduce matching on linked code host handles.
func (refs userReferences) containsHandle(handle string) bool {
	if strings.HasPrefix(handle, "@") {
		handle = handle[1:]
	}
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
//   - if email reference matches the primary email,
//     TODO: Match also other verified emails
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

// Clusters is a set of associated references with the use of a database.
// The way it works is that:
// 1. references are first added as leads,
// 2. database is used to resolve set of different references,
// 3. then the same references can be looked up and resolved to concrete people or teams.
type Clusters interface {
	// Add a reference to enrich the data set
	Add(ref Reference)

	// Resolve all the added references, and cluster by owner identity.
	Resolve(context.Context, database.EnterpriseDB) error

	// Look up resolved references. If two references evaluate to a single
	// resolved owner, the result for them is guaranteed to be the same,
	// and different resolved owners are guaranteed to have different Identifier().
	Lookup(ref Reference) codeowners.ResolvedOwner
}

// Example: resolution for file/repo/directory ownership
func resolutionExample(db database.EnterpriseDB) error {
	ctx := context.Background()
	// here are some signals and owner data that is returned for given
	// file or directory:
	ownerReferences := []Reference{
		// email entry in CODEOWNERS
		{
			Email: "john.doe@sourcegraph.com",
			RepoContext: &RepoContext{
				Name:         "github.com/sourcegraph/sourcegraph",
				CodehostKind: "github",
			},
		},
		// user handle entry in CODEOWNERS
		{
			Handle: "johndoe",
			RepoContext: &RepoContext{
				Name:         "github.com/sourcegraph/sourcegraph",
				CodehostKind: "github",
			},
		},
		// team handle in CODEOWNERS - we know it's a team because of github code host and / in the name
		{
			Handle: "sourcegraph/own",
			RepoContext: &RepoContext{
				Name:         "github.com/sourcegraph/sourcegraph",
				CodehostKind: "github",
			},
		},
		// Contributor email contains github username, we can figure this out based on github code host.
		{
			Email: "githubusername@users.noreply.github.com",
			// alternative:  "userID+userName@users.noreply.github.com",
			// repo context where the commit is from
			RepoContext: &RepoContext{
				Name:         "github.com/sourcegraph/sourcegraph",
				CodehostKind: "github",
			},
		},
		{UserID: 42}, // John's user ID from assigned ownership
		{UserID: 42}, // User ID originating from recent viewer signal.
	}
	var cls Clusters = nil // construct somehow
	for _, r := range ownerReferences {
		cls.Add(r)
	}
	if err := cls.Resolve(ctx, db); err != nil {
		return err
	}
	// iterate through ownership and signals found again to group
	grouped := map[string][]Reference{}
	for _, r := range ownerReferences {
		// Here we iterate through references, but in the ownership blob/repo/dir
		// resolver each reference will be attached to a signal, so we'll be able
		// to group these, like so:
		o := cls.Lookup(r)
		rs := grouped[o.Identifier()]
		rs = append(rs, r)
		grouped[o.Identifier()] = rs
	}
	// We have a map of references by resolved owner identity. In resolver
	// we're likely also accumulating the resolved owner with signals.
	return nil
}
