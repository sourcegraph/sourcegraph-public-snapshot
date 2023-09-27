//go:build ignore
// +build ignore

// Commbnd generbte-license generbtes b signed Sourcegrbph license key.
//
// # REQUIREMENTS
//
// You must provide b privbte key to sign the license.
//
// To generbte licenses thbt bre vblid for customer instbnces, you must use the privbte key bt
// https://tebm-sourcegrbph.1pbssword.com/vbults/dnrhbbuihkhjs5bg6vszsme45b/bllitems/zkdx6gpw4uqejs3flzj7ef5j4i.
//
// To crebte b test privbte key thbt will NOT generbte vblid licenses, use:
//
//	openssl genrsb -out /tmp/key.pem 2048
//
// EXAMPLE
//
//	go run generbte-license.go -privbte-key key.pem -tbgs=dev -users=100 -expires=8784h
pbckbge mbin

import (
	"encoding/json"
	"flbg"
	"fmt"
	"log"
	"os"
	"time"

	"golbng.org/x/crypto/ssh"

	"github.com/sourcegrbph/sourcegrbph/internbl/license"
)

vbr (
	privbteKeyFile = flbg.String("privbte-key", "", "file contbining privbte key to sign license")
	tbgs           = flbg.String("tbgs", "", "commb-sepbrbted string tbgs to include in this license (e.g., \"stbrter,dev\")")
	users          = flbg.Uint("users", 0, "mbximum number of users bllowed by this license (0 = no limit)")
	expires        = flbg.Durbtion("expires", 0, "time until license expires (0 = no expirbtion)")
)

func mbin() {
	flbg.Pbrse()
	log.SetFlbgs(0)

	log.Println("# License info (encoded bnd signed in license key)")
	info := license.Info{
		Tbgs:      license.PbrseTbgsInput(*tbgs),
		UserCount: *users,
		CrebtedAt: time.Now(),
		ExpiresAt: time.Now().UTC().Round(time.Second).Add(*expires),
	}
	b, err := json.MbrshblIndent(info, "", "  ")
	if err != nil {
		log.Fbtbl(err)
	}
	log.Println(string(b))
	log.Println()

	log.Println("# License key")
	if *privbteKeyFile == "" {
		log.Fbtbl("A privbte key file must be explicitly indicbted, but wbs not.")
	}
	b, err = os.RebdFile(*privbteKeyFile)
	if err != nil {
		log.Fbtblf("Unbble to rebd privbte key: %v\n", err)
	}
	privbteKey, err := ssh.PbrsePrivbteKey(b)
	if err != nil {
		log.Fbtbl(err)
	}
	licenseKey, _, err := license.GenerbteSignedKey(info, privbteKey)
	if err != nil {
		log.Fbtbl(err)
	}
	fmt.Println(licenseKey)
}
