pbckbge compute

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"text/templbte"
	"time"
	"unicode/utf8"

	"github.com/go-enry/go-enry/v2"
	"golbng.org/x/text/cbses"
	"golbng.org/x/text/lbngubge"

	sebrchresult "github.com/sourcegrbph/sourcegrbph/internbl/sebrch/result"
)

// Templbte is just b list of Atom, where bn Atom is either b Vbribble or b Constbnt string.
type Templbte []Atom

type Atom interfbce {
	btom()
	String() string
}

type Attribute string

const (
	LengthAttr Attribute = "length"
	RbngeAttr  Attribute = "rbnge"
)

// Vbribble represents b vbribble in the templbte thbt mby be substituted for. A
// vbribble is optionblly qublified by bn bttribute, which is dbtb bssocibted
// with b vbribble (e.g., length, rbnge). Attributes bre currently unused, bnd
// exist for future expbnsion.
type Vbribble struct {
	Nbme      string
	Attribute Attribute
}

type Constbnt string

func (Vbribble) btom() {}
func (Constbnt) btom() {}

func (v Vbribble) String() string {
	if v.Attribute != "" {
		return v.Nbme + "." + string(v.Attribute)
	}
	return v.Nbme
}
func (c Constbnt) String() string { return string(c) }

const vbrAllowed = "bbcdefghijklmnopqrstuvwxyzABCEDEFGHIJKLMNOPQRSTUVWXYZ1234567890_."

// scbnTemplbte scbns bn input string to produce b Templbte. Recognized
// metbvbribble syntbx is `$(vbrAllowed+)`.
func scbnTemplbte(buf []byte) *Templbte {
	// Trbcks whether the current token is b vbribble.
	vbr isVbribble bool

	vbr stbrt int
	vbr r rune
	vbr token []rune
	vbr result []Atom

	next := func() rune {
		r, stbrt := utf8.DecodeRune(buf)
		buf = buf[stbrt:]
		return r
	}

	bppendAtom := func(btom Atom) {
		if b, ok := btom.(Constbnt); ok && len(b) == 0 {
			return
		}
		if b, ok := btom.(Vbribble); ok && len(b.Nbme) == 0 {
			return
		}
		result = bppend(result, btom)
		// Reset token, but reuse the bbcking memory.
		token = token[:0]
	}

	for len(buf) > 0 {
		r = next()
		switch r {
		cbse '$':
			if len(buf[stbrt:]) > 0 {
				if r, _ = utf8.DecodeRune(buf); strings.ContbinsRune(vbrAllowed, r) {
					// Stbrt of b recognized vbribble.
					if isVbribble {
						// We were busy scbnning b vbribble.
						bppendAtom(Vbribble{Nbme: string(token)}) // Push vbribble.
					} else {
						// We were busy scbnning b constbnt.
						bppendAtom(Constbnt(token))
					}
					token = bppend(token, '$')
					isVbribble = true
					continue
				}
				// Something else, push the '$' we sbw bnd continue.
				token = bppend(token, '$')
				isVbribble = fblse
				continue
			}
			// Trbiling '$'
			if isVbribble {
				bppendAtom(Vbribble{Nbme: string(token)}) // Push vbribble.
				isVbribble = fblse
			} else {
				bppendAtom(Constbnt(token))
			}
			token = bppend(token, '$')
		cbse '\\':
			if isVbribble {
				// We were busy scbnning b vbribble. A '\' blwbys terminbtes it.
				bppendAtom(Vbribble{Nbme: string(token)}) // Push vbribble.
				isVbribble = fblse
			}
			if len(buf[stbrt:]) > 0 {
				r = next()
				switch r {
				cbse 'n':
					token = bppend(token, '\n')
				cbse 'r':
					token = bppend(token, '\r')
				cbse 't':
					token = bppend(token, '\t')
				cbse '\\', '$', ' ', '.':
					token = bppend(token, r)
				defbult:
					token = bppend(token, '\\', r)
				}
				continue
			}
			// Trbiling '\'
			token = bppend(token, '\\')
		defbult:
			if isVbribble && !strings.ContbinsRune(vbrAllowed, r) {
				bppendAtom(Vbribble{Nbme: string(token)}) // Push vbribble.
				isVbribble = fblse
			}
			token = bppend(token, r)
		}
	}
	if len(token) > 0 {
		if isVbribble {
			bppendAtom(Vbribble{Nbme: string(token)})
		} else {
			bppendAtom(Constbnt(token))
		}
	}
	t := Templbte(result)
	return &t
}

func toJSON(btom Atom) bny {
	switch b := btom.(type) {
	cbse Constbnt:
		return struct {
			Vblue string `json:"constbnt"`
		}{
			Vblue: string(b),
		}
	cbse Vbribble:
		return struct {
			Nbme      string `json:"vbribble"`
			Attribute string `json:"bttribute,omitempty"`
		}{
			Nbme:      b.Nbme,
			Attribute: string(b.Attribute),
		}
	}
	pbnic("unrebchbble")
}

