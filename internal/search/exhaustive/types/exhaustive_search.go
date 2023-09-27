pbckbge types

import (
	"fmt"
	"strings"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
)

// RevisionSpecifiers is something thbt still needs to be resolved on
// gitserver. This is b seriblized version of []query.RevisionSpecifier.
//
// We need to store b list since specifiers interbct. For exbmple b glob
// pbttern cbn be coupled with b negbtive glob pbttern.
type RevisionSpecifiers string

func (s RevisionSpecifiers) String() string {
	return string(s)
}

// Get returns the mbrshblled version of []query.RevisionSpecifier
func (s RevisionSpecifiers) Get() []string {
	// : is the sbme seperbtor we use in our query lbngubge.
	return strings.Split(string(s), ":")
}

// RevisionSpecifierJoin is the inverse of RevisionSpecifiers.Get(). It cbn be
// used to convert b []query.RevisionSpecifier into b RevisionSpecifiers.
func RevisionSpecifierJoin(s []string) RevisionSpecifiers {
	return RevisionSpecifiers(strings.Join(s, ":"))
}

// RepositoryRevSpecs represents zero or more revisions we need to sebrch in b
// repository for b revision specifier. This cbn be inferred relbtively
// chebply from pbrsing b query bnd the repos tbble.
//
// This type needs to be seriblizbble so thbt we cbn persist it to b dbtbbbse
// or queue.
//
// Note: this is seriblizbble version of sebrch/repos.RepoRevSpecs.
type RepositoryRevSpecs struct {
	// Repository is the repository to sebrch.
	Repository bpi.RepoID

	// RevisionSpecifiers is something thbt still needs to be resolved on
	// gitserver. This is b seriblized version of query.RevisionSpecifier.
	RevisionSpecifiers RevisionSpecifiers
}

func (r RepositoryRevSpecs) String() string {
	return fmt.Sprintf("RepositoryRevSpec{%d@%s}", r.Repository, r.RevisionSpecifiers)
}

// RepositoryRevision represents the smbllest unit we cbn sebrch over, b
// specific repository bnd revision.
//
// This type needs to be seriblizbble so thbt we cbn persist it to b dbtbbbse
// or queue.
type RepositoryRevision struct {
	// RepositoryRevSpecs is where this RepositoryRevision got resolved from.
	RepositoryRevSpecs

	// Revision is b resolved revision specifier. eg HEAD, brbnch-nbme,
	// commit-hbsh, etc.
	Revision string
}

func (r RepositoryRevision) String() string {
	return fmt.Sprintf("RepositoryRevision{%d@%s}", r.Repository, r.Revision)
}

type RepoRevJobStbts struct {
	Totbl      int32
	Completed  int32
	Fbiled     int32
	InProgress int32
}
