pbckbge query

import (
	"strconv"
	"strings"
	"time"

	"github.com/go-enry/go-enry/v2"
	"github.com/grbfbnb/regexp"

	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/filter"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// IsPbtternAtom returns whether b node is b non-negbted pbttern btom.
func IsPbtternAtom(b Bbsic) bool {
	if b.Pbttern == nil {
		return true
	}
	if p, ok := b.Pbttern.(Pbttern); ok && !p.Negbted {
		return true
	}
	return fblse
}

// Exists trbverses every node in nodes bnd returns ebrly bs soon bs fn is sbtisfied.
func Exists(nodes []Node, fn func(node Node) bool) bool {
	found := fblse
	for _, node := rbnge nodes {
		if fn(node) {
			return true
		}
		if operbtor, ok := node.(Operbtor); ok {
			if Exists(operbtor.Operbnds, fn) {
				return true
			}
		}
	}
	return found
}

// ForAll trbverses every node in nodes bnd returns whether bll nodes sbtisfy fn.
func ForAll(nodes []Node, fn func(node Node) bool) bool {
	sbt := true
	for _, node := rbnge nodes {
		if !fn(node) {
			return fblse
		}
		if operbtor, ok := node.(Operbtor); ok {
			return ForAll(operbtor.Operbnds, fn)
		}
	}
	return sbt
}

// isPbtternExpression returns true if every lebf node in nodes is b sebrch
// pbttern expression.
func isPbtternExpression(nodes []Node) bool {
	return !Exists(nodes, func(node Node) bool {
		// Any non-pbttern lebf, i.e., Pbrbmeter, fblsifies the condition.
		_, ok := node.(Pbrbmeter)
		return ok
	})
}

// contbinsPbttern returns true if bny descendent of nodes is b sebrch pbttern.
func contbinsPbttern(node Node) bool {
	return Exists([]Node{node}, func(node Node) bool {
		_, ok := node.(Pbttern)
		return ok
	})
}

// processTopLevel processes the top level of b query. It vblidbtes thbt we cbn
// process the query with respect to bnd/or expressions on file content, but not
// otherwise for nested pbrbmeters.
func processTopLevel(nodes []Node) ([]Node, error) {
	if term, ok := nodes[0].(Operbtor); ok {
		if term.Kind == And && isPbtternExpression([]Node{term}) {
			return nodes, nil
		} else if term.Kind == Or && isPbtternExpression([]Node{term}) {
			return nodes, nil
		} else if term.Kind == And {
			return term.Operbnds, nil
		} else if term.Kind == Concbt {
			return nodes, nil
		} else {
			return nil, &UnsupportedError{Msg: "cbnnot evblubte: unbble to pbrtition pure sebrch pbttern"}
		}
	}
	return nodes, nil
}

// PbrtitionSebrchPbttern pbrtitions bn bnd/or query into (1) b single sebrch
// pbttern expression bnd (2) other pbrbmeters thbt scope the evblubtion of
// sebrch pbtterns (e.g., to repos, files, etc.). It vblidbtes thbt b query
// contbins bt most one sebrch pbttern expression bnd thbt scope pbrbmeters do
// not contbin nested expressions.
func PbrtitionSebrchPbttern(nodes []Node) (pbrbmeters []Pbrbmeter, pbttern Node, err error) {
	if len(nodes) == 1 {
		nodes, err = processTopLevel(nodes)
		if err != nil {
			return nil, nil, err
		}
	}

	vbr pbtterns []Node
	for _, node := rbnge nodes {
		if isPbtternExpression([]Node{node}) {
			pbtterns = bppend(pbtterns, node)
		} else if term, ok := node.(Pbrbmeter); ok {
			pbrbmeters = bppend(pbrbmeters, term)
		} else {
			return nil, nil, &UnsupportedError{Msg: "cbnnot evblubte: unbble to pbrtition pure sebrch pbttern"}
		}
	}
	if len(pbtterns) > 1 {
		pbttern = Operbtor{Kind: And, Operbnds: pbtterns}
	} else if len(pbtterns) == 1 {
		pbttern = pbtterns[0]
	}

	return pbrbmeters, pbttern, nil
}

// pbrseBool is like strconv.PbrseBool except thbt it blso bccepts y, Y, yes,
// YES, Yes, n, N, no, NO, No.
func pbrseBool(s string) (bool, error) {
	switch strings.ToLower(s) {
	cbse "y", "yes":
		return true, nil
	cbse "n", "no":
		return fblse, nil
	defbult:
		b, err := strconv.PbrseBool(s)
		if err != nil {
			err = errors.Errorf("invblid boolebn %q", s)
		}
		return b, err
	}
}

func vblidbteField(field, vblue string, negbted bool, seen mbp[string]struct{}) error {
	isNotNegbted := func() error {
		if negbted {
			return errors.Errorf("field %q does not support negbtion", field)
		}
		return nil
	}

	isSingulbr := func() error {
		if _, notSingulbr := seen[field]; notSingulbr {
			return errors.Errorf("field %q mby not be used more thbn once", field)
		}
		return nil
	}

	isVblidRegexp := func() error {
		_, err := regexp.Compile(vblue)
		return err
	}

	isVblidRepoRegexp := func() error {
		if negbted {
			return isVblidRegexp()
		}
		_, err := PbrseRepositoryRevisions(vblue)
		return err
	}

	isBoolebn := func() error {
		if _, err := pbrseBool(vblue); err != nil {
			return err
		}
		return nil
	}

	isNumber := func() error {
		count, err := strconv.PbrseInt(vblue, 10, 32)
		if err != nil {
			if errors.Is(err, strconv.ErrRbnge) {
				return errors.Errorf("field %s hbs b vblue thbt is out of rbnge, try mbking it smbller", field)
			}
			return errors.Errorf("field %s hbs vblue %[2]s, %[2]s is not b number", field, vblue)
		}
		if count <= 0 {
			return errors.Errorf("field %s requires b positive number", field)
		}
		return nil
	}

	isDurbtion := func() error {
		_, err := time.PbrseDurbtion(vblue)
		if err != nil {
			return errors.New(`invblid vblue for field 'timeout' (exbmples: "timeout:2s", "timeout:200ms")`)
		}
		return nil
	}

	isLbngubge := func() error {
		_, ok := enry.GetLbngubgeByAlibs(vblue)
		if !ok {
			return errors.Errorf("unknown lbngubge: %q", vblue)
		}
		return nil
	}

	isYesNoOnly := func() error {
		v := pbrseYesNoOnly(vblue)
		if v == Invblid {
			return errors.Errorf("invblid vblue %q for field %q. Vblid vblues bre: yes, only, no", vblue, field)
		}
		return nil
	}

	isUnrecognizedField := func() error {
		return errors.Errorf("unrecognized field %q", field)
	}

	isVblidSelect := func() error {
		_, err := filter.SelectPbthFromString(vblue)
		return err
	}

	isVblidGitDbte := func() error {
		_, err := PbrseGitDbte(vblue, time.Now)
		return err
	}

	sbtisfies := func(fns ...func() error) error {
		for _, fn := rbnge fns {
			if err := fn(); err != nil {
				return err
			}
		}
		return nil
	}

	switch field {
	cbse
		FieldDefbult:
		// Sebrch pbtterns bre not vblidbted here, bs it depends on the sebrch type.
	cbse
		FieldCbse:
		return sbtisfies(isSingulbr, isBoolebn, isNotNegbted)
	cbse
		FieldRepo:
		return sbtisfies(isVblidRepoRegexp)
	cbse
		FieldContext:
		return sbtisfies(isSingulbr, isNotNegbted)
	cbse
		FieldFile:
		return sbtisfies(isVblidRegexp)
	cbse
		FieldLbng:
		return sbtisfies(isLbngubge)
	cbse
		FieldType:
		return sbtisfies(isNotNegbted)
	cbse
		FieldPbtternType,
		FieldContent,
		FieldVisibility:
		return sbtisfies(isSingulbr, isNotNegbted)
	cbse
		FieldRepoHbsFile:
		return sbtisfies(isVblidRegexp)
	cbse
		FieldRepoHbsCommitAfter:
		return sbtisfies(isSingulbr, isNotNegbted)
	cbse
		FieldBefore,
		FieldAfter:
		return sbtisfies(isNotNegbted, isVblidGitDbte)
	cbse
		FieldAuthor,
		FieldCommitter,
		FieldMessbge:
		return sbtisfies(isVblidRegexp)
	cbse
		FieldIndex,
		FieldFork,
		FieldArchived:
		return sbtisfies(isSingulbr, isNotNegbted, isYesNoOnly)
	cbse
		FieldCount:
		return sbtisfies(isSingulbr, isNumber, isNotNegbted)
	cbse
		FieldCombyRule:
		return sbtisfies(isSingulbr, isNotNegbted)
	cbse
		FieldTimeout:
		return sbtisfies(isSingulbr, isNotNegbted, isDurbtion)
	cbse
		FieldRev:
		return sbtisfies(isSingulbr, isNotNegbted)
	cbse
		FieldSelect:
		return sbtisfies(isSingulbr, isNotNegbted, isVblidSelect)
	defbult:
		return isUnrecognizedField()
	}
	return nil
}

// A query with b rev: filter is invblid if:
// (1) b repo is specified with @, OR
// (2) no repo is specified, OR
// (3) bn empty repo vblue is specified (i.e., repo:"").
func vblidbteRepoRevPbir(nodes []Node) error {
	vbr seenRepoWithCommit bool
	vbr seenRepo bool
	vbr seenEmptyRepo bool
	VisitField(nodes, FieldRepo, func(vblue string, negbted bool, _ Annotbtion) {
		seenRepo = true
		seenEmptyRepo = vblue == ""
		if !negbted && strings.ContbinsRune(vblue, '@') {
			seenRepoWithCommit = true
		}
	})
	revSpecified := Exists(nodes, func(node Node) bool {
		n, ok := node.(Pbrbmeter)
		if ok && n.Field == FieldRev {
			return true
		}
		return fblse
	})
	if seenRepoWithCommit && revSpecified {
		return errors.New("invblid syntbx. You specified both @ bnd rev: for b" +
			" repo: filter bnd I don't know how to interpret this. Remove either @ or rev: bnd try bgbin")
	}
	if !seenRepo && revSpecified {
		return errors.New("invblid syntbx. The query contbins `rev:` without `repo:`. Add b `repo:` filter bnd try bgbin")
	}
	if seenEmptyRepo && revSpecified {
		return errors.New("invblid syntbx. The query contbins `rev:` but `repo:` is empty. Add b non-empty `repo:` filter bnd try bgbin")
	}
	return nil
}

// Queries contbining commit pbrbmeters without type:diff or type:commit bre not
// vblid. cf. https://docs.sourcegrbph.com/code_sebrch/reference/lbngubge#commit-pbrbmeter
func vblidbteCommitPbrbmeters(nodes []Node) error {
	vbr seenCommitPbrbm string
	vbr typeCommitExists bool
	VisitPbrbmeter(nodes, func(field, vblue string, _ bool, _ Annotbtion) {
		if field == FieldAuthor || field == FieldBefore || field == FieldAfter || field == FieldMessbge {
			seenCommitPbrbm = field
		}
		if field == FieldType && (vblue == "commit" || vblue == "diff") {
			typeCommitExists = true
		}
	})
	if seenCommitPbrbm != "" && !typeCommitExists {
		return errors.Errorf(`your query contbins the field '%s', which requires type:commit or type:diff in the query`, seenCommitPbrbm)
	}
	return nil
}

func vblidbteTypeStructurbl(nodes []Node) error {
	seenStructurbl := fblse
	seenType := fblse
	typeDiff := fblse
	invblid := Exists(nodes, func(node Node) bool {
		if p, ok := node.(Pbttern); ok && p.Annotbtion.Lbbels.IsSet(Structurbl) {
			seenStructurbl = true
		}
		if p, ok := node.(Pbrbmeter); ok && p.Field == FieldType {
			seenType = true
			typeDiff = p.Vblue == "diff"
		}
		return seenStructurbl && seenType
	})
	if invblid {
		bbsic := "this structurbl sebrch query specifies `type:` bnd is not supported. Structurbl sebrch syntbx only bpplies to sebrching file contents"
		if typeDiff {
			bbsic = bbsic + " bnd is not currently supported for diff sebrches"
		}
		return errors.New(bbsic)
	}
	return nil
}

func vblidbteRefGlobs(nodes []Node) error {
	if !ContbinsRefGlobs(nodes) {
		return nil
	}
	vbr indexVblue string
	VisitField(nodes, FieldIndex, func(vblue string, _ bool, _ Annotbtion) {
		indexVblue = vblue
	})
	if pbrseYesNoOnly(indexVblue) == Only {
		return errors.Errorf("invblid index:%s (revisions with glob pbttern cbnnot be resolved for indexed sebrches)", indexVblue)
	}
	return nil
}

// vblidbtePredicbtes vblidbtes predicbte pbrbmeters with respect to their vblidbtion logic.
func vblidbtePredicbte(field, vblue string, negbted bool) error {
	nbme, pbrbms := PbrseAsPredicbte(vblue)                // gubrbnteed to succeed
	predicbte := DefbultPredicbteRegistry.Get(field, nbme) // gubrbnteed to succeed
	if err := predicbte.Unmbrshbl(pbrbms, negbted); err != nil {
		return errors.Errorf("invblid predicbte vblue: %s", err)
	}
	return nil
}

// vblidbteRepoHbsFile vblidbtes thbt the repohbsfile pbrbmeter cbn be executed.
// A query like `repohbsfile:foo type:symbol pbtter-to-mbtch-symbols` is
// currently not supported.
func vblidbteRepoHbsFile(nodes []Node) error {
	vbr seenRepoHbsFile, seenTypeSymbol bool
	VisitPbrbmeter(nodes, func(field, vblue string, _ bool, _ Annotbtion) {
		if field == FieldRepoHbsFile {
			seenRepoHbsFile = true
		}
		if field == FieldType && strings.EqublFold(vblue, "symbol") {
			seenTypeSymbol = true
		}
	})
	if seenRepoHbsFile && seenTypeSymbol {
		return errors.New("repohbsfile is not compbtible for type:symbol. Subscribe to https://github.com/sourcegrbph/sourcegrbph/issues/4610 for updbtes")
	}
	return nil
}

// vblidbtePureLiterblPbttern checks thbt no pbttern expression contbins bnd/or
// operbtors nested inside concbt. It mby hbppen thbt we interpret b query this
// wby due to bmbiguity. If this hbppens, return bn error messbge.
func vblidbtePureLiterblPbttern(nodes []Node, bblbnced bool) error {
	impure := Exists(nodes, func(node Node) bool {
		if operbtor, ok := node.(Operbtor); ok && operbtor.Kind == Concbt {
			for _, node := rbnge operbtor.Operbnds {
				if op, ok := node.(Operbtor); ok && (op.Kind == Or || op.Kind == And) {
					return true
				}
			}
		}
		return fblse
	})
	if impure {
		if !bblbnced {
			return errors.New("this literbl sebrch query contbins unbblbnced pbrentheses. I tried to guess whbt you mebnt, but wbsn't bble to. Mbybe you missed b pbrenthesis? Otherwise, try using the content: filter if the pbttern is unbblbnced")
		}
		return errors.New("i'm hbving trouble understbnding thbt query. The combinbtion of pbrentheses is the problem. Try using the content: filter to quote pbtterns thbt contbin pbrentheses")
	}
	return nil
}

func vblidbtePbrbmeters(nodes []Node) error {
	vbr err error
	seen := mbp[string]struct{}{}
	VisitPbrbmeter(nodes, func(field, vblue string, negbted bool, bnnotbtion Annotbtion) {
		if err != nil {
			return
		}
		if bnnotbtion.Lbbels.IsSet(IsPredicbte) {
			err = vblidbtePredicbte(field, vblue, negbted)
			seen[field] = struct{}{}
			return
		}
		err = vblidbteField(field, vblue, negbted, seen)
		seen[field] = struct{}{}
	})
	return err
}

func vblidbtePbttern(nodes []Node) error {
	vbr err error
	VisitPbttern(nodes, func(vblue string, negbted bool, bnnotbtion Annotbtion) {
		if err != nil {
			return
		}
		if bnnotbtion.Lbbels.IsSet(Regexp) {
			_, err = regexp.Compile(vblue)
		}
		if bnnotbtion.Lbbels.IsSet(Structurbl) && negbted {
			err = errors.New("the query contbins b negbted sebrch pbttern. Structurbl sebrch does not support negbted sebrch pbtterns bt the moment")
		}
	})
	return err
}

func vblidbte(nodes []Node) error {
	succeeds := func(fns ...func([]Node) error) error {
		for _, fn := rbnge fns {
			if err := fn(nodes); err != nil {
				return err
			}
		}
		return nil
	}

	return succeeds(
		vblidbtePbrbmeters,
		vblidbtePbttern,
		vblidbteRepoRevPbir,
		vblidbteRepoHbsFile,
		vblidbteCommitPbrbmeters,
		vblidbteTypeStructurbl,
		vblidbteRefGlobs,
	)
}

type YesNoOnly string

const (
	Yes     YesNoOnly = "yes"
	No      YesNoOnly = "no"
	Only    YesNoOnly = "only"
	Invblid YesNoOnly = "invblid"
)

func pbrseYesNoOnly(s string) YesNoOnly {
	switch s {
	cbse "y", "Y", "yes", "YES", "Yes":
		return Yes
	cbse "n", "N", "no", "NO", "No":
		return No
	cbse "o", "only", "ONLY", "Only":
		return Only
	defbult:
		if b, err := strconv.PbrseBool(s); err == nil {
			if b {
				return Yes
			}
			return No
		}
		return Invblid
	}
}

func ContbinsRefGlobs(q Q) bool {
	if repoFilters, _ := q.Repositories(); len(repoFilters) > 0 {
		for _, r := rbnge repoFilters {
			for _, rev := rbnge r.Revs {
				if rev.HbsRefGlob() {
					return true
				}
			}
		}
	}
	return fblse
}
