package own

import (
	"bytes"
	"context"
	"fmt"
	"net/mail"
	"strings"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/collections"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/errcode"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/azuredevops"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/bitbucketcloud"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/github"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/gitlab"
	"github.com/sourcegraph/sourcegraph/internal/own/codeowners"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
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
	// TeamID indicates identifying a specific team.
	TeamID int32
	// Handle is either a sourcegraph username, a code-host handle or a team name,
	// and can be considered within or outside of the repo context.
	Handle string
	// Email can be found in a CODEOWNERS entry, but can also
	// be a commit author email, which means it can be a code-host specific
	// email generated for the purpose of merging a pull-request.
	Email string
}

func (r Reference) ResolutionGuess() codeowners.ResolvedOwner {
	if r.Handle == "" && r.Email == "" {
		return nil
	}
	// If this is a GitHub repo and the handle contains a "/", then we can tell that this is a team.
	// TODO this does not work well with team resolver which expects team to be in the DB.
	if r.RepoContext != nil && strings.ToLower(r.RepoContext.CodeHostKind) == extsvc.VariantGitHub.AsType() && strings.Contains(r.Handle, "/") {
		return &codeowners.Team{
			Handle: r.Handle,
			Email:  r.Email,
			Team: &types.Team{
				Name:        r.Handle,
				DisplayName: r.Handle,
			},
		}
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
	Resolve(context.Context, database.DB)
}

func EmptyBag() *bag {
	return &bag{
		resolvedUsers: map[int32]*userReferences{},
		resolvedTeams: map[int32]*teamReferences{},
		references:    map[refKey]*refContext{},
	}
}

// ByTextReference returns a Bag of all the forms (users, persons, teams)
// that can be referred to by given text (name or email alike).
// This can be used in search to find relevant owners by different identifiers
// that the database reveals.
// TODO(#52141): Search by code host handle.
// TODO(#52246): ByTextReference uses fewer queries.
func ByTextReference(ctx context.Context, db database.DB, text ...string) Bag {
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
	b.Resolve(ctx, db)
	return b
}

// bag is implemented as a map of resolved users and map of references.
type bag struct {
	// resolvedUsers map from user id to `userReferences` which contain
	// all the references found in the database for a given user.
	// These references are linked back to the `references` via `resolve`
	// call.
	resolvedUsers map[int32]*userReferences
	// resolvedTeams map from team id to `teamReferences` which contains
	// just the team name.
	resolvedTeams map[int32]*teamReferences
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
	if id := ref.TeamID; id != 0 {
		ks = append(ks, refKey{teamID: id})
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
	if id := ref.TeamID; id != 0 {
		ks = append(ks, refKey{teamID: id})
	}
	// Attempt to find user by any reference:
	for _, k := range ks {
		if refCtx, ok := b.references[k]; ok {
			if id := refCtx.resolvedUserID; id != 0 {
				userRefs := b.resolvedUsers[id]
				if userRefs == nil || userRefs.user == nil {
					continue
				}
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
			if id := refCtx.resolvedTeamID; id != 0 {
				teamRefs := b.resolvedTeams[id]
				if teamRefs == nil || teamRefs.team == nil {
					continue
				}
				return &codeowners.Team{
					Team:   teamRefs.team,
					Handle: teamRefs.team.Name,
				}, true
			}
		}
	}
	return nil, false
}

func (b bag) String() string {
	var mapping []string
	for k, refCtx := range b.references {
		mapping = append(mapping, fmt.Sprintf("%s->%s", k, refCtx.resolvedIDForDebugging()))
	}
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
	if id := ref.TeamID; id != 0 {
		b.add(refKey{teamID: id})
	}
}

// add inserts given reference key (one of: user ID, team ID, email, handle)
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
func (b *bag) Resolve(ctx context.Context, db database.DB) {
	usersMap := make(map[*refContext]*userReferences)
	var userBatch userReferencesBatch
	for k, refCtx := range b.references {
		if !refCtx.resolutionDone {
			userRefs, teamRefs, err := k.fetch(ctx, db)
			refCtx.resolutionDone = true
			if err != nil {
				refCtx.appendErr(err)
			}
			// User resolved, adding to the map, to batch-augment them later.
			if userRefs != nil {
				// Checking added users in resolvedUsers map, but adding them to usersMap because
				// we need to have the whole refContext.
				if _, ok := b.resolvedUsers[userRefs.id]; !ok {
					usersMap[refCtx] = userRefs
					userBatch = append(userBatch, userRefs)
				}
			}
			// Team resolved
			if teamRefs != nil {
				id := teamRefs.team.ID
				if _, ok := b.resolvedTeams[id]; !ok {
					b.resolvedTeams[id] = teamRefs
				}
				// Team was referred to either by ID or by name, need to link back.
				teamRefs.linkBack(b)
				refCtx.resolvedTeamID = id
			}
		}
	}
	// Batch augment.
	userBatch.augment(ctx, db)
	// Post-augmentation actions.
	for refCtx, userRefs := range usersMap {
		b.resolvedUsers[userRefs.id] = userRefs
		userRefs.linkBack(b)
		refCtx.resolvedUserID = userRefs.id
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
	errs            []error
}

func (r *userReferences) appendErr(err error) {
	r.errs = append(r.errs, err)
}

type userReferencesBatch []*userReferences

// augment fetches all the references that are missing for all users in a batch.
// These can then be linked back into the bag using `linkBack`. In order to call
// augment, `id`.
func (b userReferencesBatch) augment(ctx context.Context, db database.DB) {
	userIDsToFetchHandles := collections.NewSet[int32]()
	for _, r := range b {
		// User references has to have an ID.
		if r.id == 0 {
			r.appendErr(errors.New("userReferences needs id set for augmenting"))
			continue
		}
		var err error
		if r.user == nil {
			r.user, err = db.Users().GetByID(ctx, r.id)
			if err != nil {
				r.appendErr(errors.Wrap(err, "augmenting user"))
			}
		}
		// Just adding the user ID to the set for a batch request.
		if len(r.codeHostHandles) == 0 {
			userIDsToFetchHandles.Add(r.id)
		}
		if len(r.verifiedEmails) == 0 {
			r.verifiedEmails, err = fetchVerifiedEmails(ctx, db, r.id)
			if err != nil {
				r.appendErr(errors.Wrap(err, "augmenting verified emails"))
			}
		}
	}
	if userIDsToFetchHandles.IsEmpty() {
		return
	}
	// Now we batch fetch all user accounts.
	handlesByUser, err := batchFetchCodeHostHandles(ctx, db, userIDsToFetchHandles.Values())
	if err != nil {
		// Well, we need to append errors to all the references.
		for _, r := range b {
			r.appendErr(err)
		}
		return
	}
	for _, r := range b {
		if handles, ok := handlesByUser[r.id]; ok {
			r.codeHostHandles = handles
		}
	}
}

func fetchVerifiedEmails(ctx context.Context, db database.DB, userID int32) ([]string, error) {
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

func batchFetchCodeHostHandles(ctx context.Context, db database.DB, userIDs []int32) (map[int32][]string, error) {
	accounts, err := db.UserExternalAccounts().ListForUsers(ctx, userIDs)
	if err != nil {
		return nil, errors.Wrap(err, "UserExternalAccounts.ListForUsers")
	}
	handlesByUser := make(map[int32][]string)
	for userID, accts := range accounts {
		handlesByUser[userID] = fetchCodeHostHandles(ctx, accts)
	}
	return handlesByUser, nil
}

func fetchCodeHostHandles(ctx context.Context, accounts []*extsvc.Account) []string {
	codeHostHandles := make([]string, 0, len(accounts))
	for _, account := range accounts {
		var accountInfo *extsvc.PublicAccountData
		var err error
		switch account.ServiceType {
		case extsvc.TypeGitHub:
			accountInfo, err = github.GetPublicExternalAccountData(ctx, &account.AccountData)
		case extsvc.TypeGitLab:
			accountInfo, err = gitlab.GetPublicExternalAccountData(ctx, &account.AccountData)
		case extsvc.TypeBitbucketCloud:
			accountInfo, err = bitbucketcloud.GetPublicExternalAccountData(ctx, &account.AccountData)
		case extsvc.TypeAzureDevOps:
			accountInfo, err = azuredevops.GetPublicExternalAccountData(ctx, &account.AccountData)
		}
		if err != nil {
			continue
		}

		if accountInfo != nil && accountInfo.Login != "" {
			codeHostHandles = append(codeHostHandles, accountInfo.Login)
		}
	}
	return codeHostHandles
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
func (r *userReferences) linkBack(b *bag) {
	ks := []refKey{{userID: r.id}}
	if u := r.user; u != nil {
		ks = append(ks, refKey{handle: u.Username})
	}
	for _, e := range r.verifiedEmails {
		if _, ok := b.references[refKey{email: e}]; !ok {
			ks = append(ks, refKey{email: e})
		}
	}
	for _, h := range r.codeHostHandles {
		if _, ok := b.references[refKey{handle: h}]; !ok {
			ks = append(ks, refKey{handle: h})
		}
	}
	for _, k := range ks {
		// Reference already present.
		// TODO(#52441): Keeping context with reference key can improve resolution.
		// For instance teams and users under the same name can be discerned
		// in github CODEOWNERS context (where only team name in CODEOWNERS
		// must contain `/`).
		if r, ok := b.references[k]; ok && r.successfullyResolved() {
			continue
		}
		b.references[k] = &refContext{
			resolvedUserID: r.id,
			resolutionDone: true,
		}
	}
}

type teamReferences struct {
	team *types.Team
}

func (r *teamReferences) linkBack(b *bag) {
	for _, k := range []refKey{{teamID: r.team.ID}, {handle: r.team.Name}} {
		// Reference already present.
		// TODO(#52441): Keeping context can improve conflict resolution.
		// For instance teams and users under the same name can be discerned
		// in github CODEOWNERS context (where only team name in CODEOWNERS
		// must contain `/`).
		if r, ok := b.references[k]; ok && r.successfullyResolved() {
			continue
		}
		b.references[k] = &refContext{
			resolvedTeamID: r.team.ID,
			resolutionDone: true,
		}
	}
}

// refKey is how the bag keys the references. Only one of the fields is filled.
type refKey struct {
	userID int32
	teamID int32
	handle string
	email  string
}

func (k refKey) String() string {
	if id := k.userID; id != 0 {
		return fmt.Sprintf("u%d", id)
	}
	if id := k.teamID; id != 0 {
		return fmt.Sprintf("t%d", id)
	}
	if h := k.handle; h != "" {
		return fmt.Sprintf("@%s", h)
	}
	if e := k.email; e != "" {
		return e
	}
	return "<empty refKey>"
}

// fetch pulls userReferences or teamReferences for given key from the database.
// It queries by email, userID, user name or team name based on what information
// is available.
func (k refKey) fetch(ctx context.Context, db database.DB) (*userReferences, *teamReferences, error) {
	if k.userID != 0 {
		return &userReferences{id: k.userID}, nil, nil
	}
	if k.teamID != 0 {
		t, err := findTeamByID(ctx, db, k.teamID)
		if err != nil {
			return nil, nil, err
		}
		// Weird situation: team is not found by ID. Cannot do much here.
		if t == nil {
			return nil, nil, errors.Newf("cannot find team by ID: %d", k.teamID)
		}
		return nil, &teamReferences{t}, nil
	}
	if k.handle != "" {
		u, err := findUserByUsername(ctx, db, k.handle)
		if err != nil {
			return nil, nil, err
		}
		if u != nil {
			return &userReferences{id: u.ID, user: u}, nil, nil
		}
		t, err := findTeamByName(ctx, db, k.handle)
		if err != nil {
			return nil, nil, err
		}
		if t != nil {
			return nil, &teamReferences{t}, nil
		}
	}
	if k.email != "" {
		u, err := findUserByEmail(ctx, db, k.email)
		if err != nil {
			return nil, nil, err
		}
		if u != nil {
			return &userReferences{id: u.ID, user: u}, nil, nil
		}
	}
	// Neither user nor team was found.
	return nil, nil, nil
}

func findUserByUsername(ctx context.Context, db database.DB, handle string) (*types.User, error) {
	user, err := db.Users().GetByUsername(ctx, handle)
	if err != nil && !errcode.IsNotFound(err) {
		return nil, errors.Wrap(err, "Users.GetByUsername")
	}
	return user, nil
}

func findUserByEmail(ctx context.Context, db database.DB, email string) (*types.User, error) {
	// Checking that provided email is verified.
	user, err := db.Users().GetByVerifiedEmail(ctx, email)
	if err != nil && !errcode.IsNotFound(err) {
		return nil, errors.Wrap(err, "findUserIDByEmail")
	}
	return user, nil
}

func findTeamByID(ctx context.Context, db database.DB, id int32) (*types.Team, error) {
	team, err := db.Teams().GetTeamByID(ctx, id)
	if err != nil && !errcode.IsNotFound(err) {
		return nil, errors.Wrap(err, "Teams.GetTeamByID")
	}
	return team, nil
}

func findTeamByName(ctx context.Context, db database.DB, name string) (*types.Team, error) {
	team, err := db.Teams().GetTeamByName(ctx, name)
	if err != nil && !errcode.IsNotFound(err) {
		return nil, errors.Wrap(err, "Teams.GetTeamByName")
	}
	return team, nil
}

// refContext contains information about resolving a reference to a user.
type refContext struct {
	// resolvedUserID is not 0 if this reference has been recognized as a user.
	resolvedUserID int32
	// resolvedTeamID is not 0 if this reference has been recognized as a team.
	resolvedTeamID int32
	// resolutionDone is set to true after the reference pointing at this refContext
	// has been attempted to be resolved.
	resolutionDone bool
	resolutionErrs []error
}

// successfullyResolved context either points to a team or to a user.
func (c refContext) successfullyResolved() bool {
	return c.resolvedUserID != 0 || c.resolvedTeamID != 0
}

func (c *refContext) appendErr(err error) {
	c.resolutionErrs = append(c.resolutionErrs, err)
}

func (c refContext) resolvedIDForDebugging() string {
	if id := c.resolvedUserID; id != 0 {
		return fmt.Sprintf("user-%d", id)
	}
	if id := c.resolvedTeamID; id != 0 {
		return fmt.Sprintf("team-%d", id)
	}
	return "<nil>"
}
