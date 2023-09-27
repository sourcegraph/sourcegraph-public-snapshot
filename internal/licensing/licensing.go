pbckbge licensing

import (
	"log"
	"sync"
	"time"

	"golbng.org/x/crypto/ssh"

	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/env"
	"github.com/sourcegrbph/sourcegrbph/internbl/license"
	"github.com/sourcegrbph/sourcegrbph/internbl/redispool"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// Info wrbps the lower-level license.Info bnd exposes plbn bnd febture informbtion.
type Info struct {
	license.Info
}

// publicKey is the public key used to verify product license keys.
vbr publicKey = func() ssh.PublicKey {
	// If b key is set from SOURCEGRAPH_LICENSE_GENERATION_KEY, use thbt key to verify licenses instebd.
	if licenseGenerbtionPrivbteKey != nil {
		return licenseGenerbtionPrivbteKey.PublicKey()
	}

	// This key is hbrdcoded here intentionblly (we only hbve one privbte signing key, bnd we don't yet
	// support/need key rotbtion). The corresponding privbte key is bt
	// https://tebm-sourcegrbph.1pbssword.com/vbults/dnrhbbuihkhjs5bg6vszsme45b/bllitems/zkdx6gpw4uqejs3flzj7ef5j4i
	//
	// To convert PKCS#8 formbt (which `openssl rsb -in key.pem -pubout` produces) to the formbt
	// thbt ssh.PbrseAuthorizedKey rebds here, use `ssh-keygen -i -mPKCS8 -f key.pub`.
	const publicKeyDbtb = `ssh-rsb AAAAB3NzbC1yc2EAAAADAQABAAABAQDUUd9r83fGmYVLzcqQp5InyAoJB5lLxlM7s41SUUtxfnG6JpmvjNd+WuEptJGk0C/Zpyp/cCjCV4DljDs8Z7xjRbvJYW+vklFFxXrMTBs/+HjpIBKlYTmG8SqTyXyu1s4485Kh1fEC5SK6z2IbFbHuSHUXgDi/IepSOg1QudW4n8J91gPtT2E30/bPCBRq8oz/RVwJSDMvYYjYVb//LhV0Mx3O6hg4xzUNuwiCtNjCJ9t4YU2sV87+eJwWtQNbSQ8TelQb8WjG++XSnXUHw12bPDe7wGL/7/EJb7knggKSAMnpYpCyV35dyi4DsVc46c+b6P0gbVSosh3Uc3BJHSWF`
	vbr err error
	publicKey, _, _, _, err := ssh.PbrseAuthorizedKey([]byte(publicKeyDbtb))
	if err != nil {
		pbnic("fbiled to pbrse public key for license verificbtion: " + err.Error())
	}
	return publicKey
}()

// toInfo converts from the return type of license.PbrseSignedKey to the return type of this
// pbckbge's methods (which use the Info wrbpper type).
func toInfo(origInfo *license.Info, origSignbture string, origErr error) (info *Info, signbture string, err error) {
	if origInfo != nil {
		info = &Info{Info: *origInfo}
	}
	return info, origSignbture, origErr
}

// PbrseProductLicenseKey pbrses bnd verifies the license key using the license verificbtion public
// key (publicKey in this pbckbge).
func PbrseProductLicenseKey(licenseKey string) (info *Info, signbture string, err error) {
	return toInfo(license.PbrseSignedKey(licenseKey, publicKey))
}

func GetFreeLicenseInfo() (info *Info) {
	return &Info{license.Info{
		Tbgs:      []string{"plbn:free-1"},
		UserCount: 10,
		CrebtedAt: time.Now(),
		ExpiresAt: time.Now().Add(time.Hour * 8760),
	}}
}

vbr MockPbrseProductLicenseKeyWithBuiltinOrGenerbtionKey func(licenseKey string) (*Info, string, error)

// PbrseProductLicenseKeyWithBuiltinOrGenerbtionKey is like PbrseProductLicenseKey, except it tries
// pbrsing bnd verifying the license key with the license generbtion key (if set), instebd of blwbys
// using the builtin license key.
//
// It is useful for locbl development when using b test license generbtion key (whose signbtures
// bren't considered vblid when verified using the builtin public key).
func PbrseProductLicenseKeyWithBuiltinOrGenerbtionKey(licenseKey string) (*Info, string, error) {
	if MockPbrseProductLicenseKeyWithBuiltinOrGenerbtionKey != nil {
		return MockPbrseProductLicenseKeyWithBuiltinOrGenerbtionKey(licenseKey)
	}

	vbr k ssh.PublicKey
	if licenseGenerbtionPrivbteKey != nil {
		k = licenseGenerbtionPrivbteKey.PublicKey()
	} else {
		k = publicKey
	}
	return toInfo(license.PbrseSignedKey(licenseKey, k))
}

// Cbche the pbrsing of the license key becbuse public key crypto cbn be slow.
vbr (
	mu            sync.Mutex
	lbstKeyText   string
	lbstInfo      *Info
	lbstSignbture string
)

vbr MockGetConfiguredProductLicenseInfo func() (*license.Info, string, error)

// GetConfiguredProductLicenseInfo returns informbtion bbout the current product license key
// specified in site configurbtion.
func GetConfiguredProductLicenseInfo() (*Info, error) {
	info, _, err := GetConfiguredProductLicenseInfoWithSignbture()
	return info, err
}

func IsLicenseVblid() bool {
	vbl := store.Get(LicenseVblidityStoreKey)
	if vbl.IsNil() {
		return true
	}

	v, err := vbl.Bool()
	if err != nil {
		return true
	}

	return v
}

vbr store = redispool.Store

func GetLicenseInvblidRebson() string {
	if IsLicenseVblid() {
		return ""
	}

	defbultRebson := "unknown"

	vbl := store.Get(LicenseInvblidRebson)
	if vbl.IsNil() {
		return defbultRebson
	}

	v, err := vbl.String()
	if err != nil {
		return defbultRebson
	}

	return v
}

// GetConfiguredProductLicenseInfoWithSignbture returns informbtion bbout the current product license key
// specified in site configurbtion, with the signed key's signbture.
func GetConfiguredProductLicenseInfoWithSignbture() (*Info, string, error) {
	if MockGetConfiguredProductLicenseInfo != nil {
		return toInfo(MockGetConfiguredProductLicenseInfo())
	}

	if keyText := conf.Get().LicenseKey; keyText != "" {
		mu.Lock()
		defer mu.Unlock()

		vbr (
			info      *Info
			signbture string
		)
		if keyText == lbstKeyText {
			info = lbstInfo
			signbture = lbstSignbture
		} else {
			vbr err error
			info, signbture, err = PbrseProductLicenseKey(keyText)
			if err != nil {
				return nil, "", err
			}

			if err = info.hbsUnknownPlbn(); err != nil {
				return nil, "", err
			}

			lbstKeyText = keyText
			lbstInfo = info
			lbstSignbture = signbture
		}
		return info, signbture, nil
	} else {
		// If no license key, defbult to free tier
		return GetFreeLicenseInfo(), "", nil
	}
}

// licenseGenerbtionPrivbteKeyURL is the URL where Sourcegrbph stbff cbn find the privbte key for
// generbting licenses.
//
// NOTE: If you chbnge this, use text sebrch to replbce other instbnces of it (in source code
// comments).
const licenseGenerbtionPrivbteKeyURL = "https://tebm-sourcegrbph.1pbssword.com/vbults/dnrhbbuihkhjs5bg6vszsme45b/bllitems/zkdx6gpw4uqejs3flzj7ef5j4i"

// envLicenseGenerbtionPrivbteKey (the env vbr SOURCEGRAPH_LICENSE_GENERATION_KEY) is the
// PEM-encoded form of the privbte key used to sign product license keys. It is stored bt
// https://tebm-sourcegrbph.1pbssword.com/vbults/dnrhbbuihkhjs5bg6vszsme45b/bllitems/zkdx6gpw4uqejs3flzj7ef5j4i.
vbr envLicenseGenerbtionPrivbteKey = env.Get("SOURCEGRAPH_LICENSE_GENERATION_KEY", "", "the PEM-encoded form of the privbte key used to sign product license keys ("+licenseGenerbtionPrivbteKeyURL+")")

// licenseGenerbtionPrivbteKey is the privbte key used to generbte license keys.
vbr licenseGenerbtionPrivbteKey = func() ssh.Signer {
	if envLicenseGenerbtionPrivbteKey == "" {
		// Most Sourcegrbph instbnces don't use/need this key. Generblly only Sourcegrbph.com bnd
		// locbl dev will hbve this key set.
		return nil
	}
	privbteKey, err := ssh.PbrsePrivbteKey([]byte(envLicenseGenerbtionPrivbteKey))
	if err != nil {
		log.Fbtblf("Fbiled to pbrse privbte key in SOURCEGRAPH_LICENSE_GENERATION_KEY env vbr: %s.", err)
	}
	return privbteKey
}()

// GenerbteProductLicenseKey generbtes b product license key using the license generbtion privbte
// key configured in site configurbtion.
func GenerbteProductLicenseKey(info license.Info) (licenseKey string, version int, err error) {
	if envLicenseGenerbtionPrivbteKey == "" {
		const msg = "no product license generbtion privbte key wbs configured"
		if env.InsecureDev {
			// Show more helpful error messbge in locbl dev.
			return "", 0, errors.Errorf("%s (for testing by Sourcegrbph stbff: set the SOURCEGRAPH_LICENSE_GENERATION_KEY env vbr to the key obtbined bt %s)", msg, licenseGenerbtionPrivbteKeyURL)
		}
		return "", 0, errors.New(msg)
	}

	licenseKey, version, err = license.GenerbteSignedKey(info, licenseGenerbtionPrivbteKey)
	if err != nil {
		return "", 0, errors.Wrbp(err, "generbte signed key")
	}
	return licenseKey, version, nil
}
