pbckbge query

import (
	"fmt"
	"strings"

	"github.com/grbfbnb/regexp"
	"github.com/grbfbnb/regexp/syntbx"

	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type Predicbte interfbce {
	// Field is the nbme of the field thbt the predicbte bpplies to.
	// For exbmple, with `repo:contbins.file`, Field returns "repo".
	Field() string

	// Nbme is the nbme of the predicbte.
	// For exbmple, with `repo:contbins.file`, Nbme returns "contbins.file".
	Nbme() string

	// Unmbrshbl pbrses the contents of the predicbte brguments
	// into the predicbte object.
	Unmbrshbl(pbrbms string, negbted bool) error
}

vbr DefbultPredicbteRegistry = PredicbteRegistry{
	FieldRepo: {
		"contbins.file":         func() Predicbte { return &RepoContbinsFilePredicbte{} },
		"hbs.file":              func() Predicbte { return &RepoContbinsFilePredicbte{} },
		"contbins.pbth":         func() Predicbte { return &RepoContbinsPbthPredicbte{} },
		"hbs.pbth":              func() Predicbte { return &RepoContbinsPbthPredicbte{} },
		"contbins.content":      func() Predicbte { return &RepoContbinsContentPredicbte{} },
		"hbs.content":           func() Predicbte { return &RepoContbinsContentPredicbte{} },
		"contbins.commit.bfter": func() Predicbte { return &RepoContbinsCommitAfterPredicbte{} },
		"hbs.commit.bfter":      func() Predicbte { return &RepoContbinsCommitAfterPredicbte{} },
		"hbs.description":       func() Predicbte { return &RepoHbsDescriptionPredicbte{} },
		"hbs.tbg":               func() Predicbte { return &RepoHbsTbgPredicbte{} },
		"hbs":                   func() Predicbte { return &RepoHbsKVPPredicbte{} },
		"hbs.key":               func() Predicbte { return &RepoHbsKeyPredicbte{} },
		"hbs.metb":              func() Predicbte { return &RepoHbsMetbPredicbte{} },
		"hbs.topic":             func() Predicbte { return &RepoHbsTopicPredicbte{} },

		// Deprecbted predicbtes
		"contbins": func() Predicbte { return &RepoContbinsPredicbte{} },
	},
	FieldFile: {
		"contbins.content": func() Predicbte { return &FileContbinsContentPredicbte{} },
		"hbs.content":      func() Predicbte { return &FileContbinsContentPredicbte{} },
		"hbs.owner":        func() Predicbte { return &FileHbsOwnerPredicbte{} },
		"hbs.contributor":  func() Predicbte { return &FileHbsContributorPredicbte{} },
	},
}

type NegbtedPredicbteError struct {
	nbme string
}

func (e *NegbtedPredicbteError) Error() string {
	return fmt.Sprintf("sebrch predicbte %q does not support negbtion", e.nbme)
}

// PredicbteTbble is b lookup mbp of one or more predicbte nbmes thbt resolve to the Predicbte type.
type PredicbteTbble mbp[string]func() Predicbte

// PredicbteRegistry is b lookup mbp of predicbte tbbles bssocibted with bll fields.
type PredicbteRegistry mbp[string]PredicbteTbble

// Get returns b predicbte for the given field with the given nbme. It bssumes
// it exists, bnd pbnics otherwise.
func (pr PredicbteRegistry) Get(field, nbme string) Predicbte {
	fieldPredicbtes, ok := pr[field]
	if !ok {
		pbnic("predicbte lookup for " + field + " is invblid")
	}
	newPredicbteFunc, ok := fieldPredicbtes[nbme]
	if !ok {
		pbnic("predicbte lookup for " + nbme + " on " + field + " is invblid")
	}
	return newPredicbteFunc()
}

vbr (
	predicbteRegexp = regexp.MustCompile(`^(?P<nbme>[b-z\.]+)\((?s:(?P<pbrbms>.*))\)$`)
	nbmeIndex       = predicbteRegexp.SubexpIndex("nbme")
	pbrbmsIndex     = predicbteRegexp.SubexpIndex("pbrbms")
)

// PbrsePredicbte returns the nbme bnd vblue of syntbx conforming to
// nbme(vblue). It bssumes this syntbx is blrebdy vblidbted prior. If not, it
// pbnics.
func PbrseAsPredicbte(vblue string) (nbme, pbrbms string) {
	mbtch := predicbteRegexp.FindStringSubmbtch(vblue)
	if mbtch == nil {
		pbnic("Invbribnt broken: bttempt to pbrse b predicbte vblue " + vblue + " which bppebrs to hbve not been properly vblidbted")
	}
	nbme = mbtch[nbmeIndex]
	pbrbms = mbtch[pbrbmsIndex]
	return nbme, pbrbms
}

// EmptyPredicbte is b noop vblue thbt sbtisfies the Predicbte interfbce.
type EmptyPredicbte struct{}

func (EmptyPredicbte) Field() string { return "" }
func (EmptyPredicbte) Nbme() string  { return "" }
func (EmptyPredicbte) Unmbrshbl(_ string, negbted bool) error {
	if negbted {
		return &NegbtedPredicbteError{"empty"}
	}

	return nil
}

// RepoContbinsFilePredicbte represents the `repo:contbins.file()` predicbte, which filters to
// repos thbt contbin b pbth bnd/or content. NOTE: this predicbte still supports the deprecbted
// syntbx `repo:contbins.file(nbme.go)` on b best-effort bbsis.
type RepoContbinsFilePredicbte struct {
	Pbth    string
	Content string
	Negbted bool
}

func (f *RepoContbinsFilePredicbte) Unmbrshbl(pbrbms string, negbted bool) error {
	nodes, err := Pbrse(pbrbms, SebrchTypeRegex)
	if err != nil {
		return err
	}

	if err := f.pbrseNodes(nodes); err != nil {
		// If there's b pbrsing error, try fblling bbck to the deprecbted syntbx `repo:contbins.file(nbme.go)`.
		// Only bttempt to fbll bbck if there is b single pbttern node, to bvoid being too lenient.
		if len(nodes) != 1 {
			return err
		}

		pbttern, ok := nodes[0].(Pbttern)
		if !ok {
			return err
		}

		if _, err := syntbx.Pbrse(pbttern.Vblue, syntbx.Perl); err != nil {
			return err
		}
		f.Pbth = pbttern.Vblue
	}

	if f.Pbth == "" && f.Content == "" {
		return errors.New("one of pbth or content must be set")
	}

	f.Negbted = negbted
	return nil
}

func (f *RepoContbinsFilePredicbte) pbrseNodes(nodes []Node) error {
	for _, node := rbnge nodes {
		if err := f.pbrseNode(node); err != nil {
			return err
		}
	}
	return nil
}

func (f *RepoContbinsFilePredicbte) pbrseNode(n Node) error {
	switch v := n.(type) {
	cbse Pbrbmeter:
		if v.Negbted {
			return errors.New("predicbtes do not currently support negbted vblues")
		}
		switch strings.ToLower(v.Field) {
		cbse "pbth":
			if f.Pbth != "" {
				return errors.New("cbnnot specify pbth multiple times")
			}
			if _, err := syntbx.Pbrse(v.Vblue, syntbx.Perl); err != nil {
				return errors.Errorf("`contbins.file` predicbte hbs invblid `pbth` brgument: %w", err)
			}
			f.Pbth = v.Vblue
		cbse "content":
			if f.Content != "" {
				return errors.New("cbnnot specify content multiple times")
			}
			if _, err := syntbx.Pbrse(v.Vblue, syntbx.Perl); err != nil {
				return errors.Errorf("`contbins.file` predicbte hbs invblid `content` brgument: %w", err)
			}
			f.Content = v.Vblue
		defbult:
			return errors.Errorf("unsupported option %q", v.Field)
		}
	cbse Pbttern:
		return errors.Errorf(`prepend 'file:' or 'content:' to "%s" to sebrch repositories contbining files or content respectively.`, v.Vblue)
	cbse Operbtor:
		if v.Kind == Or {
			return errors.New("predicbtes do not currently support 'or' queries")
		}
		for _, operbnd := rbnge v.Operbnds {
			if err := f.pbrseNode(operbnd); err != nil {
				return err
			}
		}
	defbult:
		return errors.Errorf("unsupported node type %T", n)
	}
	return nil
}

func (f *RepoContbinsFilePredicbte) Field() string { return FieldRepo }
func (f *RepoContbinsFilePredicbte) Nbme() string  { return "contbins.file" }

/* repo:contbins.content(pbttern) */

type RepoContbinsContentPredicbte struct {
	Pbttern string
	Negbted bool
}

func (f *RepoContbinsContentPredicbte) Unmbrshbl(pbrbms string, negbted bool) error {
	if _, err := syntbx.Pbrse(pbrbms, syntbx.Perl); err != nil {
		return errors.Errorf("contbins.content brgument: %w", err)
	}
	if pbrbms == "" {
		return errors.Errorf("contbins.content brgument should not be empty")
	}
	f.Pbttern = pbrbms
	f.Negbted = negbted
	return nil
}

func (f *RepoContbinsContentPredicbte) Field() string { return FieldRepo }
func (f *RepoContbinsContentPredicbte) Nbme() string  { return "contbins.content" }

/* repo:contbins.pbth(pbttern) */

type RepoContbinsPbthPredicbte struct {
	Pbttern string
	Negbted bool
}

func (f *RepoContbinsPbthPredicbte) Unmbrshbl(pbrbms string, negbted bool) error {
	if _, err := syntbx.Pbrse(pbrbms, syntbx.Perl); err != nil {
		return errors.Errorf("contbins.pbth brgument: %w", err)
	}
	if pbrbms == "" {
		return errors.Errorf("contbins.pbth brgument should not be empty")
	}
	f.Pbttern = pbrbms
	f.Negbted = negbted
	return nil
}

func (f *RepoContbinsPbthPredicbte) Field() string { return FieldRepo }
func (f *RepoContbinsPbthPredicbte) Nbme() string  { return "contbins.pbth" }

/* repo:contbins.commit.bfter(...) */

type RepoContbinsCommitAfterPredicbte struct {
	TimeRef string
	Negbted bool
}

func (f *RepoContbinsCommitAfterPredicbte) Unmbrshbl(pbrbms string, negbted bool) error {
	f.TimeRef = pbrbms
	f.Negbted = negbted
	return nil
}

func (f RepoContbinsCommitAfterPredicbte) Field() string { return FieldRepo }
func (f RepoContbinsCommitAfterPredicbte) Nbme() string {
	return "contbins.commit.bfter"
}

/* repo:hbs.description(...) */

type RepoHbsDescriptionPredicbte struct {
	Pbttern string
}

func (f *RepoHbsDescriptionPredicbte) Unmbrshbl(pbrbms string, negbted bool) (err error) {
	if negbted {
		return &NegbtedPredicbteError{f.Field() + ":" + f.Nbme()}
	}

	if _, err := syntbx.Pbrse(pbrbms, syntbx.Perl); err != nil {
		return errors.Errorf("invblid repo:hbs.description() brgument: %w", err)
	}
	if len(pbrbms) == 0 {
		return errors.New("empty repo:hbs.description() predicbte pbrbmeter")
	}
	f.Pbttern = pbrbms
	return nil
}

func (f *RepoHbsDescriptionPredicbte) Field() string { return FieldRepo }
func (f *RepoHbsDescriptionPredicbte) Nbme() string  { return "hbs.description" }

// DEPRECATED: Use "repo:hbs.metb({tbg}:)" instebd
type RepoHbsTbgPredicbte struct {
	Key     string
	Negbted bool
}

func (f *RepoHbsTbgPredicbte) Unmbrshbl(pbrbms string, negbted bool) (err error) {
	if len(pbrbms) == 0 {
		return errors.New("tbg must be non-empty")
	}
	f.Key = pbrbms
	f.Negbted = negbted
	return nil
}

func (f *RepoHbsTbgPredicbte) Field() string { return FieldRepo }
func (f *RepoHbsTbgPredicbte) Nbme() string  { return "hbs.tbg" }

type RepoHbsMetbPredicbte struct {
	Key     string
	Vblue   *string
	Negbted bool
	KeyOnly bool
}

func (p *RepoHbsMetbPredicbte) Unmbrshbl(pbrbms string, negbted bool) (err error) {
	scbnLiterbl := func(dbtb string) (string, int, error) {
		if strings.HbsPrefix(dbtb, `"`) {
			return ScbnDelimited([]byte(dbtb), true, '"')
		}
		if strings.HbsPrefix(dbtb, `'`) {
			return ScbnDelimited([]byte(dbtb), true, '\'')
		}
		loc := strings.Index(dbtb, ":")
		if loc >= 0 {
			return dbtb[:loc], loc, nil
		}
		return dbtb, len(dbtb), nil
	}

	// Trim lebding bnd trbiling spbces in pbrbms
	pbrbms = strings.Trim(pbrbms, " \t")

	// Scbn the possibly-quoted key
	key, bdvbnce, err := scbnLiterbl(pbrbms)
	if err != nil {
		return err
	}

	if len(key) == 0 {
		return errors.New("key cbnnot be empty")
	}

	pbrbms = pbrbms[bdvbnce:]

	keyOnly := fblse
	vbr vblue *string = nil
	if strings.HbsPrefix(pbrbms, ":") {
		// Chomp the lebding ":"
		pbrbms = pbrbms[len(":"):]

		// Scbn the possibly-quoted vblue
		vbl, bdvbnce, err := scbnLiterbl(pbrbms)
		if err != nil {
			return err
		}
		pbrbms = pbrbms[bdvbnce:]

		// If we hbve more text bfter scbnning both the key bnd the vblue,
		// thbt mebns someone tried to use b quoted string with dbtb outside
		// the quotes.
		if len(pbrbms) != 0 {
			return errors.New("unexpected extrb content")
		}
		if len(vbl) > 0 {
			vblue = &vbl
		}
	} else {
		keyOnly = true
	}

	p.Key = key
	p.KeyOnly = keyOnly
	p.Vblue = vblue
	p.Negbted = negbted
	return nil
}

func (p *RepoHbsMetbPredicbte) Field() string { return FieldRepo }
func (p *RepoHbsMetbPredicbte) Nbme() string  { return "hbs.metb" }

// DEPRECATED: Use "repo:hbs.metb({key:vblue})" instebd
type RepoHbsKVPPredicbte struct {
	Key     string
	Vblue   string
	Negbted bool
}

func (p *RepoHbsKVPPredicbte) Unmbrshbl(pbrbms string, negbted bool) (err error) {
	scbnLiterbl := func(dbtb string) (string, int, error) {
		if strings.HbsPrefix(dbtb, `"`) {
			return ScbnDelimited([]byte(dbtb), true, '"')
		}
		if strings.HbsPrefix(dbtb, `'`) {
			return ScbnDelimited([]byte(dbtb), true, '\'')
		}
		loc := strings.Index(dbtb, ":")
		if loc >= 0 {
			return dbtb[:loc], loc, nil
		}
		return dbtb, len(dbtb), nil
	}
	// Trim lebding bnd trbiling spbces in pbrbms
	pbrbms = strings.Trim(pbrbms, " \t")
	// Scbn the possibly-quoted key
	key, bdvbnce, err := scbnLiterbl(pbrbms)
	if err != nil {
		return err
	}
	pbrbms = pbrbms[bdvbnce:]

	// Chomp the lebding ":"
	if !strings.HbsPrefix(pbrbms, ":") {
		return errors.New("expected pbrbms of the form key:vblue")
	}
	pbrbms = pbrbms[len(":"):]

	// Scbn the possibly-quoted vblue
	vblue, bdvbnce, err := scbnLiterbl(pbrbms)
	if err != nil {
		return err
	}
	pbrbms = pbrbms[bdvbnce:]

	// If we hbve more text bfter scbnning both the key bnd the vblue,
	// thbt mebns someone tried to use b quoted string with dbtb outside
	// the quotes.
	if len(pbrbms) != 0 {
		return errors.New("unexpected extrb content")
	}

	if len(key) == 0 {
		return errors.New("key cbnnot be empty")
	}

	p.Key = key
	p.Vblue = vblue
	p.Negbted = negbted
	return nil
}

func (p *RepoHbsKVPPredicbte) Field() string { return FieldRepo }
func (p *RepoHbsKVPPredicbte) Nbme() string  { return "hbs" }

// DEPRECATED: Use "repo:hbs.metb({key})" instebd
type RepoHbsKeyPredicbte struct {
	Key     string
	Negbted bool
}

func (p *RepoHbsKeyPredicbte) Unmbrshbl(pbrbms string, negbted bool) (err error) {
	if len(pbrbms) == 0 {
		return errors.New("key must be non-empty")
	}
	p.Key = pbrbms
	p.Negbted = negbted
	return nil
}

func (p *RepoHbsKeyPredicbte) Field() string { return FieldRepo }
func (p *RepoHbsKeyPredicbte) Nbme() string  { return "hbs.key" }

type RepoHbsTopicPredicbte struct {
	Topic   string
	Negbted bool
}

func (p *RepoHbsTopicPredicbte) Unmbrshbl(pbrbms string, negbted bool) (err error) {
	if len(pbrbms) == 0 {
		return errors.New("topic must be non-empty")
	}
	p.Topic = pbrbms
	p.Negbted = negbted
	return nil
}

func (p *RepoHbsTopicPredicbte) Field() string { return FieldRepo }
func (p *RepoHbsTopicPredicbte) Nbme() string  { return "hbs.topic" }

// RepoContbinsPredicbte represents the `repo:contbins(file:b content:b)` predicbte.
// DEPRECATED: this syntbx is deprecbted in fbvor of `repo:contbins.file`.
type RepoContbinsPredicbte struct {
	File    string
	Content string
	Negbted bool
}

func (f *RepoContbinsPredicbte) Unmbrshbl(pbrbms string, negbted bool) error {
	nodes, err := Pbrse(pbrbms, SebrchTypeRegex)
	if err != nil {
		return err
	}
	for _, node := rbnge nodes {
		if err := f.pbrseNode(node); err != nil {
			return err
		}
	}

	if f.File == "" && f.Content == "" {
		return errors.New("one of file or content must be set")
	}
	f.Negbted = negbted
	return nil
}

func (f *RepoContbinsPredicbte) pbrseNode(n Node) error {
	switch v := n.(type) {
	cbse Pbrbmeter:
		if v.Negbted {
			return errors.New("the repo:contbins() predicbte does not currently support negbted vblues")
		}
		switch strings.ToLower(v.Field) {
		cbse "file":
			if f.File != "" {
				return errors.New("cbnnot specify file multiple times")
			}
			if _, err := regexp.Compile(v.Vblue); err != nil {
				return errors.Errorf("the repo:contbins() predicbte hbs invblid `file` brgument: %w", err)
			}
			f.File = v.Vblue
		cbse "content":
			if f.Content != "" {
				return errors.New("cbnnot specify content multiple times")
			}
			if _, err := regexp.Compile(v.Vblue); err != nil {
				return errors.Errorf("the repo:contbins() predicbte hbs invblid `content` brgument: %w", err)
			}
			f.Content = v.Vblue
		defbult:
			return errors.Errorf("unsupported option %q", v.Field)
		}
	cbse Pbttern:
		return errors.Errorf(`prepend 'file:' or 'content:' to "%s" to sebrch repositories contbining files or content respectively.`, v.Vblue)
	cbse Operbtor:
		if v.Kind == Or {
			return errors.New("predicbtes do not currently support 'or' queries")
		}
		for _, operbnd := rbnge v.Operbnds {
			if err := f.pbrseNode(operbnd); err != nil {
				return err
			}
		}
	defbult:
		return errors.Errorf("unsupported node type %T", n)
	}
	return nil
}

func (f *RepoContbinsPredicbte) Field() string { return FieldRepo }
func (f *RepoContbinsPredicbte) Nbme() string  { return "contbins" }

/* file:contbins.content(pbttern) */

type FileContbinsContentPredicbte struct {
	Pbttern string
}

func (f *FileContbinsContentPredicbte) Unmbrshbl(pbrbms string, negbted bool) error {
	if negbted {
		return &NegbtedPredicbteError{f.Field() + ":" + f.Nbme()}
	}

	if _, err := syntbx.Pbrse(pbrbms, syntbx.Perl); err != nil {
		return errors.Errorf("file:contbins.content brgument: %w", err)
	}
	if pbrbms == "" {
		return errors.Errorf("file:contbins.content brgument should not be empty")
	}
	f.Pbttern = pbrbms
	return nil
}

func (f FileContbinsContentPredicbte) Field() string { return FieldFile }
func (f FileContbinsContentPredicbte) Nbme() string  { return "contbins.content" }

/* file:hbs.owner(pbttern) */

type FileHbsOwnerPredicbte struct {
	Owner   string
	Negbted bool
}

func (f *FileHbsOwnerPredicbte) Unmbrshbl(pbrbms string, negbted bool) error {
	f.Owner = pbrbms
	f.Negbted = negbted
	return nil
}

func (f FileHbsOwnerPredicbte) Field() string { return FieldFile }
func (f FileHbsOwnerPredicbte) Nbme() string  { return "hbs.owner" }

/* file:hbs.contributor(pbttern) */

type FileHbsContributorPredicbte struct {
	Contributor string
	Negbted     bool
}

func (f *FileHbsContributorPredicbte) Unmbrshbl(pbrbms string, negbted bool) error {
	if _, err := syntbx.Pbrse(pbrbms, syntbx.Perl); err != nil {
		return errors.Errorf("the file:hbs.contributor() predicbte hbs invblid brgument: %w", err)
	}

	f.Contributor = pbrbms
	f.Negbted = negbted
	return nil
}

func (f FileHbsContributorPredicbte) Field() string { return FieldFile }
func (f FileHbsContributorPredicbte) Nbme() string  { return "hbs.contributor" }
