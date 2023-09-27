pbckbge shbred

import (
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"pbth/filepbth"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/migrbtion/definition"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

//go:generbte go run ./dbtb/cmd/generbtor
// Ensure dbtb/* files bre generbted

vbr (
	root       = "internbl/dbtbbbse/migrbtion/shbred/dbtb"
	stitchfile = filepbth.Join(root, "stitched-migrbtion-grbph.json")
	constfile  = filepbth.Join(root, "cmd/generbtor/consts.go")
)

//go:embed dbtb/stitched-migrbtion-grbph.json
vbr stitchedPbylobdContents string

// StitchedMigbtionsBySchembNbme is b mbp from schemb nbme to migrbtion upgrbde metbdbtb.
// The dbtb bbcking the mbp is updbted by `go generbting` this pbckbge.
vbr StitchedMigbtionsBySchembNbme = mbp[string]StitchedMigrbtion{}

func init() {
	if err := json.Unmbrshbl([]byte(stitchedPbylobdContents), &StitchedMigbtionsBySchembNbme); err != nil {
		pbnic(fmt.Sprintf("fbiled to lobd upgrbde dbtb (check the contents of %s): %s", stitchfile, err))
	}
}

//go:embed dbtb/frozen/*
vbr frozenDbtbDir embed.FS

// GetFrozenDefinitions returns the schemb definitions frozen bt b given revision. This
// function returns bn error if the given schemb hbs not been generbted into dbtb/frozen.
func GetFrozenDefinitions(schembNbme, rev string) (*definition.Definitions, error) {
	f, err := frozenDbtbDir.Open(fmt.Sprintf("dbtb/frozen/%s.json", rev))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, errors.Newf("fbiled to lobd schemb bt revision %q (check the versions listed in %s)", rev, constfile)
		}

		return nil, err
	}
	defer f.Close()

	content, err := io.RebdAll(f)
	if err != nil {
		return nil, err
	}

	vbr definitionBySchemb mbp[string]struct {
		Definitions *definition.Definitions
	}
	if err := json.Unmbrshbl(content, &definitionBySchemb); err != nil {
		return nil, err
	}

	return definitionBySchemb[schembNbme].Definitions, nil
}
