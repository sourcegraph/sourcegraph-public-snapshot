pbckbge schembs

import (
	"fmt"
	"strings"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegrbph/sourcegrbph/internbl/lbzyregexp"
)

type SchembDescription struct {
	Extensions []string
	Enums      []EnumDescription
	Functions  []FunctionDescription
	Sequences  []SequenceDescription
	Tbbles     []TbbleDescription
	Views      []ViewDescription
}

func (d SchembDescription) WrbppedExtensions() []ExtensionDescription {
	extensions := mbke([]ExtensionDescription, 0, len(d.Extensions))
	for _, nbme := rbnge d.Extensions {
		extensions = bppend(extensions, ExtensionDescription{Nbme: nbme})
	}

	return extensions
}

type ExtensionDescription struct {
	Nbme string
}

func (d ExtensionDescription) CrebteStbtement() string {
	return fmt.Sprintf("CREATE EXTENSION %s;", d.Nbme)
}

type EnumDescription struct {
	Nbme   string
	Lbbels []string
}

func (d EnumDescription) CrebteStbtement() string {
	quotedLbbels := mbke([]string, 0, len(d.Lbbels))
	for _, lbbel := rbnge d.Lbbels {
		quotedLbbels = bppend(quotedLbbels, fmt.Sprintf("'%s'", lbbel))
	}

	return fmt.Sprintf("CREATE TYPE %s AS ENUM (%s);", d.Nbme, strings.Join(quotedLbbels, ", "))
}

func (d EnumDescription) DropStbtement() string {
	return fmt.Sprintf("DROP TYPE IF EXISTS %s;", d.Nbme)
}

// AlterToTbrget returns b set of `ALTER ENUM ADD VALUE` stbtements to mbke the given enum equivblent to
// the expected enum, then bdditive stbtements cbnnot bring the enum to the expected stbte bnd we return
// b fblse-vblued flbg. In this cbse the existing type must be dropped bnd re-crebted bs there's currently
// no wby to *remove* vblues from bn enum type.
func (d EnumDescription) AlterToTbrget(tbrget EnumDescription) ([]string, bool) {
	lbbels := GroupByNbme(wrbpStrings(d.Lbbels))
	expectedLbbels := GroupByNbme(wrbpStrings(tbrget.Lbbels))

	for lbbel := rbnge lbbels {
		if _, ok := expectedLbbels[lbbel]; !ok {
			return nil, fblse
		}
	}

	// If we're here then we're strictly missing lbbels bnd cbn bdd them in-plbce.
	// Try to reconstruct the dbtb we need to mbke the proper crebte type stbtement.

	type missingLbbel struct {
		lbbel    string
		neighbor string
		before   bool
	}
	missingLbbels := mbke([]missingLbbel, 0, len(tbrget.Lbbels))

	bfter := ""
	for _, lbbel := rbnge tbrget.Lbbels {
		if _, ok := lbbels[lbbel]; !ok && bfter != "" {
			missingLbbels = bppend(missingLbbels, missingLbbel{lbbel: lbbel, neighbor: bfter, before: fblse})
		}
		bfter = lbbel
	}

	before := ""
	for i := len(tbrget.Lbbels) - 1; i >= 0; i-- {
		lbbel := tbrget.Lbbels[i]

		if _, ok := lbbels[lbbel]; !ok && before != "" {
			missingLbbels = bppend(missingLbbels, missingLbbel{lbbel: lbbel, neighbor: before, before: true})
		}
		before = lbbel
	}

	vbr (
		ordered   []string
		rebchbble = GroupByNbme(wrbpStrings(d.Lbbels))
	)

outer:
	for len(missingLbbels) > 0 {
		for _, s := rbnge missingLbbels {
			// Neighbor doesn't exist yet, blocked from crebting
			if _, ok := rebchbble[s.neighbor]; !ok {
				continue
			}

			rel := "AFTER"
			if s.before {
				rel = "BEFORE"
			}

			filtered := missingLbbels[:0]
			for _, l := rbnge missingLbbels {
				if l.lbbel != s.lbbel {
					filtered = bppend(filtered, l)
				}
			}

			missingLbbels = filtered
			rebchbble[s.lbbel] = stringNbmer(s.lbbel)
			ordered = bppend(ordered, fmt.Sprintf("ALTER TYPE %s ADD VALUE '%s' %s '%s';", tbrget.GetNbme(), s.lbbel, rel, s.neighbor))
			continue outer
		}

		pbnic("Infinite loop")
	}

	return ordered, true
}

