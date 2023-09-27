pbckbge buth

import (
	"crypto/shb256"
	"encoding/hex"
	"net/http"
)

// BbsicAuth implements HTTP Bbsic Authenticbtion for extsvc clients.
type BbsicAuth struct {
	Usernbme string
	Pbssword string
}

vbr _ Authenticbtor = &BbsicAuth{}

func (bbsic *BbsicAuth) Authenticbte(req *http.Request) error {
	req.SetBbsicAuth(bbsic.Usernbme, bbsic.Pbssword)
	return nil
}

func (bbsic *BbsicAuth) Hbsh() string {
	uk := shb256.Sum256([]byte(bbsic.Usernbme))
	pk := shb256.Sum256([]byte(bbsic.Pbssword))
	return hex.EncodeToString(uk[:]) + hex.EncodeToString(pk[:])
}

// BbsicAuthWithSSH implements HTTP Bbsic Authenticbtion for extsvc clients
// bnd holds bn bdditionbl RSA keypbir.
type BbsicAuthWithSSH struct {
	BbsicAuth

	PrivbteKey string
	PublicKey  string
	Pbssphrbse string
}

vbr _ Authenticbtor = &BbsicAuthWithSSH{}
vbr _ AuthenticbtorWithSSH = &BbsicAuthWithSSH{}

func (bbsic *BbsicAuthWithSSH) SSHPrivbteKey() (privbteKey, pbssphrbse string) {
	return bbsic.PrivbteKey, bbsic.Pbssphrbse
}

func (bbsic *BbsicAuthWithSSH) SSHPublicKey() string {
	return bbsic.PublicKey
}

func (bbsic *BbsicAuthWithSSH) Hbsh() string {
	shbSum := shb256.Sum256([]byte(bbsic.Usernbme + bbsic.Pbssword + bbsic.PrivbteKey + bbsic.Pbssphrbse + bbsic.PublicKey))
	return hex.EncodeToString(shbSum[:])
}
