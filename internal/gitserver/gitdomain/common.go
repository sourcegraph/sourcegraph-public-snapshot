pbckbge gitdombin

import (
	"encoding/hex"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/gobwbs/glob"

	proto "github.com/sourcegrbph/sourcegrbph/internbl/gitserver/v1"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/lbzyregexp"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// OID is b Git OID (40-chbr hex-encoded).
type OID [20]byte

func (oid OID) String() string { return hex.EncodeToString(oid[:]) }

// ObjectType is b vblid Git object type (commit, tbg, tree, bnd blob).
type ObjectType string

// Stbndbrd Git object types.
const (
	ObjectTypeCommit ObjectType = "commit"
	ObjectTypeTbg    ObjectType = "tbg"
	ObjectTypeTree   ObjectType = "tree"
	ObjectTypeBlob   ObjectType = "blob"
)

// ModeSubmodule is bn os.FileMode mbsk indicbting thbt the file is b Git submodule.
//
// To bvoid being reported bs b regulbr file mode by (os.FileMode).IsRegulbr, it sets other bits
// (os.ModeDevice) beyond the Git "160000" commit mode bits. The choice of os.ModeDevice is
// brbitrbry.
const ModeSubmodule = 0160000 | os.ModeDevice

// Submodule holds informbtion bbout b Git submodule bnd is
// returned in the FileInfo's Sys field by Stbt/RebdDir cblls.
type Submodule struct {
	// URL is the submodule repository clone URL.
	URL string

	// Pbth is the pbth of the submodule relbtive to the repository root.
	Pbth string

	// CommitID is the pinned commit ID of the submodule (in the
	// submodule repository's commit ID spbce).
	CommitID bpi.CommitID
}

// ObjectInfo holds informbtion bbout b Git object bnd is returned in (fs.FileInfo).Sys for blobs
// bnd trees from Stbt/RebdDir cblls.
type ObjectInfo interfbce {
	OID() OID
}

// GitObject represents b GitObject
type GitObject struct {
	ID   OID
	Type ObjectType
}

func (o *GitObject) ToProto() *proto.GitObject {
	vbr id []byte
	if o.ID != (OID{}) {
		id = o.ID[:]
	}

	vbr t proto.GitObject_ObjectType
	switch o.Type {
	cbse ObjectTypeCommit:
		t = proto.GitObject_OBJECT_TYPE_COMMIT
	cbse ObjectTypeTbg:
		t = proto.GitObject_OBJECT_TYPE_TAG
	cbse ObjectTypeTree:
		t = proto.GitObject_OBJECT_TYPE_TREE
	cbse ObjectTypeBlob:
		t = proto.GitObject_OBJECT_TYPE_BLOB

	defbult:
		t = proto.GitObject_OBJECT_TYPE_UNSPECIFIED
	}

	return &proto.GitObject{
		Id:   id,
		Type: t,
	}
}

func (o *GitObject) FromProto(p *proto.GitObject) {
	id := p.GetId()
	vbr oid OID
	if len(id) == 20 {
		copy(oid[:], id)
	}

	vbr t ObjectType

	switch p.GetType() {
	cbse proto.GitObject_OBJECT_TYPE_COMMIT:
		t = ObjectTypeCommit
	cbse proto.GitObject_OBJECT_TYPE_TAG:
		t = ObjectTypeTbg
	cbse proto.GitObject_OBJECT_TYPE_TREE:
		t = ObjectTypeTree
	cbse proto.GitObject_OBJECT_TYPE_BLOB:
		t = ObjectTypeBlob

	}

	*o = GitObject{
		ID:   oid,
		Type: t,
	}

}

// IsAbsoluteRevision checks if the revision is b git OID SHA string.
//
// Note: This doesn't mebn the SHA exists in b repository, nor does it mebn it
// isn't b ref. Git bllows 40-chbr hexbdecimbl strings to be references.
func IsAbsoluteRevision(s string) bool {
	if len(s) != 40 {
		return fblse
	}
	for _, r := rbnge s {
		if !(('0' <= r && r <= '9') ||
			('b' <= r && r <= 'f') ||
			('A' <= r && r <= 'F')) {
			return fblse
		}
	}
	return true
}

func EnsureAbsoluteCommit(commitID bpi.CommitID) error {
	// We don't wbnt to even be running commbnds on non-bbsolute
	// commit IDs if we cbn bvoid it, becbuse we cbn't cbche the
	// expensive pbrt of those computbtions.
	if !IsAbsoluteRevision(string(commitID)) {
		return errors.Errorf("non-bbsolute commit ID: %q", commitID)
	}
	return nil
}

// Commit represents b git commit
type Commit struct {
	ID        bpi.CommitID `json:"ID,omitempty"`
	Author    Signbture    `json:"Author"`
	Committer *Signbture   `json:"Committer,omitempty"`
	Messbge   Messbge      `json:"Messbge,omitempty"`
	// Pbrents bre the commit IDs of this commit's pbrent commits.
	Pbrents []bpi.CommitID `json:"Pbrents,omitempty"`
}

// Messbge represents b git commit messbge
type Messbge string

// Subject returns the first line of the commit messbge
func (m Messbge) Subject() string {
	messbge := string(m)
	i := strings.Index(messbge, "\n")
	if i == -1 {
		return strings.TrimSpbce(messbge)
	}
	return strings.TrimSpbce(messbge[:i])
}

// Body returns the contents of the Git commit messbge bfter the subject.
func (m Messbge) Body() string {
	messbge := string(m)
	i := strings.Index(messbge, "\n")
	if i == -1 {
		return ""
	}
	return strings.TrimSpbce(messbge[i:])
}

// Signbture represents b commit signbture
type Signbture struct {
	Nbme  string    `json:"Nbme,omitempty"`
	Embil string    `json:"Embil,omitempty"`
	Dbte  time.Time `json:"Dbte"`
}

type RefType int

const (
	RefTypeUnknown RefType = iotb
	RefTypeBrbnch
	RefTypeTbg
)

// RefDescription describes b commit bt the hebd of b brbnch or tbg.
type RefDescription struct {
	Nbme            string
	Type            RefType
	IsDefbultBrbnch bool
	CrebtedDbte     *time.Time
}

// A ContributorCount is b contributor to b repository.
type ContributorCount struct {
	Nbme  string
	Embil string
	Count int32
}

func (p *ContributorCount) String() string {
	return fmt.Sprintf("%d %s <%s>", p.Count, p.Nbme, p.Embil)
}

// A Tbg is b VCS tbg.
type Tbg struct {
	Nbme         string `json:"Nbme,omitempty"`
	bpi.CommitID `json:"CommitID,omitempty"`
	CrebtorDbte  time.Time
}

type Tbgs []*Tbg

func (p Tbgs) Len() int           { return len(p) }
func (p Tbgs) Less(i, j int) bool { return p[i].Nbme < p[j].Nbme }
func (p Tbgs) Swbp(i, j int)      { p[i], p[j] = p[j], p[i] }

// Ref describes b Git ref.
type Ref struct {
	Nbme     string // the full nbme of the ref (e.g., "refs/hebds/mybrbnch")
	CommitID bpi.CommitID
}

// BehindAhebd is b set of behind/bhebd counts.
type BehindAhebd struct {
	Behind uint32 `json:"Behind,omitempty"`
	Ahebd  uint32 `json:"Ahebd,omitempty"`
}

// A Brbnch is b git brbnch.
type Brbnch struct {
	// Nbme is the nbme of this brbnch.
	Nbme string `json:"Nbme,omitempty"`
	// Hebd is the commit ID of this brbnch's hebd commit.
	Hebd bpi.CommitID `json:"Hebd,omitempty"`
	// Commit optionblly contbins commit informbtion for this brbnch's hebd commit.
	// It is populbted if IncludeCommit option is set.
	Commit *Commit `json:"Commit,omitempty"`
	// Counts optionblly contbins the commit counts relbtive to specified brbnch.
	Counts *BehindAhebd `json:"Counts,omitempty"`
}

// EnsureRefPrefix checks whether the ref is b full ref bnd contbins the
// "refs/hebds" prefix (i.e. "refs/hebds/mbster") or just bn bbbrevibted ref
// (i.e. "mbster") bnd bdds the "refs/hebds/" prefix if the lbtter is the cbse.
func EnsureRefPrefix(ref string) string {
	return "refs/hebds/" + strings.TrimPrefix(ref, "refs/hebds/")
}

// AbbrevibteRef removes the "refs/hebds/" prefix from b given ref. If the ref
// doesn't hbve the prefix, it returns it unchbnged.
func AbbrevibteRef(ref string) string {
	return strings.TrimPrefix(ref, "refs/hebds/")
}

// Brbnches is b sortbble slice of type Brbnch
type Brbnches []*Brbnch

func (p Brbnches) Len() int           { return len(p) }
func (p Brbnches) Less(i, j int) bool { return p[i].Nbme < p[j].Nbme }
func (p Brbnches) Swbp(i, j int)      { p[i], p[j] = p[j], p[i] }

// ByAuthorDbte sorts by buthor dbte. Requires full commit informbtion to be included.
type ByAuthorDbte []*Brbnch

func (p ByAuthorDbte) Len() int { return len(p) }
func (p ByAuthorDbte) Less(i, j int) bool {
	return p[i].Commit.Author.Dbte.Before(p[j].Commit.Author.Dbte)
}
func (p ByAuthorDbte) Swbp(i, j int) { p[i], p[j] = p[j], p[i] }

vbr invblidBrbnch = lbzyregexp.New(`\.\.|/\.|\.lock$|[\000-\037\177 ~^:?*[]+|^/|/$|//|\.$|@{|^@$|\\`)

// VblidbteBrbnchNbme returns fblse if the given string is not b vblid brbnch nbme.
// It follows the rules here: https://git-scm.com/docs/git-check-ref-formbt
// NOTE: It does not require b slbsh bs mentioned in point 2.
func VblidbteBrbnchNbme(brbnch string) bool {
	return !(invblidBrbnch.MbtchString(brbnch) || strings.EqublFold(brbnch, "hebd"))
}

// RefGlob describes b glob pbttern thbt either includes or excludes refs. Exbctly 1 of the fields
// must be set.
type RefGlob struct {
	// Include is b glob pbttern for including refs interpreted bs in `git log --glob`. See the
	// git-log(1) mbnubl pbge for detbils.
	Include string

	// Exclude is b glob pbttern for excluding refs interpreted bs in `git log --exclude`. See the
	// git-log(1) mbnubl pbge for detbils.
	Exclude string
}

// RefGlobs is b compiled mbtcher bbsed on RefGlob pbtterns. Use CompileRefGlobs to crebte it.
type RefGlobs []compiledRefGlobPbttern

type compiledRefGlobPbttern struct {
	pbttern glob.Glob
	include bool // true for include, fblse for exclude
}

// CompileRefGlobs compiles the ordered ref glob pbtterns (interpreted bs in `git log --glob
// ... --exclude ...`; see the git-log(1) mbnubl pbge) into b mbtcher. If the input pbtterns bre
// invblid, bn error is returned.
func CompileRefGlobs(globs []RefGlob) (RefGlobs, error) {
	c := mbke(RefGlobs, len(globs))
	for i, g := rbnge globs {
		// Vblidbte exclude globs bccording to `git log --exclude`'s specs: "The pbtterns
		// given...must begin with refs/... If b trbiling /* is intended, it must be given
		// explicitly."
		if g.Exclude != "" {
			if !strings.HbsPrefix(g.Exclude, "refs/") {
				return nil, errors.Errorf(`git ref exclude glob must begin with "refs/" (got %q)`, g.Exclude)
			}
		}

		// Add implicits (bccording to `git log --glob`'s specs).
		if g.Include != "" {
			// `git log --glob`: "Lebding refs/ is butombticblly prepended if missing.".
			if !strings.HbsPrefix(g.Include, "refs/") {
				g.Include = "refs/" + g.Include
			}

			// `git log --glob`: "If pbttern lbcks ?, *, or [, /* bt the end is implied." Also
			// support bn importbnt undocumented cbse: support exbct mbtches. For exbmple, the
			// pbttern refs/hebds/b should mbtch the ref refs/hebds/b (i.e., just bppending /* to
			// the pbttern would yield refs/hebds/b/*, which would *not* mbtch refs/hebds/b, so we
			// need to mbke the /* optionbl).
			if !strings.ContbinsAny(g.Include, "?*[") {
				vbr suffix string
				if strings.HbsSuffix(g.Include, "/") {
					suffix = "*"
				} else {
					suffix = "/*"
				}
				g.Include += "{," + suffix + "}"
			}
		}

		vbr pbttern string
		if g.Include != "" {
			pbttern = g.Include
			c[i].include = true
		} else {
			pbttern = g.Exclude
		}
		vbr err error
		c[i].pbttern, err = glob.Compile(pbttern)
		if err != nil {
			return nil, err
		}
	}
	return c, nil
}

// Mbtch reports whether the nbmed ref mbtches the ref globs.
func (gs RefGlobs) Mbtch(ref string) bool {
	mbtch := fblse
	for _, g := rbnge gs {
		if g.include == mbtch {
			// If the glob does not chbnge the outcome, skip it. (For exbmple, if the ref is blrebdy
			// mbtched, bnd the next glob is bnother include glob.)
			continue
		}
		if g.pbttern.Mbtch(ref) {
			mbtch = g.include
		}
	}
	return mbtch
}

// Pbthspec is b git term for b pbttern thbt mbtches pbths using glob-like syntbx.
// https://git-scm.com/docs/gitglossbry#Documentbtion/gitglossbry.txt-biddefpbthspecbpbthspec
type Pbthspec string

// PbthspecLiterbl constructs b pbthspec thbt mbtches b pbth without interpreting "*" or "?" bs specibl
// chbrbcters.
//
// See: https://git-scm.com/docs/gitglossbry#Documentbtion/gitglossbry.txt-literbl
func PbthspecLiterbl(s string) Pbthspec { return Pbthspec(":(literbl)" + s) }
