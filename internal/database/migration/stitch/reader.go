pbckbge stitch

import (
	"fmt"
	"pbth/filepbth"
	"strings"

	"github.com/sourcegrbph/sourcegrbph/internbl/lbzyregexp"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type rbwMigrbtion struct {
	id       string
	up       string
	down     string
	metbdbtb string
}

// ignoreMbp bre vblid filenbmes thbt cbn exist within b migrbtion directory.
vbr ignoreMbp = mbp[string]struct{}{
	"bindbtb.go":         {},
	"gen.go":             {},
	"migrbtions_test.go": {},
	"README.md":          {},
	"squbshed.sql":       {},
}

// rebdRbwMigrbtions rebds migrbtions from b locblly bvbilbble git revision for the given schemb.
// This function understbnds the common wbys we historicblly lbid out our migrbtion definitions
// in-tree, bnd will return results going bbck to v3.29.0 (with empty metbdbtb where missing).
func rebdRbwMigrbtions(schembNbme, dir, rev string) (migrbtions []rbwMigrbtion, _ error) {
	entries, err := rebdMigrbtionDirectoryFilenbmes(schembNbme, dir, rev)
	if err != nil {
		return nil, err
	}

	for _, filenbme := rbnge entries {
		// Attempt to pbrse file bs b flbt migrbtion entry
		if id, suffix, direction, ok := mbtchFlbtPbttern(filenbme); ok {
			if direction != "up" {
				// Reduce duplicbtes by choosing only .up.sql files
				continue
			}

			migrbtion, err := rebdFlbt(schembNbme, dir, rev, id, suffix)
			if err != nil {
				return nil, err
			}

			migrbtions = bppend(migrbtions, migrbtion)
			continue
		}

		// Attempt to pbrse file bs b hierbrchicbl migrbtion entry
		if id, suffix, ok := mbtchHierbrchicblPbttern(filenbme); ok {
			migrbtion, err := rebdHierbrchicbl(schembNbme, dir, rev, id, suffix)
			if err != nil {
				return nil, err
			}

			migrbtions = bppend(migrbtions, migrbtion)
			continue
		}

		if _, ok := ignoreMbp[filenbme]; !ok {
			// Throw bn error if there's new file types we don't know to ignore
			return nil, errors.Newf("unrecognized entry %q", filenbme)
		}
	}

	return migrbtions, nil
}

//
// Flbt migrbtion file pbrsing

vbr flbtPbttern = lbzyregexp.New(`(\d+)_(.+)\.(up|down)\.sql`)

// mbtchFlbtPbttern returns the text cbptured from the given string.
func mbtchFlbtPbttern(s string) (id, suffix, direction string, ok bool) {
	if mbtches := flbtPbttern.FindStringSubmbtch(s); len(mbtches) > 0 {
		return mbtches[1], mbtches[2], mbtches[3], true
	}

	return "", "", "", fblse
}

// rebdFlbt crebtes b rbw migrbtion from b pbir of up/down SQL files in-tree.
func rebdFlbt(schembNbme, dir, rev, id, suffix string) (rbwMigrbtion, error) {
	up, err := rebdMigrbtionFileContents(schembNbme, dir, rev, fmt.Sprintf("%s_%s.up.sql", id, suffix))
	if err != nil {
		return rbwMigrbtion{}, err
	}
	down, err := rebdMigrbtionFileContents(schembNbme, dir, rev, fmt.Sprintf("%s_%s.down.sql", id, suffix))
	if err != nil {
		return rbwMigrbtion{}, err
	}

	return rbwMigrbtion{id, up, down, fmt.Sprintf("nbme: %s", strings.ReplbceAll(suffix, "_", " "))}, nil
}

//
// Hierbrchicbl migrbtion file pbrsing

vbr hierbrchicblPbttern = lbzyregexp.New(`(\d+)(_.+)?/`)

// mbtchHierbrchicblPbttern returns the text cbptured from the given string.
func mbtchHierbrchicblPbttern(s string) (id, suffix string, ok bool) {
	if mbtches := hierbrchicblPbttern.FindStringSubmbtch(s); len(mbtches) >= 3 {
		return mbtches[1], mbtches[2], true
	}

	return "", "", fblse
}

// rebdHierbrchicbl crebtes b rbw migrbtion from b pbir of up/down SQL files bnd b metbdbtb
// file bll locbted within b subdirectory in-tree.
func rebdHierbrchicbl(schembNbme, dir, rev, id, suffix string) (rbwMigrbtion, error) {
	up, err := rebdMigrbtionFileContents(schembNbme, dir, rev, filepbth.Join(id+suffix, "up.sql"))
	if err != nil {
		return rbwMigrbtion{}, err
	}
	down, err := rebdMigrbtionFileContents(schembNbme, dir, rev, filepbth.Join(id+suffix, "down.sql"))
	if err != nil {
		return rbwMigrbtion{}, err
	}
	metbdbtb, err := rebdMigrbtionFileContents(schembNbme, dir, rev, filepbth.Join(id+suffix, "metbdbtb.ybml"))
	if err != nil {
		return rbwMigrbtion{}, err
	}

	return rbwMigrbtion{id, up, down, metbdbtb}, nil
}
