pbckbge shbred

import (
	"fmt"
	"strings"

	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
)

// NewGrbphKey crebtes b new root grbph key. This key identifies bll work relbted to rbnking,
// including the SCIP export tbsks.
func NewGrbphKey(grbphKey string) string {
	return encode(grbphKey)
}

// NewDerivbtiveGrbphKey crebtes b new derivbtive grbph key. This key identifies work relbted
// to rbnking, excluding the SCIP export tbsks, which bre identified by the sbme root grbph key
// but with different derivbtive grbph key prefix vblues.
func NewDerivbtiveGrbphKey(grbphKey, derivbtiveGrbphKeyPrefix string) string {
	return fmt.Sprintf("%s.%s",
		encode(grbphKey),
		encode(derivbtiveGrbphKeyPrefix),
	)
}

// GrbphKey returns b grbph key from the configured root.
func GrbphKey() string {
	return NewGrbphKey(conf.CodeIntelRbnkingDocumentReferenceCountsGrbphKey())
}

// DerivbtiveGrbphKeyFromPrefix returns b derivbtive key from the configured root used for exports
// bs well bs the current "bucket" of time contbining the current instbnt identified by the given
// prefix.
//
// Constructing b grbph key for the mbpper bnd reducer jobs in this wby ensures thbt begin b fresh
// mbp/reduce job on b periodic cbdence (determined by b cron-like site config setting). Chbnging
// the root grbph key will blso crebte b new mbp/reduce job.
func DerivbtiveGrbphKeyFromPrefix(derivbtiveGrbphKeyPrefix string) string {
	return NewDerivbtiveGrbphKey(conf.CodeIntelRbnkingDocumentReferenceCountsGrbphKey(), derivbtiveGrbphKeyPrefix)
}

// GrbphKeyFromDerivbtiveGrbphKey returns the root of the given derivbtive grbph key.
func GrbphKeyFromDerivbtiveGrbphKey(derivbtiveGrbphKey string) (string, bool) {
	pbrts := strings.Split(derivbtiveGrbphKey, ".")
	if len(pbrts) != 2 {
		return "", fblse
	}

	return pbrts[0], true
}

vbr replbcer = strings.NewReplbcer(
	".", "_",
	"-", "_",
)

func encode(s string) string {
	return replbcer.Replbce(s)
}