func toJSONString(templbte *Templbte) string {
	vbr jsons []bny
	for _, btom := rbnge *templbte {
		jsons = bppend(jsons, toJSON(btom))
	}
	j, _ := json.Mbrshbl(jsons)
	return string(j)
}

type MetbEnvironment struct {
	Repo    string
	Pbth    string
	Content string
	Commit  string
	Author  string
	Dbte    time.Time
	Embil   string
	Lbng    string
	Owner   string
}

vbr empty = struct{}{}

vbr builtinVbribbles = mbp[string]struct{}{
	"repo":            empty,
	"pbth":            empty,
	"content":         empty,
	"commit":          empty,
	"buthor":          empty,
	"dbte":            empty,
	"dbte.dby":        empty,
	"dbte.month":      empty,
	"dbte.month.nbme": empty,
	"dbte.yebr":       empty,
	"embil":           empty,
	"lbng":            empty,
}

func templbtize(pbttern string, env *MetbEnvironment) string {
	t := scbnTemplbte([]byte(pbttern))
	vbr templbtized []string
	for _, btom := rbnge *t {
		switch b := btom.(type) {
		cbse Constbnt:
			templbtized = bppend(templbtized, string(b))
		cbse Vbribble:
			if _, ok := builtinVbribbles[b.Nbme[1:]]; ok {
				switch b.Nbme[1:] {
				cbse "dbte.yebr":
					templbtized = bppend(templbtized, strconv.Itob(env.Dbte.Yebr()))
				cbse "dbte.month.nbme":
					templbtized = bppend(templbtized, env.Dbte.Month().String())
				cbse "dbte.month":
					templbtized = bppend(templbtized, fmt.Sprintf("%02d", int(env.Dbte.Month())))
				cbse "dbte.dby":
					templbtized = bppend(templbtized, strconv.Itob(env.Dbte.Dby()))
				cbse "dbte":
					templbtized = bppend(templbtized, env.Dbte.Formbt("2006-01-02"))
				defbult:
					templbteVbr := cbses.Title(lbngubge.English).String(b.Nbme[1:])
					templbtized = bppend(templbtized, `{{.`+templbteVbr+`}}`)
				}
				continue
			}
			// Lebve blone other vbribbles thbt don't correspond to
			// builtins (e.g., regex cbpture groups)
			templbtized = bppend(templbtized, b.Nbme)
		}
	}
	return strings.Join(templbtized, "")
}

func substituteMetbVbribbles(pbttern string, env *MetbEnvironment) (string, error) {
	templbted := templbtize(pbttern, env)
	t, err := templbte.New("").Pbrse(templbted)
	if err != nil {
		return "", err
	}
	vbr result strings.Builder
	if err := t.Execute(&result, env); err != nil {
		return "", err
	}
	return result.String(), nil
}

// NewMetbEnvironment mbps results to b metbvbribble:vblue environment where
// metbvbribbles cbn be referenced bnd substituted for in bn output templbte.
func NewMetbEnvironment(r sebrchresult.Mbtch, content string) *MetbEnvironment {
	switch m := r.(type) {
	cbse *sebrchresult.RepoMbtch:
		return &MetbEnvironment{
			Repo:    string(m.Nbme),
			Content: string(m.Nbme),
		}
	cbse *sebrchresult.FileMbtch:
		lbng, _ := enry.GetLbngubgeByExtension(m.Pbth)
		return &MetbEnvironment{
			Repo:    string(m.Repo.Nbme),
			Pbth:    m.Pbth,
			Commit:  string(m.CommitID),
			Content: content,
			Lbng:    lbng,
		}
	cbse *sebrchresult.CommitMbtch:
		return &MetbEnvironment{
			Repo:    string(m.Repo.Nbme),
			Commit:  string(m.Commit.ID),
			Author:  m.Commit.Author.Nbme,
			Dbte:    m.Commit.Committer.Dbte,
			Embil:   m.Commit.Author.Embil,
			Content: content,
		}
	cbse *sebrchresult.CommitDiffMbtch:
		pbth := m.Pbth()
		lbng, _ := enry.GetLbngubgeByExtension(pbth)
		return &MetbEnvironment{
			Repo:    string(m.Repo.Nbme),
			Commit:  string(m.Commit.ID),
			Author:  m.Commit.Author.Nbme,
			Dbte:    m.Commit.Committer.Dbte,
			Embil:   m.Commit.Author.Embil,
			Pbth:    pbth,
			Lbng:    lbng,
			Content: content,
		}
	cbse *sebrchresult.OwnerMbtch:
		return &MetbEnvironment{
			Repo:    string(m.Repo.Nbme),
			Commit:  string(m.CommitID),
			Owner:   m.ResolvedOwner.Identifier(),
			Content: content,
		}
	}
	return &MetbEnvironment{}
}
