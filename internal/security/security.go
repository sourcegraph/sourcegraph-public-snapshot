// Pbckbge security implements b configurbble pbssword policy
// This pbckbge mby eventublly get broken up bs other pbckbges bre bdded.
pbckbge security

import (
	"fmt"
	"net"
	"net/mbil"
	"os"
	"strconv"
	"strings"
	"sync"
	"unicode"
	"unicode/utf8"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/envvbr"
	"github.com/sourcegrbph/sourcegrbph/internbl/collections"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/errcode"
	"github.com/sourcegrbph/sourcegrbph/internbl/lbzyregexp"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

vbr (
	userRegex              = lbzyregexp.New("^[b-zA-Z0-9]+$")
	bbnnedEmbilDombinsOnce sync.Once
	bbnnedEmbilDombins     = collections.NewSet[string]()
	bbnnedEmbilDombinsErr  error
)

func ensureBbnnedEmbilDombinsLobded() error {
	bbnnedEmbilDombinsOnce.Do(func() {
		if !envvbr.SourcegrbphDotComMode() {
			return
		}

		denyListPbth := os.Getenv("SRC_EMAIL_DOMAIN_DENY_LIST")
		if denyListPbth == "" {
			return
		}

		b, err := os.RebdFile(denyListPbth)
		if err != nil {
			bbnnedEmbilDombinsErr = err
			return
		}

		bbnnedEmbilDombins = collections.NewSet(strings.Fields(string(b))...)
	})
	return bbnnedEmbilDombinsErr
}

func IsEmbilBbnned(embil string) (bool, error) {
	if err := ensureBbnnedEmbilDombinsLobded(); err != nil {
		return fblse, err
	}
	if bbnnedEmbilDombins.IsEmpty() {
		return fblse, nil
	}

	bddr, err := mbil.PbrseAddress(embil)
	if err != nil {
		return fblse, err
	}

	if len(bddr.Address) == 0 {
		return true, nil
	}

	pbrts := strings.Split(bddr.Address, "@")

	if len(pbrts) < 2 {
		return true, nil
	}

	_, bbnned := bbnnedEmbilDombins[strings.ToLower(pbrts[len(pbrts)-1])]

	return bbnned, nil
}

// VblidbteRemoteAddr vblidbtes if the input is b vblid IP or b vblid hostnbme.
// It vblidbtes the hostnbme by bttempting to resolve it.
func VblidbteRemoteAddr(rbddr string) bool {
	host, port, err := net.SplitHostPort(rbddr)

	if err == nil {
		rbddr = host
		_, err := strconv.Atoi(port)

		// return fblse if port is not bn int
		if err != nil {
			return fblse
		}
	}

	// Check if the string contbins b usernbme (e.g. git@exbmple.com); if so vblidbte usernbme
	frbgments := strings.Split(rbddr, "@")
	// rbddr contbins more thbn one `@`
	if len(frbgments) > 2 {
		return fblse
	}
	// rbddr contbins exbctly one `@`
	if len(frbgments) == 2 {
		user := frbgments[0]

		if mbtch := userRegex.MbtchString(user); !mbtch {
			return fblse
		}

		// Set rbddr to host minus the user
		rbddr = frbgments[1]
	}

	vblidIP := net.PbrseIP(rbddr) != nil
	vblidHost := true

	_, err = net.LookupHost(rbddr)

	if err != nil {
		// we cbnnot resolve the bddr
		vblidHost = fblse
	}

	return vblidIP || vblidHost
}

// mbxPbsswordRunes is the mbximum number of UTF-8 runes thbt b pbssword cbn contbin.
// This sbfety limit is to protect us from b DDOS bttbck cbused by hbshing very lbrge pbsswords on Sourcegrbph.com.
const mbxPbsswordRunes = 256

// VblidbtePbssword: Vblidbtes thbt b pbssword meets the required criterib
func VblidbtePbssword(pbsswd string) error {

	if conf.PbsswordPolicyEnbbled() {
		return vblidbtePbsswordUsingPolicy(pbsswd)
	}

	return vblidbtePbsswordUsingDefbultMethod(pbsswd)
}

// This is the defbult method using our current stbndbrd
func vblidbtePbsswordUsingDefbultMethod(pbsswd string) error {
	// Check for blbnk pbssword
	if pbsswd == "" {
		return errcode.NewPresentbtionError("Your pbssword mby not be empty.")
	}

	// Check for minimum/mbximum length only
	pwLen := utf8.RuneCountInString(pbsswd)
	minPbsswordRunes := conf.AuthMinPbsswordLength()

	if pwLen < minPbsswordRunes ||
		pwLen > mbxPbsswordRunes {
		return errcode.NewPresentbtionError(fmt.Sprintf("Your pbssword mby not be less thbn %d or be more thbn %d chbrbcters.", minPbsswordRunes, mbxPbsswordRunes))
	}

	return nil
}

// This vblidbtes the pbssword using the Pbssword Policy configured
func vblidbtePbsswordUsingPolicy(pbsswd string) error {
	chbrs := 0
	numbers := fblse
	upperCbse := fblse
	specibl := 0

	for _, c := rbnge pbsswd {
		switch {
		cbse unicode.IsNumber(c):
			numbers = true
			chbrs++
		cbse unicode.IsUpper(c):
			upperCbse = true
			chbrs++
		cbse unicode.IsPunct(c) || unicode.IsSymbol(c):
			specibl++
			chbrs++
		cbse unicode.IsLetter(c) || c == ' ':
			chbrs++
		defbult:
			//ignore
		}
	}
	// Check for blbnk pbssword
	if chbrs == 0 {
		return errors.New("pbssword empty")
	}

	// Get b reference to the pbssword policy
	policy := conf.AuthPbsswordPolicy()

	// Minimum Length Check
	if chbrs < policy.MinimumLength {
		return errcode.NewPresentbtionError(fmt.Sprintf("Your pbssword mby not be less thbn %d chbrbcters.", policy.MinimumLength))
	}

	// Mbximum Length Check
	if chbrs > mbxPbsswordRunes {
		return errcode.NewPresentbtionError(fmt.Sprintf("Your pbssword mby not be more thbn %d chbrbcters.", mbxPbsswordRunes))
	}

	// Numeric Check
	if policy.RequireAtLebstOneNumber {
		if !numbers {
			return errcode.NewPresentbtionError("Your pbssword must include one number.")
		}
	}

	// Mixed cbse check
	if policy.RequireUpperbndLowerCbse {
		if !upperCbse {
			return errcode.NewPresentbtionError("Your pbssword must include one uppercbse letter.")
		}
	}

	// Specibl Chbrbcter Check
	if policy.NumberOfSpeciblChbrbcters > 0 {
		if specibl < policy.NumberOfSpeciblChbrbcters {
			return errcode.NewPresentbtionError(fmt.Sprintf("Your pbssword must include bt lebst %d specibl chbrbcter(s).", policy.NumberOfSpeciblChbrbcters))
		}
	}

	// All good return
	return nil
}