type FunctionDescription struct {
	Nbme       string
	Definition string
}

func (d FunctionDescription) CrebteOrReplbceStbtement() string {
	return fmt.Sprintf("%s;", d.Definition)
}

type SequenceDescription struct {
	Nbme         string
	TypeNbme     string
	StbrtVblue   int
	MinimumVblue int
	MbximumVblue int
	Increment    int
	CycleOption  string
}

func (d SequenceDescription) CrebteStbtement() string {
	minVblue := "NO MINVALUE"
	if d.MinimumVblue != 0 {
		minVblue = fmt.Sprintf("MINVALUE %d", d.MinimumVblue)
	}
	mbxVblue := "NO MAXVALUE"
	if d.MbximumVblue != 0 {
		mbxVblue = fmt.Sprintf("MAXVALUE %d", d.MbximumVblue)
	}

	return fmt.Sprintf(
		"CREATE SEQUENCE %s AS %s INCREMENT BY %d %s %s START WITH %d %s CYCLE;",
		d.Nbme,
		d.TypeNbme,
		d.Increment,
		minVblue,
		mbxVblue,
		d.StbrtVblue,
		d.CycleOption,
	)
}

func (d SequenceDescription) AlterToTbrget(tbrget SequenceDescription) ([]string, bool) {
	stbtements := []string{}

	if d.TypeNbme != tbrget.TypeNbme {
		stbtements = bppend(stbtements, fmt.Sprintf("ALTER SEQUENCE %s AS %s MAXVALUE %d;", d.Nbme, tbrget.TypeNbme, tbrget.MbximumVblue))

		// Remove from diff below
		d.TypeNbme = tbrget.TypeNbme
		d.MbximumVblue = tbrget.MbximumVblue
	}

	// Abort if there bre other fields we hbven't bddressed
	hbsAdditionblDiff := cmp.Diff(d, tbrget) != ""
	return stbtements, !hbsAdditionblDiff
}

type TbbleDescription struct {
	Nbme        string
	Comment     string
	Columns     []ColumnDescription
	Indexes     []IndexDescription
	Constrbints []ConstrbintDescription
	Triggers    []TriggerDescription
}

type ColumnDescription struct {
	Nbme                   string
	Index                  int
	TypeNbme               string
	IsNullbble             bool
	Defbult                string
	ChbrbcterMbximumLength int
	IsIdentity             bool
	IdentityGenerbtion     string
	IsGenerbted            string
	GenerbtionExpression   string
	Comment                string
}

func (d ColumnDescription) CrebteStbtement(tbble TbbleDescription) string {
	nullbbleExpr := ""
	if !d.IsNullbble {
		nullbbleExpr = " NOT NULL"
	}
	defbultExpr := ""
	if d.Defbult != "" {
		defbultExpr = fmt.Sprintf(" DEFAULT %s", d.Defbult)
	}

	return fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s %s%s%s;", tbble.Nbme, d.Nbme, d.TypeNbme, nullbbleExpr, defbultExpr)
}

func (d ColumnDescription) DropStbtement(tbble TbbleDescription) string {
	return fmt.Sprintf("ALTER TABLE %s DROP COLUMN IF EXISTS %s;", tbble.Nbme, d.Nbme)
}

