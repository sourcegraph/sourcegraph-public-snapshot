package own

import (
	"context"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/database"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/own/codeowners"
	"github.com/sourcegraph/sourcegraph/internal/api"
)

// repoContext allows us to anchor an author reference to a repo where it stems from.
// For instance a handle from a CODEOWNERS file comes from github.com/sourcegraph/sourcegraph.
// This is important for resolving namespaced owner names
// (like CODEOWNERS file can refer to team handle "own"), while the name in the database is "sourcegraph/own"
// because it was pulled from github, and by convention organiziation name is prepended.
type repoContext struct {
	name         api.RepoName
	codehostKind string
}

// Reference is whatever we get from a data source, like a commit,
// CODEOWNERS entry or file view.
type Reference struct {
	// repoContext is present if given owner reference is associated
	// with specific repository.
	repoContext *repoContext
	// userID indicates identifying a specific user.
	userID int
	// handle is either a sourcegraph or code-host handle,
	// and can be considered within or outside of
	handle string
	// email can be found in a CODEOWNERS entry, but can also
	// be a commit author email, which means it can be a code-host specific
	// email generated for the purpose of merging a pull-request.
	email string
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
func ByTextReference(context.Context, database.EnterpriseDB, string) (Bag, error) {
	return nil, nil
}

// Example: Implement search for f:has.owner(eseliger)
func searchExample(db database.EnterpriseDB) error {
	ctx := context.Background()
	ownerSearchTerm := "jdoe"
	// Do this at first during search and hold references to all the known entities
	// that can be referred to by given search term
	bag, err := ByTextReference(ctx, db, ownerSearchTerm)
	if err != nil {
		return err
	}

	// Then for given file we have owner matches (translated to references here):
	ownerReferences := []Reference{
		// Some possible matching entries:
		// email entry in CODEOWNERS
		{
			email: "john.doe@sourcegraph.com",
			repoContext: &repoContext{
				name:         "github.com/sourcegraph/sourcegraph",
				codehostKind: "github",
			},
		},
		// @jdoe entry in CODEOWNERS
		{
			handle: "jdoe",
			repoContext: &repoContext{
				name:         "github.com/sourcegraph/sourcegraph",
				codehostKind: "github",
			},
		},
		{
			handle: "jdoe.sourcegraph",
			repoContext: &repoContext{
				name:         "github.com/sourcegraph/sourcegraph",
				codehostKind: "github",
			},
		},
		{userID: 42}, // John Doe's user ID from assigned ownership
	}
	var matches bool
	for _, ref := range ownerReferences {
		if bag.Contains(ref) {
			matches = true
		}
	}
	if matches {
		// Great! We're selecting the result of filtering.
	}
	return nil
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
			email: "john.doe@sourcegraph.com",
			repoContext: &repoContext{
				name:         "github.com/sourcegraph/sourcegraph",
				codehostKind: "github",
			},
		},
		// user handle entry in CODEOWNERS
		{
			handle: "johndoe",
			repoContext: &repoContext{
				name:         "github.com/sourcegraph/sourcegraph",
				codehostKind: "github",
			},
		},
		// team handle in CODEOWNERS - we know it's a team because of github code host and / in the name
		{
			handle: "sourcegraph/own",
			repoContext: &repoContext{
				name:         "github.com/sourcegraph/sourcegraph",
				codehostKind: "github",
			},
		},
		// Contributor email contains github username, we can figure this out based on github code host.
		{
			email: "githubusername@users.noreply.github.com",
			// alternative:  "userID+userName@users.noreply.github.com",
			// repo context where the commit is from
			repoContext: &repoContext{
				name:         "github.com/sourcegraph/sourcegraph",
				codehostKind: "github",
			},
		},
		{userID: 42}, // John's user ID from assigned ownership
		{userID: 42}, // User ID originating from recent viewer signal.
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
