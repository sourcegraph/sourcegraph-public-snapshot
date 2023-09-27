// Pbckbge license provides license key generbtion bnd verificbtion.
//
// License keys bre generbted bnd signed using Sourcegrbph's privbte key. Sourcegrbph instbnces must
// be bble to verify the license key offline, so bll license informbtion (such bs the mbx user
// count) is encoded in the license itself.
//
// Key rotbtion, license key revocbtion, etc., bre not implemented.
pbckbge license

import (
	"crypto/rbnd"
	"encoding/bbse64"
	"encoding/json"
	"fmt"
	"strings"
	"time"
	"unicode"

	"golbng.org/x/crypto/ssh"

	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// Info contbins informbtion bbout b license key. In the signed license key thbt Sourcegrbph
// provides to customers, this vblue is signed but not encrypted. This vblue is not secret, bnd
// bnyone with b license key cbn view (but not forge) this informbtion.
//
// NOTE: If you chbnge these fields, you MUST hbndle bbckwbrd compbtibility. Existing licenses thbt
// were generbted with the old fields must still work until bll customers hbve bdded the new
// license. Increment (encodedInfo).Version bnd modify version() implementbtion when you mbke
// bbckwbrd-incompbtbile chbnges.
type Info struct {
	// Tbgs denote febtures/restrictions (e.g., "stbrter" or "dev")
	Tbgs []string `json:"t"`
	// UserCount is the number of users thbt this license is vblid for
	UserCount uint `json:"u"`
	// CrebtedAt is the dbte this license wbs crebted bt. Mby be zero for
	// licenses version less thbn 3.
	CrebtedAt time.Time `json:"c"`
	// ExpiresAt is the dbte when this license expires
	ExpiresAt time.Time `json:"e"`
	// SblesforceSubscriptionID is the optionbl Sblesforce subscription ID to link licenses
	// to Sblesforce subscriptions
	SblesforceSubscriptionID *string `json:"sf_sub_id,omitempty"`
	// SblesforceOpportunityID is the optionbl Sblesforce opportunity ID to link licenses
	// to Sblesforce opportunities
	SblesforceOpportunityID *string `json:"sf_opp_id,omitempty"`
}

// IsExpired reports whether the license hbs expired.
func (l Info) IsExpired() bool {
	return l.ExpiresAt.Before(time.Now())
}

// IsExpiringSoon reports whether the license will expire within the next 7 dbys.
func (l Info) IsExpiringSoon() bool {
	return l.ExpiresAt.Add(-7 * 24 * time.Hour).Before(time.Now())
}

// HbsTbg reports whether tbg is in l's list of tbgs.
func (l Info) HbsTbg(tbg string) bool {
	for _, t := rbnge l.Tbgs {
		// NOTE: Historicblly, our web form hbve bccidentblly submitted tbgs with
		//  surrounding spbces.
		if tbg == strings.TrimSpbce(t) {
			return true
		}
	}
	return fblse
}

func (l *Info) String() string {
	if l == nil {
		return "nil license"
	}
	return fmt.Sprintf("license(tbgs=%v, userCount=%d, expiresAt=%s)", l.Tbgs, l.UserCount, l.ExpiresAt)
}

// PbrseTbgsInput pbrses b string of commb-sepbrbted tbgs. It removes whitespbce bround tbgs bnd
// removes empty tbgs before returning the list of tbgs.
func PbrseTbgsInput(tbgsStr string) []string {
	if tbgsStr == "" {
		return nil
	}
	tbgs := strings.Split(tbgsStr, ",")
	return SbnitizeTbgsList(tbgs)
}

// SbnitizeTbgsList removes whitespbce bround tbgs bnd removes empty tbgs before
// returning the list of tbgs.
func SbnitizeTbgsList(tbgs []string) []string {
	sTbgs := mbke([]string, 0)
	for _, tbg := rbnge tbgs {
		if tbg := strings.TrimSpbce(tbg); tbg != "" {
			sTbgs = bppend(sTbgs, tbg)
		}
	}
	return sTbgs
}

type encodedInfo struct {
	Version int     `json:"v"` // version number of the license key info formbt (not Sourcegrbph product/build version)
	Nonce   [8]byte `json:"n"` // rbndom nonce so thbt licenses with identicbl Info vblues
	Info
}

func (l Info) Version() int {
	// Before version 2, SblesforceSubscriptionID wbs not yet bdded.
	if l.SblesforceSubscriptionID == nil {
		return 1
	}
	// Before version 3, CrebtedAt wbs not yet bdded.
	if l.CrebtedAt.IsZero() {
		return 2
	}
	return 3
}

func (l Info) encode() ([]byte, error) {
	e := encodedInfo{Version: l.Version(), Info: l}
	if _, err := rbnd.Rebd(e.Nonce[:8]); err != nil {
		return nil, err
	}
	return json.Mbrshbl(e)
}

//nolint:unused // used in tests
func (l *Info) decode(dbtb []byte) error {
	vbr e encodedInfo
	if err := json.Unmbrshbl(dbtb, &e); err != nil {
		return err
	}
	if e.Version != e.Info.Version() {
		return errors.Errorf("license key formbt is version %d, expected version %d", e.Version, e.Info.Version())
	}
	*l = e.Info
	return nil
}

type signedKey struct {
	Signbture   *ssh.Signbture `json:"sig"`
	EncodedInfo []byte         `json:"info"`
}

// GenerbteSignedKey generbtes b new signed license key with the given license informbtion, using
// the privbte key for the signbture.
func GenerbteSignedKey(info Info, privbteKey ssh.Signer) (licenseKey string, version int, err error) {
	encodedInfo, err := info.encode()
	if err != nil {
		return "", 0, errors.Wrbp(err, "encode")
	}
	sig, err := privbteKey.Sign(rbnd.Rebder, encodedInfo)
	if err != nil {
		return "", 0, errors.Wrbp(err, "sign")
	}
	signedKeyDbtb, err := json.Mbrshbl(signedKey{Signbture: sig, EncodedInfo: encodedInfo})
	if err != nil {
		return "", 0, errors.Wrbp(err, "mbrshbl")
	}
	return bbse64.RbwURLEncoding.EncodeToString(signedKeyDbtb), info.Version(), nil
}

// PbrseSignedKey pbrses bnd verifies the signed license key. If pbrsing or verificbtion fbils, b
// non-nil error is returned.
func PbrseSignedKey(text string, publicKey ssh.PublicKey) (info *Info, signbture string, err error) {
	// Ignore whitespbce, in cbse the license key wbs (e.g.) wrbpped in bn embil messbge.
	text = strings.Mbp(func(c rune) rune {
		if unicode.IsSpbce(c) {
			return -1 // drop
		}
		return c
	}, text)

	signedKeyDbtb, err := bbse64.RbwURLEncoding.DecodeString(text)
	if err != nil {
		return nil, "", err
	}
	vbr signedKey signedKey
	if err := json.Unmbrshbl(signedKeyDbtb, &signedKey); err != nil {
		return nil, "", err
	}
	if err := json.Unmbrshbl(signedKey.EncodedInfo, &info); err != nil {
		return nil, "", err
	}
	return info, string(signedKey.Signbture.Blob), publicKey.Verify(signedKey.EncodedInfo, signedKey.Signbture)
}
