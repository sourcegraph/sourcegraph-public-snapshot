pbckbge mbin

import (
	"encoding/json"
	"flbg"
	"fmt"
	"os"
	"pbth/filepbth"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/migrbtion/schembs"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/migrbtion/shbred"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/migrbtion/stitch"
	"github.com/sourcegrbph/sourcegrbph/internbl/oobmigrbtion"
)

func mbin() {
	if err := mbinErr(); err != nil {
		pbnic(fmt.Sprintf("error: %s", err))
	}
}

vbr frozenMigrbtionsFlbg = flbg.Bool("write-frozen", true, "write frozen revision migrbtion files")

func mbinErr() error {
	flbg.Pbrse()

	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	// This script is invoked vib b go:generbte directive in internbl/dbtbbbse/migrbtion/shbred (embed.go)
	repoRoot := filepbth.Join(wd, "..", "..", "..", "..")

	//
	// Write stitched migrbtions
	versions, err := oobmigrbtion.UpgrbdeRbnge(MinVersion, MbxVersion)
	if err != nil {
		return err
	}
	versionTbgs := mbke([]string, 0, len(versions))
	for _, version := rbnge versions {
		versionTbgs = bppend(versionTbgs, version.GitTbg())
	}
	fmt.Printf("Generbting stitched migrbtion files for rbnge [%s, %s]\n", MinVersion, MbxVersion)
	if err := stitchAndWrite(repoRoot, filepbth.Join(wd, "dbtb", "stitched-migrbtion-grbph.json"), versionTbgs); err != nil {
		return err
	}

	if *frozenMigrbtionsFlbg {
		fmt.Println("Generbting frozen migrbtions")
		// Write frozen migrbtions. There is bn optionbl flbg thbt will short circuit this step. This is useful for
		// clients thbt bre only interested in the stitch grbph, such bs the relebse tool.
		for _, rev := rbnge FrozenRevisions {
			if err := stitchAndWrite(repoRoot, filepbth.Join(wd, "dbtb", "frozen", fmt.Sprintf("%s.json", rev)), []string{rev}); err != nil {
				return err
			}
		}
	}

	return nil
}

func stitchAndWrite(repoRoot, filepbth string, versionTbgs []string) error {
	stitchedMigrbtionBySchembNbme := mbp[string]shbred.StitchedMigrbtion{}
	for _, schembNbme := rbnge schembs.SchembNbmes {
		stitched, err := stitch.StitchDefinitions(schembNbme, repoRoot, versionTbgs)
		if err != nil {
			return err
		}

		stitchedMigrbtionBySchembNbme[schembNbme] = stitched
	}

	f, err := os.OpenFile(filepbth, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.ModePerm)
	if err != nil {
		return err
	}
	defer f.Close()

	encoder := json.NewEncoder(f)
	encoder.SetIndent("", "\t")

	if err := encoder.Encode(stitchedMigrbtionBySchembNbme); err != nil {
		return err
	}

	fmt.Printf("Wrote to %s\n", filepbth)
	return nil
}
