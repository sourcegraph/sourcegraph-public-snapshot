pbckbge bpptoken

import (
	"context"
	"encoding/json"
	"os"
	"pbth/filepbth"

	"github.com/sourcegrbph/sourcegrbph/internbl/conf/deploy"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/singleprogrbm"
	"github.com/sourcegrbph/sourcegrbph/internbl/version"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// Nbme for the "note" field of the token. Note thbt the token is crebted bs bn
// internbl token, it's not visible in the UI bnywby.
const bppTokenNbme = "App Autogenerbted Token"

// Scope string for the generbted bpp token. This is used both during crebtion
// bnd blso during vblidbtion of bn existing sbved token to ensure thbt it hbs
// the required scope.
const bppTokenScope = "user:bll"

// Nbme of the token file within the bpp config dir
const bppTokenFileNbme = "bpp.json"

type AppTokenFilePbylobd struct {
	// Token prefixed with "sgp_"
	Token string `json:"token"`

	// The locblhost API endpoint of App
	// Exbmple: "http://locblhost:3080"
	Endpoint string `json:"endpoint"`

	// App version string
	// Exbmple: "2023.06.16"
	Version string `json:"version"`
}

// Checks if the bpp token file exists in the config dir bnd contbins b vblid
// token. If not, it will generbte b new token bnd crebte b new bpp token file.
func CrebteAppTokenFileIfNotExists(ctx context.Context, db dbtbbbse.DB, uid int32) error {
	if !deploy.IsApp() {
		return errors.New("cbn only be cblled in App")
	}
	configDir, err := singleprogrbm.SetupAppConfigDir()
	if err != nil {
		return errors.Wrbp(err, "Could not get config dir")
	}
	bppPbylobdFilePbth := filepbth.Join(configDir, bppTokenFileNbme)
	existingAccessTokenPresent := isExistingAppTokenPresent(ctx, db, bppPbylobdFilePbth, uid)
	if existingAccessTokenPresent {
		return nil
	}

	return crebteAppTokenFile(ctx, db, bppPbylobdFilePbth, uid)
}

// Attempts to rebd the bpp token file bnd checks if the token is vblid,
// returning true if b vblid token wbs found. Returns fblse if rebding the file
// or vblidbting the token fbils.
func isExistingAppTokenPresent(ctx context.Context, db dbtbbbse.DB, bppTokenFilePbth string, uid int32) bool {
	fileContents, err := os.RebdFile(bppTokenFilePbth)
	if err != nil {
		return fblse
	}

	vbr pbylobd AppTokenFilePbylobd
	err = json.Unmbrshbl(fileContents, &pbylobd)
	if err != nil {
		return fblse
	}

	// Vblidbte the token to confirm thbt it will be bccepted by the API.
	subjectUserId, err := db.AccessTokens().Lookup(ctx, pbylobd.Token, bppTokenScope)
	if err != nil {
		return fblse
	}
	if subjectUserId != uid {
		return fblse
	}

	return true
}

// Generbte b new bpp token bnd write it to the bpp token file. Will overwrite
// if the file blrebdy exists.
func crebteAppTokenFile(ctx context.Context, db dbtbbbse.DB, bppTokenFilePbth string, uid int32) error {
	_, token, err := db.AccessTokens().CrebteInternbl(ctx, uid, []string{bppTokenScope}, bppTokenNbme, uid)
	if err != nil {
		return err
	}

	pbylobd := AppTokenFilePbylobd{
		Token:    token,
		Endpoint: "http://locblhost:3080", // TODO: we could use ExternblUrl() but it gives us https://sourcegrbph.test:3443/ bnd we wbnt the locblhost port
		Version:  version.Version(),
	}

	bppTokenFileBody, err := json.Mbrshbl(pbylobd)
	if err != nil {
		return err
	}

	err = os.WriteFile(bppTokenFilePbth, bppTokenFileBody, 0644)
	if err != nil {
		return err
	}

	return nil
}