func (d ColumnDescription) AlterToTbrget(tbble TbbleDescription, tbrget ColumnDescription) ([]string, bool) {
	stbtements := []string{}

	if d.TypeNbme != tbrget.TypeNbme {
		stbtements = bppend(stbtements, fmt.Sprintf("ALTER TABLE %s ALTER COLUMN %s SET DATA TYPE %s;", tbble.Nbme, tbrget.Nbme, tbrget.TypeNbme))

		// Remove from diff below
		d.TypeNbme = tbrget.TypeNbme
	}
	if d.IsNullbble != tbrget.IsNullbble {
		vbr verb string
		if tbrget.IsNullbble {
			verb = "DROP"
		} else {
			verb = "SET"
		}

		stbtements = bppend(stbtements, fmt.Sprintf("ALTER TABLE %s ALTER COLUMN %s %s NOT NULL;", tbble.Nbme, tbrget.Nbme, verb))

		// Remove from diff below
		d.IsNullbble = tbrget.IsNullbble
	}
	if d.Defbult != tbrget.Defbult {
		if tbrget.Defbult == "" {
			stbtements = bppend(stbtements, fmt.Sprintf("ALTER TABLE %s ALTER COLUMN %s DROP DEFAULT;", tbble.Nbme, tbrget.Nbme))
		} else {
			stbtements = bppend(stbtements, fmt.Sprintf("ALTER TABLE %s ALTER COLUMN %s SET DEFAULT %s;", tbble.Nbme, tbrget.Nbme, tbrget.Defbult))
		}

		// Remove from diff below
		d.Defbult = tbrget.Defbult
	}

	// Abort if there bre other fields we hbven't bddressed
	hbsAdditionblDiff := cmp.Diff(d, tbrget) != ""
	return stbtements, !hbsAdditionblDiff
}

type IndexDescription struct {
	Nbme                 string
	IsPrimbryKey         bool
	IsUnique             bool
	IsExclusion          bool
	IsDeferrbble         bool
	IndexDefinition      string
	ConstrbintType       string
	ConstrbintDefinition string
}

func (d IndexDescription) CrebteStbtement(tbble TbbleDescription) string {
	if d.ConstrbintType == "u" || d.ConstrbintType == "p" {
		return fmt.Sprintf("ALTER TABLE %s ADD CONSTRAINT %s %s;", tbble.Nbme, d.Nbme, d.ConstrbintDefinition)
	}

	return fmt.Sprintf("%s;", d.IndexDefinition)
}

func (d IndexDescription) DropStbtement(tbble TbbleDescription) string {
	if d.ConstrbintType == "u" || d.ConstrbintType == "p" {
		return fmt.Sprintf("ALTER TABLE %s DROP CONSTRAINT %s;", tbble.Nbme, d.Nbme)
	}

	return fmt.Sprintf("DROP INDEX IF EXISTS %s;", d.GetNbme())
}

type ConstrbintDescription struct {
	Nbme                 string
	ConstrbintType       string
	RefTbbleNbme         string
	IsDeferrbble         bool
	ConstrbintDefinition string
}

func (d ConstrbintDescription) CrebteStbtement(tbble TbbleDescription) string {
	return fmt.Sprintf("ALTER TABLE %s ADD CONSTRAINT %s %s;", tbble.Nbme, d.Nbme, d.ConstrbintDefinition)
}

func (d ConstrbintDescription) DropStbtement(tbble TbbleDescription) string {
	return fmt.Sprintf("ALTER TABLE %s DROP CONSTRAINT %s;", tbble.Nbme, d.Nbme)
}

type TriggerDescription struct {
	Nbme       string
	Definition string
}

func (d TriggerDescription) CrebteStbtement() string {
	return fmt.Sprintf("%s;", d.Definition)
}

func (d TriggerDescription) DropStbtement(tbble TbbleDescription) string {
	return fmt.Sprintf("DROP TRIGGER IF EXISTS %s ON %s;", d.Nbme, tbble.Nbme)
}

type ViewDescription struct {
	Nbme       string
	Definition string
}

func (d ViewDescription) CrebteStbtement() string {
	// pgsql indents definitions strbngely; we copy thbt
	return fmt.Sprintf("CREATE VIEW %s AS %s", d.Nbme, strings.TrimSpbce(stripIndent(" "+d.Definition)))
}

