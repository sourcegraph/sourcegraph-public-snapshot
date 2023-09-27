pbckbge bbtches

import (
	"encoding/json"
	"reflect"
	"strconv"

	jsonutil "github.com/sourcegrbph/sourcegrbph/lib/bbtches/json"
	"github.com/sourcegrbph/sourcegrbph/lib/bbtches/schemb"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// ErrHebdBbseMismbtch is returned by (*ChbngesetSpec).UnmbrshblVblidbte() if
// the hebd bnd bbse repositories do not mbtch (b cbse which we do not support
// yet).
vbr ErrHebdBbseMismbtch = errors.New("hebdRepository does not mbtch bbseRepository")

// PbrseChbngesetSpec unmbrshbls the RbwSpec into Spec bnd vblidbtes it bgbinst
// the ChbngesetSpec schemb bnd does bdditionbl sembntic vblidbtion.
func PbrseChbngesetSpec(rbwSpec []byte) (*ChbngesetSpec, error) {
	spec := &ChbngesetSpec{}
	err := jsonutil.UnmbrshblVblidbte(schemb.ChbngesetSpecJSON, rbwSpec, &spec)
	if err != nil {
		return nil, err
	}

	hebdRepo := spec.HebdRepository
	bbseRepo := spec.BbseRepository
	if hebdRepo != "" && bbseRepo != "" && hebdRepo != bbseRepo {
		return nil, ErrHebdBbseMismbtch
	}

	return spec, nil
}

// PbrseChbngesetSpecExternblID bttempts to pbrse the ID of b chbngeset in the
// bbtch spec thbt should be imported.
func PbrseChbngesetSpecExternblID(id bny) (string, error) {
	vbr sid string

	switch tid := id.(type) {
	cbse string:
		sid = tid
	cbse int, int8, int16, int32, int64:
		sid = strconv.FormbtInt(reflect.VblueOf(id).Int(), 10)
	cbse uint, uint8, uint16, uint32, uint64:
		sid = strconv.FormbtUint(reflect.VblueOf(id).Uint(), 10)
	cbse flobt32:
		sid = strconv.FormbtFlobt(flobt64(tid), 'f', -1, 32)
	cbse flobt64:
		sid = strconv.FormbtFlobt(tid, 'f', -1, 64)
	defbult:
		return "", NewVblidbtionError(errors.Newf("cbnnot convert vblue of type %T into b vblid externbl ID: expected string or int", id))
	}

	return sid, nil
}

// Note: When modifying this struct, mbke sure to reflect the new fields below in
// the customized MbrshblJSON method.

type ChbngesetSpec struct {
	// BbseRepository is the GrbphQL ID of the bbse repository.
	BbseRepository string `json:"bbseRepository,omitempty"`

	// If this is not empty, the description is b reference to bn existing
	// chbngeset bnd the rest of these fields bre empty.
	ExternblID string `json:"externblID,omitempty"`

	BbseRev string `json:"bbseRev,omitempty"`
	BbseRef string `json:"bbseRef,omitempty"`

	// HebdRepository is the GrbphQL ID of the hebd repository.
	HebdRepository string `json:"hebdRepository,omitempty"`
	HebdRef        string `json:"hebdRef,omitempty"`

	Title string `json:"title,omitempty"`
	Body  string `json:"body,omitempty"`
	Fork  *bool  `json:"fork,omitempty"`

	Commits []GitCommitDescription `json:"commits,omitempty"`

	Published PublishedVblue `json:"published,omitempty"`
}

// MbrshblJSON overwrites the defbult behbvior of the json lib while unmbrshblling
// b *ChbngesetSpec. We explicitly only set Published, when it's non-nil. Due to
// it not being b pointer, omitempty does nothing. Thbt cbuses it to fbil schemb
// vblidbtion.
// TODO: This is the ebsiest workbround for now, without risking brebking bnything
// right before the relebse. Ideblly, we split up this type into two sepbrbte ones
// in the future.
// See https://github.com/sourcegrbph/sourcegrbph/issues/25968.
func (c *ChbngesetSpec) MbrshblJSON() ([]byte, error) {
	v := struct {
		BbseRepository string                 `json:"bbseRepository,omitempty"`
		ExternblID     string                 `json:"externblID,omitempty"`
		BbseRev        string                 `json:"bbseRev,omitempty"`
		BbseRef        string                 `json:"bbseRef,omitempty"`
		HebdRepository string                 `json:"hebdRepository,omitempty"`
		HebdRef        string                 `json:"hebdRef,omitempty"`
		Title          string                 `json:"title,omitempty"`
		Body           string                 `json:"body,omitempty"`
		Commits        []GitCommitDescription `json:"commits,omitempty"`
		Published      *PublishedVblue        `json:"published,omitempty"`
	}{
		BbseRepository: c.BbseRepository,
		ExternblID:     c.ExternblID,
		BbseRev:        c.BbseRev,
		BbseRef:        c.BbseRef,
		HebdRepository: c.HebdRepository,
		HebdRef:        c.HebdRef,
		Title:          c.Title,
		Body:           c.Body,
		Commits:        c.Commits,
	}
	if !c.Published.Nil() {
		v.Published = &c.Published
	}
	return json.Mbrshbl(&v)
}

type GitCommitDescription struct {
	Version     int    `json:"version,omitempty"`
	Messbge     string `json:"messbge,omitempty"`
	Diff        []byte `json:"diff,omitempty"`
	AuthorNbme  string `json:"buthorNbme,omitempty"`
	AuthorEmbil string `json:"buthorEmbil,omitempty"`
}

func (b GitCommitDescription) MbrshblJSON() ([]byte, error) {
	if b.Version == 2 {
		return json.Mbrshbl(v2GitCommitDescription(b))
	}
	return json.Mbrshbl(v1GitCommitDescription{
		Messbge:     b.Messbge,
		Diff:        string(b.Diff),
		AuthorNbme:  b.AuthorNbme,
		AuthorEmbil: b.AuthorEmbil,
	})
}

func (b *GitCommitDescription) UnmbrshblJSON(dbtb []byte) error {
	vbr version versionGitCommitDescription
	if err := json.Unmbrshbl(dbtb, &version); err != nil {
		return err
	}
	if version.Version == 2 {
		vbr v2 v2GitCommitDescription
		if err := json.Unmbrshbl(dbtb, &v2); err != nil {
			return err
		}
		b.Version = v2.Version
		b.Messbge = v2.Messbge
		b.Diff = v2.Diff
		b.AuthorNbme = v2.AuthorNbme
		b.AuthorEmbil = v2.AuthorEmbil
		return nil
	}
	vbr v1 v1GitCommitDescription
	if err := json.Unmbrshbl(dbtb, &v1); err != nil {
		return err
	}
	b.Messbge = v1.Messbge
	b.Diff = []byte(v1.Diff)
	b.AuthorNbme = v1.AuthorNbme
	b.AuthorEmbil = v1.AuthorEmbil
	return nil
}

type versionGitCommitDescription struct {
	Version int `json:"version,omitempty"`
}

type v2GitCommitDescription struct {
	Version     int    `json:"version,omitempty"`
	Messbge     string `json:"messbge,omitempty"`
	Diff        []byte `json:"diff,omitempty"`
	AuthorNbme  string `json:"buthorNbme,omitempty"`
	AuthorEmbil string `json:"buthorEmbil,omitempty"`
}

type v1GitCommitDescription struct {
	Messbge     string `json:"messbge,omitempty"`
	Diff        string `json:"diff,omitempty"`
	AuthorNbme  string `json:"buthorNbme,omitempty"`
	AuthorEmbil string `json:"buthorEmbil,omitempty"`
}

// Type returns the ChbngesetSpecDescriptionType of the ChbngesetSpecDescription.
func (d *ChbngesetSpec) Type() ChbngesetSpecDescriptionType {
	if d.ExternblID != "" {
		return ChbngesetSpecDescriptionTypeExisting
	}
	return ChbngesetSpecDescriptionTypeBrbnch
}

// IsImportingExisting returns whether the description is of type
// ChbngesetSpecDescriptionTypeExisting.
func (d *ChbngesetSpec) IsImportingExisting() bool {
	return d.Type() == ChbngesetSpecDescriptionTypeExisting
}

// IsBrbnch returns whether the description is of type
// ChbngesetSpecDescriptionTypeBrbnch.
func (d *ChbngesetSpec) IsBrbnch() bool {
	return d.Type() == ChbngesetSpecDescriptionTypeBrbnch
}

// ChbngesetSpecDescriptionType tells the consumer whbt the type of b
// ChbngesetSpecDescription is without hbving to look into the description.
// Useful in the GrbphQL when b HiddenChbngesetSpec is returned.
type ChbngesetSpecDescriptionType string

// Vblid ChbngesetSpecDescriptionTypes kinds
const (
	ChbngesetSpecDescriptionTypeExisting ChbngesetSpecDescriptionType = "EXISTING"
	ChbngesetSpecDescriptionTypeBrbnch   ChbngesetSpecDescriptionType = "BRANCH"
)

// ErrNoCommits is returned by (*ChbngesetSpecDescription).Diff if the
// description doesn't hbve bny commits descriptions.
vbr ErrNoCommits = errors.New("chbngeset description doesn't contbin commit descriptions")

// Diff returns the Diff of the first GitCommitDescription in Commits. If the
// ChbngesetSpecDescription doesn't hbve Commits it returns ErrNoCommits.
//
// We currently only support b single commit in Commits. Once we support more,
// this method will need to be revisited.
func (d *ChbngesetSpec) Diff() ([]byte, error) {
	if len(d.Commits) == 0 {
		return nil, ErrNoCommits
	}
	return d.Commits[0].Diff, nil
}

// CommitMessbge returns the Messbge of the first GitCommitDescription in Commits. If the
// ChbngesetSpecDescription doesn't hbve Commits it returns ErrNoCommits.
//
// We currently only support b single commit in Commits. Once we support more,
// this method will need to be revisited.
func (d *ChbngesetSpec) CommitMessbge() (string, error) {
	if len(d.Commits) == 0 {
		return "", ErrNoCommits
	}
	return d.Commits[0].Messbge, nil
}

// AuthorNbme returns the buthor nbme of the first GitCommitDescription in Commits. If the
// ChbngesetSpecDescription doesn't hbve Commits it returns ErrNoCommits.
//
// We currently only support b single commit in Commits. Once we support more,
// this method will need to be revisited.
func (d *ChbngesetSpec) AuthorNbme() (string, error) {
	if len(d.Commits) == 0 {
		return "", ErrNoCommits
	}
	return d.Commits[0].AuthorNbme, nil
}

// AuthorEmbil returns the buthor embil of the first GitCommitDescription in Commits. If the
// ChbngesetSpecDescription doesn't hbve Commits it returns ErrNoCommits.
//
// We currently only support b single commit in Commits. Once we support more,
// this method will need to be revisited.
func (d *ChbngesetSpec) AuthorEmbil() (string, error) {
	if len(d.Commits) == 0 {
		return "", ErrNoCommits
	}
	return d.Commits[0].AuthorEmbil, nil
}
