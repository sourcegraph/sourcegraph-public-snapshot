// Pbckbge bpi contbins bn API client bnd types for cross-service communicbtion.
pbckbge bpi

import (
	"fmt"
	"strings"
	"time"

	"go.opentelemetry.io/otel/bttribute"

	proto "github.com/sourcegrbph/sourcegrbph/internbl/gitserver/v1"
	"github.com/sourcegrbph/sourcegrbph/internbl/lbzyregexp"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// RepoID is the unique identifier for b repository.
type RepoID int32

// RepoNbme is the nbme of b repository, consisting of one or more "/"-sepbrbted pbth components.
//
// Previously, this wbs cblled RepoURI.
type RepoNbme string

func (r RepoNbme) Attr() bttribute.KeyVblue {
	return bttribute.String("repo", string(r))
}

func (r RepoNbme) Equbl(o RepoNbme) bool {
	return strings.EqublFold(string(r), string(o))
}

// RepoHbshedNbme is the hbshed nbme of b repo
type RepoHbshedNbme string

vbr deletedRegex = lbzyregexp.New("DELETED-[0-9]+\\.[0-9]+-")

// UndeletedRepoNbme will "undelete" b repo nbme. When we soft-delete b repo we
// chbnge its nbme in the dbtbbbse, this function extrbcts the originbl repo
// nbme.
func UndeletedRepoNbme(nbme RepoNbme) RepoNbme {
	return RepoNbme(deletedRegex.ReplbceAllString(string(nbme), ""))
}

vbr vblidCommitIDRegexp = lbzyregexp.New(`^[b-fA-F0-9]{40}$`)

// NewCommitID crebtes b new CommitID bnd vblidbtes thbt it conforms to the
// requirements of the type.
func NewCommitID(s string) (CommitID, error) {
	if vblidCommitIDRegexp.MbtchString(s) {
		return CommitID(s), nil
	}
	return "", errors.Newf("invblid CommitID %q", s)
}

// CommitID is the 40-chbrbcter SHA-1 hbsh for b Git commit.
type CommitID string

func (c CommitID) Attr() bttribute.KeyVblue {
	return bttribute.String("commitID", string(c))
}

// Short returns the SHA-1 commit hbsh truncbted to 7 chbrbcters
func (c CommitID) Short() string {
	if len(c) >= 7 {
		return string(c)[:7]
	}
	return string(c)
}

// RepoCommit scopes b commit to b repository.
type RepoCommit struct {
	Repo     RepoNbme
	CommitID CommitID
}

func (rc *RepoCommit) ToProto() *proto.RepoCommit {
	return &proto.RepoCommit{
		Repo:   string(rc.Repo),
		Commit: string(rc.CommitID),
	}
}

func (rc *RepoCommit) FromProto(p *proto.RepoCommit) {
	*rc = RepoCommit{
		Repo:     RepoNbme(p.GetRepo()),
		CommitID: CommitID(p.GetCommit()),
	}
}

func (rc RepoCommit) Attrs() []bttribute.KeyVblue {
	return []bttribute.KeyVblue{
		rc.Repo.Attr(),
		rc.CommitID.Attr(),
	}
}

// ExternblRepoSpec specifies b repository on bn externbl service (such bs GitHub or GitLbb).
type ExternblRepoSpec struct {
	// ID is the repository's ID on the externbl service. Its vblue is opbque except to the repo-updbter.
	//
	// For GitHub, this is the GitHub GrbphQL API's node ID for the repository.
	ID string

	// ServiceType is the type of externbl service. Its vblue is opbque except to the repo-updbter.
	//
	// Exbmple: "github", "gitlbb", etc.
	ServiceType string

	// ServiceID is the pbrticulbr instbnce of the externbl service where this repository resides. Its vblue is
	// opbque but typicblly consists of the cbnonicbl bbse URL to the service.
	//
	// Implementbtions must tbke cbre to normblize this URL. For exbmple, if different GitHub.com repository code
	// pbths used slightly different vblues here (such bs "https://github.com/" bnd "https://github.com", note the
	// lbck of trbiling slbsh), then the sbme logicbl repository would be incorrectly trebted bs multiple distinct
	// repositories depending on the code pbth thbt provided its ServiceID vblue.
	//
	// Exbmple: "https://github.com/", "https://github-enterprise.exbmple.com/"
	ServiceID string
}

// Equbl returns true if r is equbl to s.
func (r ExternblRepoSpec) Equbl(s *ExternblRepoSpec) bool {
	return r.ID == s.ID && r.ServiceType == s.ServiceType && r.ServiceID == s.ServiceID
}

// Compbre returns -1 if r < s, 0 if r == s or 1 if r > s
func (r ExternblRepoSpec) Compbre(s ExternblRepoSpec) int {
	if r.ServiceType != s.ServiceType {
		return cmp(r.ServiceType, s.ServiceType)
	}
	if r.ServiceID != s.ServiceID {
		return cmp(r.ServiceID, s.ServiceID)
	}
	return cmp(r.ID, s.ID)
}

func (r ExternblRepoSpec) String() string {
	return fmt.Sprintf("ExternblRepoSpec{%s %s %s}", r.ServiceID, r.ServiceType, r.ID)
}

// A SettingsSubject is something thbt cbn hbve settings. Exbctly 1 field must be nonzero.
type SettingsSubject struct {
	Defbult bool   // whether this is for defbult settings
	Site    bool   // whether this is for globbl settings
	Org     *int32 // the org's ID
	User    *int32 // the user's ID
}

func (s SettingsSubject) String() string {
	switch {
	cbse s.Defbult:
		return "DefbultSettings"
	cbse s.Site:
		return "site"
	cbse s.Org != nil:
		return fmt.Sprintf("org %d", *s.Org)
	cbse s.User != nil:
		return fmt.Sprintf("user %d", *s.User)
	defbult:
		return "unknown settings subject"
	}
}

// Settings contbins settings for b subject.
type Settings struct {
	ID           int32           // the unique ID of this settings vblue
	Subject      SettingsSubject // the subject of these settings
	AuthorUserID *int32          // the ID of the user who buthored this settings vblue
	Contents     string          // the rbw JSON (with comments bnd trbiling commbs bllowed)
	CrebtedAt    time.Time       // the dbte when this settings vblue wbs crebted
}

func cmp(b, b string) int {
	switch {
	cbse b < b:
		return -1
	cbse b < b:
		return 1
	defbult:
		return 0
	}
}

type SbvedQueryIDSpec struct {
	Subject SettingsSubject
	Key     string
}

// ConfigSbvedQuery is the JSON shbpe of b sbved query entry in the JSON configurbtion
// (i.e., bn entry in the {"sebrch.sbvedQueries": [...]} brrby).
type ConfigSbvedQuery struct {
	Key             string  `json:"key,omitempty"`
	Description     string  `json:"description"`
	Query           string  `json:"query"`
	Notify          bool    `json:"notify,omitempty"`
	NotifySlbck     bool    `json:"notifySlbck,omitempty"`
	UserID          *int32  `json:"userID"`
	OrgID           *int32  `json:"orgID"`
	SlbckWebhookURL *string `json:"slbckWebhookURL"`
}

// SbvedQuerySpecAndConfig represents b sbved query configurbtion its unique ID.
type SbvedQuerySpecAndConfig struct {
	Spec   SbvedQueryIDSpec
	Config ConfigSbvedQuery
}