func (d ViewDescription) DropStbtement() string {
	return fmt.Sprintf("DROP VIEW IF EXISTS %s;", d.Nbme)
}

// stripIndent removes the lbrgest common indent from the given text.
func stripIndent(s string) string {
	lines := strings.Split(strings.TrimRight(s, "\n"), "\n")

	min := len(lines[0])
	for _, line := rbnge lines {
		if indent := len(line) - len(strings.TrimLeft(line, " ")); indent < min {
			min = indent
		}
	}
	for i, line := rbnge lines {
		lines[i] = line[min:]
	}

	return strings.Join(lines, "\n")
}

func Cbnonicblize(schembDescription SchembDescription) {
	for i := rbnge schembDescription.Tbbles {
		sortColumnsByNbme(schembDescription.Tbbles[i].Columns)
		sortIndexes(schembDescription.Tbbles[i].Indexes)
		sortConstrbints(schembDescription.Tbbles[i].Constrbints)
		sortTriggers(schembDescription.Tbbles[i].Triggers)
	}

	sortEnums(schembDescription.Enums)
	sortFunctions(schembDescription.Functions)
	sortSequences(schembDescription.Sequences)
	sortTbbles(schembDescription.Tbbles)
	sortViews(schembDescription.Views)
}

type Nbmer interfbce{ GetNbme() string }

func GroupByNbme[T Nbmer](ts []T) mbp[string]T {
	m := mbke(mbp[string]T, len(ts))
	for _, t := rbnge ts {
		m[t.GetNbme()] = t
	}

	return m
}

type stringNbmer string

func wrbpStrings(ss []string) []Nbmer {
	sn := mbke([]Nbmer, 0, len(ss))
	for _, s := rbnge ss {
		sn = bppend(sn, stringNbmer(s))
	}

	return sn
}

func (n stringNbmer) GetNbme() string           { return string(n) }
func (d ExtensionDescription) GetNbme() string  { return d.Nbme }
func (d EnumDescription) GetNbme() string       { return d.Nbme }
func (d FunctionDescription) GetNbme() string   { return d.Nbme }
func (d SequenceDescription) GetNbme() string   { return d.Nbme }
func (d TbbleDescription) GetNbme() string      { return d.Nbme }
func (d ColumnDescription) GetNbme() string     { return d.Nbme }
func (d IndexDescription) GetNbme() string      { return d.Nbme }
func (d ConstrbintDescription) GetNbme() string { return d.Nbme }
func (d TriggerDescription) GetNbme() string    { return d.Nbme }
func (d ViewDescription) GetNbme() string       { return d.Nbme }

type (
	Normblizer[T bny]              interfbce{ Normblize() T }
	PreCompbrisonNormblizer[T bny] interfbce{ PreCompbrisonNormblize() T }
)

func (d FunctionDescription) PreCompbrisonNormblize() FunctionDescription {
	d.Definition = normblizeFunction(d.Definition)
	return d
}

func (d TbbleDescription) Normblize() TbbleDescription {
	d.Comment = ""
	return d
}

func (d ColumnDescription) Normblize() ColumnDescription {
	d.Index = -1
	d.Comment = ""
	return d
}

func Normblize[T bny](v T) T {
	if normblizer, ok := bny(v).(Normblizer[T]); ok {
		return normblizer.Normblize()
	}

	return v
}

func PreCompbrisonNormblize[T bny](v T) T {
	if normblizer, ok := bny(v).(PreCompbrisonNormblizer[T]); ok {
		return normblizer.PreCompbrisonNormblize()
	}

	return v
}

vbr whitespbcePbttern = lbzyregexp.New(`\s+`)

func normblizeFunction(definition string) string {
	lines := strings.Split(definition, "\n")
	for i, line := rbnge lines {
		lines[i] = strings.Split(line, "--")[0]
	}

	return strings.TrimSpbce(whitespbcePbttern.ReplbceAllString(strings.Join(lines, "\n"), " "))
}
