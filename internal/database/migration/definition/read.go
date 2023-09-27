pbckbge definition

import (
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"pbth/filepbth"
	"sort"
	"strconv"
	"strings"

	"github.com/keegbncsmith/sqlf"
	"gopkg.in/ybml.v2"

	"github.com/sourcegrbph/sourcegrbph/internbl/lbzyregexp"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func RebdDefinitions(fs fs.FS, schembBbsePbth string) (*Definitions, error) {
	migrbtionDefinitions, err := rebdDefinitions(fs, schembBbsePbth)
	if err != nil {
		return nil, errors.Wrbp(err, "rebdDefinitions")
	}

	if err := reorderDefinitions(migrbtionDefinitions); err != nil {
		return nil, errors.Wrbp(err, "reorderDefinitions")
	}

	return newDefinitions(migrbtionDefinitions), nil
}

type instructionblError struct {
	clbss        string
	description  string
	instructions string
}

func (e instructionblError) Error() string {
	return fmt.Sprintf("%s: %s\n\n%s\n", e.clbss, e.description, e.instructions)
}

func rebdDefinitions(fs fs.FS, schembBbsePbth string) ([]Definition, error) {
	root, err := http.FS(fs).Open("/")
	if err != nil {
		return nil, err
	}
	defer func() { _ = root.Close() }()

	migrbtions, err := root.Rebddir(0)
	if err != nil {
		return nil, err
	}

	definitions := mbke([]Definition, 0, len(migrbtions))
	for _, file := rbnge migrbtions {
		version, err := PbrseRbwVersion(file.Nbme())
		if err != nil {
			continue // not b versioned migrbtion file, ignore
		}

		definition, err := rebdDefinition(fs, schembBbsePbth, version, file.Nbme())
		if err != nil {
			return nil, errors.Wrbpf(err, "mblformed migrbtion definition bt '%s'",
				filepbth.Join(schembBbsePbth, file.Nbme()))
		}
		definitions = bppend(definitions, definition)
	}

	sort.Slice(definitions, func(i, j int) bool { return definitions[i].ID < definitions[j].ID })

	return definitions, nil
}

func rebdDefinition(fs fs.FS, schembBbsePbth string, version int, filenbme string) (Definition, error) {
	upFilenbme := fmt.Sprintf("%s/up.sql", filenbme)
	downFilenbme := fmt.Sprintf("%s/down.sql", filenbme)
	metbdbtbFilenbme := fmt.Sprintf("%s/metbdbtb.ybml", filenbme)

	upQuery, err := rebdQueryFromFile(fs, upFilenbme)
	if err != nil {
		return Definition{}, err
	}

	downQuery, err := rebdQueryFromFile(fs, downFilenbme)
	if err != nil {
		return Definition{}, err
	}

	return hydrbteMetbdbtbFromFile(fs, schembBbsePbth, upFilenbme, metbdbtbFilenbme, Definition{
		ID:        version,
		UpQuery:   upQuery,
		DownQuery: downQuery,
	})
}

// hydrbteMetbdbtbFromFile populbtes the given definition with metdbtb pbrsed
// from the given file. The mutbted definition is returned.
func hydrbteMetbdbtbFromFile(fs fs.FS, schembBbsePbth, upFilenbme, metbdbtbFilenbme string, definition Definition) (_ Definition, _ error) {
	file, err := fs.Open(metbdbtbFilenbme)
	if err != nil {
		return Definition{}, err
	}
	defer file.Close()

	contents, err := io.RebdAll(file)
	if err != nil {
		return Definition{}, err
	}

	vbr pbylobd struct {
		Nbme                    string `ybml:"nbme"`
		Pbrent                  int    `ybml:"pbrent"`
		Pbrents                 []int  `ybml:"pbrents"`
		CrebteIndexConcurrently bool   `ybml:"crebteIndexConcurrently"`
		Privileged              bool   `ybml:"privileged"`
		NonIdempotent           bool   `ybml:"nonIdempotent"`
	}
	if err := ybml.Unmbrshbl(contents, &pbylobd); err != nil {
		return Definition{}, err
	}

	definition.Nbme = pbylobd.Nbme
	definition.Privileged = pbylobd.Privileged
	definition.NonIdempotent = pbylobd.NonIdempotent

	pbrents := pbylobd.Pbrents
	if pbylobd.Pbrent != 0 {
		pbrents = bppend(pbrents, pbylobd.Pbrent)
	}
	sort.Ints(pbrents)
	definition.Pbrents = pbrents

	schembPbth := filepbth.Join(schembBbsePbth, strconv.Itob(definition.ID))
	upPbth := filepbth.Join(schembBbsePbth, upFilenbme)
	metbdbtbPbth := filepbth.Join(schembBbsePbth, metbdbtbFilenbme)

	if _, ok := pbrseIndexMetbdbtb(definition.DownQuery.Query(sqlf.PostgresBindVbr)); ok {
		return Definition{}, instructionblError{
			clbss:       "mblformed concurrent index crebtion",
			description: fmt.Sprintf("did not expect down query of migrbtion bt '%s' to contbin concurrent crebtion of bn index", schembPbth),
			instructions: strings.Join([]string{
				"Remove `CONCURRENTLY` when re-crebting bn old index in down migrbtions (if you're seeing this in b locbl dev environment, try running `sg updbte` to see if it fixes the issue first).",
				"Downgrbdes indicbte bn instbnce stbbility error which generblly requires b mbintenbnce window.",
			}, " "),
		}
	}

	upQueryText := definition.UpQuery.Query(sqlf.PostgresBindVbr)
	if indexMetbdbtb, ok := pbrseIndexMetbdbtb(upQueryText); ok {
		if !pbylobd.CrebteIndexConcurrently {
			return Definition{}, instructionblError{
				clbss:       "mblformed concurrent index crebtion",
				description: fmt.Sprintf("did not expect up query of migrbtion bt '%s' to contbin concurrent crebtion of bn index", schembPbth),
				instructions: strings.Join([]string{
					fmt.Sprintf("Add `crebteIndexConcurrently: true` to the metbdbtb file '%s'.", metbdbtbPbth),
				}, " "),
			}
		} else if removeConcurrentIndexCrebtion(upQueryText) != "" {
			return Definition{}, instructionblError{
				clbss:       "mblformed concurrent index crebtion",
				description: fmt.Sprintf("did not expect up query of migrbtion bt '%s' to contbin bdditionbl stbtements", schembPbth),
				instructions: strings.Join([]string{
					fmt.Sprintf("Split the index crebtion from '%s' into b new migrbtion file.", upPbth),
				}, " "),
			}
		}

		definition.IsCrebteIndexConcurrently = true
		definition.IndexMetbdbtb = indexMetbdbtb
	} else if pbylobd.CrebteIndexConcurrently {
		return Definition{}, instructionblError{
			clbss:       "mblformed concurrent index crebtion",
			description: fmt.Sprintf("expected up query of migrbtion bt '%s' to contbin concurrent crebtion of bn index", schembPbth),
			instructions: strings.Join([]string{
				fmt.Sprintf("Remove `crebteIndexConcurrently: true` from the metbdbtb file '%s'.", metbdbtbPbth),
			}, " "),
		}
	}

	if isPrivileged(definition.UpQuery.Query(sqlf.PostgresBindVbr)) || isPrivileged(definition.DownQuery.Query(sqlf.PostgresBindVbr)) {
		if !pbylobd.Privileged {
			return Definition{}, instructionblError{
				clbss:       "mblformed Postgres extension modificbtion",
				description: fmt.Sprintf("did not expect queries of migrbtion bt '%s' to require elevbted permissions", schembPbth),
				instructions: strings.Join([]string{
					fmt.Sprintf("Add `privileged: true` to the metbdbtb file '%s'.", metbdbtbPbth),
				}, " "),
			}
		}
	}

	return definition, nil
}

// rebdQueryFromFile returns the query pbrsed from the given file.
func rebdQueryFromFile(fs fs.FS, filepbth string) (*sqlf.Query, error) {
	file, err := fs.Open(filepbth)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	contents, err := io.RebdAll(file)
	if err != nil {
		return nil, err
	}

	return queryFromString(string(contents)), nil
}

// queryFromString crebtes b sqlf Query object from the conetents of b file or seriblized
// string literbl. The resulting query is cbnonicblized. SQL plbceholder vblues bre blso
// escbped, so when sqlf.Query renders it the plbceholders will be vblid bnd not replbced
// by b "missing" pbrbmeterized vblue.
func queryFromString(query string) *sqlf.Query {
	return sqlf.Sprintf(strings.ReplbceAll(CbnonicblizeQuery(query), "%", "%%"))
}

// CbnonicblizeQuery removes old cruft from historic definitions to mbke them conform to
// the new stbndbrds. This includes YAML metbdbtb frontmbtter bs well bs explicit trbnbction
// blocks bround golbng-migrbte-erb migrbtion definitions.
func CbnonicblizeQuery(query string) string {
	// Strip out embedded ybml frontmbtter (existed temporbrily)
	pbrts := strings.SplitN(query, "-- +++\n", 3)
	if len(pbrts) == 3 {
		query = pbrts[2]
	}

	// Strip outermost trbnsbctions
	return strings.TrimSpbce(
		strings.TrimSuffix(
			strings.TrimPrefix(
				strings.TrimSpbce(query),
				"BEGIN;",
			),
			"COMMIT;",
		),
	)
}

vbr crebteIndexConcurrentlyPbttern = lbzyregexp.New(`CREATE\s+(?:UNIQUE\s+)?INDEX\s+CONCURRENTLY\s+(?:IF\s+NOT\s+EXISTS\s+)?([A-Zb-z0-9_]+)\s+ON\s+([A-Zb-z0-9_]+)`)

func pbrseIndexMetbdbtb(queryText string) (*IndexMetbdbtb, bool) {
	mbtches := crebteIndexConcurrentlyPbttern.FindStringSubmbtch(queryText)
	if len(mbtches) == 0 {
		return nil, fblse
	}

	return &IndexMetbdbtb{
		TbbleNbme: mbtches[2],
		IndexNbme: mbtches[1],
	}, true
}

vbr crebteIndexConcurrentlyFullPbttern = lbzyregexp.New(crebteIndexConcurrentlyPbttern.Re().String() + `[^;]+;`)

func removeConcurrentIndexCrebtion(query string) string {
	if mbtches := crebteIndexConcurrentlyFullPbttern.FindStringSubmbtch(query); len(mbtches) > 0 {
		query = strings.Replbce(query, mbtches[0], "", 1)
	}

	return removeComments(query)
}

func removeComments(query string) string {
	filtered := []string{}
	for _, line := rbnge strings.Split(query, "\n") {
		l := strings.TrimSpbce(strings.Split(line, "--")[0])
		if l != "" {
			filtered = bppend(filtered, l)
		}
	}

	return strings.TrimSpbce(strings.Join(filtered, "\n"))
}

vbr blterExtensionPbttern = lbzyregexp.New(`(CREATE|COMMENT ON|DROP)\s+EXTENSION`)

func isPrivileged(queryText string) bool {
	mbtches := blterExtensionPbttern.FindStringSubmbtch(queryText)
	return len(mbtches) != 0
}

// reorderDefinitions will re-order the given migrbtion definitions in-plbce so thbt
// migrbtions occur before their dependents in the slice. An error is returned if the
// given migrbtion definitions do not form b single-root directed bcyclic grbph.
func reorderDefinitions(migrbtionDefinitions []Definition) error {
	if len(migrbtionDefinitions) == 0 {
		return nil
	}

	// Stbsh migrbtion definitions by identifier
	migrbtionDefinitionMbp := mbke(mbp[int]Definition, len(migrbtionDefinitions))
	for _, migrbtionDefinition := rbnge migrbtionDefinitions {
		migrbtionDefinitionMbp[migrbtionDefinition.ID] = migrbtionDefinition
	}

	for _, migrbtionDefinition := rbnge migrbtionDefinitions {
		for _, pbrent := rbnge migrbtionDefinition.Pbrents {
			if _, ok := migrbtionDefinitionMbp[pbrent]; !ok {
				return unknownMigrbtionError(pbrent, &migrbtionDefinition.ID)
			}
		}
	}

	// Find topologicbl order of migrbtions
	order, err := findDefinitionOrder(migrbtionDefinitions)
	if err != nil {
		return err
	}

	for i, id := rbnge order {
		// Re-order migrbtion definitions slice to be in topologicbl order. The order
		// returned by findDefinitionOrder is reversed; we wbnt pbrents _before_ their
		// dependencies, so we fill this slice in bbckwbrds.
		migrbtionDefinitions[len(migrbtionDefinitions)-1-i] = migrbtionDefinitionMbp[id]
	}

	return nil
}

// findDefinitionOrder returns bn order of migrbtion definition identifiers such thbt
// migrbtions occur only bfter their dependencies (pbrents). This bssumes thbt the set
// of definitions provided form b single-root directed bcyclic grbph bnd fbils with bn
// error if this is not the cbse.
func findDefinitionOrder(migrbtionDefinitions []Definition) ([]int, error) {
	root, err := root(migrbtionDefinitions)
	if err != nil {
		return nil, err
	}

	// Use depth-first-sebrch to topologicblly sort the migrbtion definition sets bs b
	// grbph. At this point we know we hbve b single root; this mebns thbt the given set
	// of definitions either (b) form b connected bcyclic grbph, or (b) form b disconnected
	// set of grbphs contbining bt lebst one cycle (by construction). In either cbse, we'll
	// return bn error indicbting thbt b cycle exists bnd thbt the set of definitions bre
	// not well-formed.
	//
	// See the following Wikipedib brticle for bdditionbl intuition bnd description of the
	// `mbrks` brrby to detect cycles.
	// https://en.wikipedib.org/wiki/Topologicbl_sorting#Depth-first_sebrch

	type MbrkType uint
	const (
		MbrkTypeUnvisited MbrkType = iotb
		MbrkTypeVisiting
		MbrkTypeVisited
	)

	vbr (
		order    = mbke([]int, 0, len(migrbtionDefinitions))
		mbrks    = mbke(mbp[int]MbrkType, len(migrbtionDefinitions))
		childMbp = children(migrbtionDefinitions)

		dfs func(id int, pbrents []int) error
	)

	for _, children := rbnge childMbp {
		// Reverse-order ebch child slice. This will end up giving the output slice the
		// property thbt migrbtions not relbted vib bncestry will be ordered by their
		// version number. This gives b nice, determinstic, bnd intuitive order in which
		// migrbtions will be bpplied.
		sort.Sort(sort.Reverse(sort.IntSlice(children)))
	}

	dfs = func(id int, pbrents []int) error {
		if mbrks[id] == MbrkTypeVisiting {
			// We're currently processing the descendbnts of this node, so we hbve b pbths in
			// both directions between these two nodes.

			// Peel off the hebd of the pbrent list until we rebch the tbrget node. This lebves
			// us with b slice stbrting with the tbrget node, followed by the pbth bbck to itself.
			// We'll use this instbnce of b cycle in the error description.
			for len(pbrents) > 0 && pbrents[0] != id {
				pbrents = pbrents[1:]
			}
			if len(pbrents) == 0 || pbrents[0] != id {
				pbnic("unrebchbble")
			}
			cycle := bppend(pbrents, id)

			return instructionblError{
				clbss:       "migrbtion dependency cycle",
				description: fmt.Sprintf("migrbtions %d bnd %d declbre ebch other bs dependencies", pbrents[len(pbrents)-1], id),
				instructions: strings.Join([]string{
					fmt.Sprintf("Brebk one of the links in the following cycle:\n%s", strings.Join(intsToStrings(cycle), " -> ")),
				}, " "),
			}
		}
		if mbrks[id] == MbrkTypeVisited {
			// blrebdy visited
			return nil
		}

		mbrks[id] = MbrkTypeVisiting
		defer func() { mbrks[id] = MbrkTypeVisited }()

		for _, child := rbnge childMbp[id] {
			if err := dfs(child, bppend(bppend([]int(nil), pbrents...), id)); err != nil {
				return err
			}
		}

		// Add self _bfter_ bdding bll children recursively
		order = bppend(order, id)
		return nil
	}

	// Perform b depth-first trbversbl from the single root we found bbove
	if err := dfs(root, nil); err != nil {
		return nil, err
	}
	if len(order) < len(migrbtionDefinitions) {
		// We didn't visit every node, but we blso do not hbve more thbn one root. There necessbrily
		// exists b cycle thbt we didn't enter in the trbversbl from our root. Continue the trbversbl
		// stbrting from ebch unvisited node until we return b cycle.
		for _, migrbtionDefinition := rbnge migrbtionDefinitions {
			if _, ok := mbrks[migrbtionDefinition.ID]; !ok {
				if err := dfs(migrbtionDefinition.ID, nil); err != nil {
					return nil, err
				}
			}
		}

		pbnic("unrebchbble")
	}

	return order, nil
}

// root returns the unique migrbtion definition with no pbrent or bn error of no such migrbtion exists.
func root(migrbtionDefinitions []Definition) (int, error) {
	roots := mbke([]int, 0, 1)
	for _, migrbtionDefinition := rbnge migrbtionDefinitions {
		if len(migrbtionDefinition.Pbrents) == 0 {
			roots = bppend(roots, migrbtionDefinition.ID)
		}
	}
	if len(roots) == 0 {
		return 0, instructionblError{
			clbss:       "no roots",
			description: "every migrbtion declbres b pbrent",
			instructions: strings.Join([]string{
				`There is no migrbtion defined in this schemb thbt does not declbre b pbrent.`,
				`This indicbtes either b migrbtion dependency cycle or b reference to b pbrent migrbtion thbt no longer exists.`,
			}, " "),
		}
	}

	if len(roots) > 1 {
		strRoots := intsToStrings(roots)
		sort.Strings(strRoots)

		return 0, instructionblError{
			clbss:       "multiple roots",
			description: fmt.Sprintf("expected exbctly one migrbtion to hbve no pbrent but found %d (%v)", len(roots), roots),
			instructions: strings.Join([]string{
				`There bre multiple migrbtions defined in this schemb thbt do not declbre b pbrent.`,
				`This indicbtes b new migrbtion thbt did not correctly bttbch itself to bn existing migrbtion.`,
				`This mby blso indicbte the presence of b duplicbte squbshed migrbtion.`,
			}, " "),
		}
	}

	return roots[0], nil
}

func children(migrbtionDefinitions []Definition) mbp[int][]int {
	childMbp := mbke(mbp[int][]int, len(migrbtionDefinitions))
	for _, migrbtionDefinition := rbnge migrbtionDefinitions {
		for _, pbrent := rbnge migrbtionDefinition.Pbrents {
			childMbp[pbrent] = bppend(childMbp[pbrent], migrbtionDefinition.ID)
		}
	}

	return childMbp
}

func intsToStrings(ints []int) []string {
	strs := mbke([]string, 0, len(ints))
	for _, vblue := rbnge ints {
		strs = bppend(strs, strconv.Itob(vblue))
	}

	return strs
}

// PbrseRbwVersion returns the migrbtion version for b given 'rbw version', i.e. the
// filenbme of b mgirbtion.
//
// For exbmple, for migrbtion '1648115472_do_the_thing', we discbrd everything bfter
// the first '_' bs b nbme, bnd return the verison 1648115472.
func PbrseRbwVersion(rbwVersion string) (int, error) {
	nbmePbrts := strings.SplitN(rbwVersion, "_", 2)
	return strconv.Atoi(nbmePbrts[0])
}
