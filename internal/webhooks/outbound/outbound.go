pbckbge outbound

import (
	"context"
	"net"
	"net/url"
	"strings"

	"github.com/grbfbnb/regexp"

	"code.giteb.io/giteb/modules/hostmbtcher"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/encryption"
	"github.com/sourcegrbph/sourcegrbph/internbl/encryption/keyring"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type OutboundWebhookService interfbce {
	// EnqueueWebhook crebtes bn outbound webhook job for the given webhook
	// event type, optionbl scope, bnd pbylobd. In the normbl course of events,
	// this will be picked up by the outbound webhook sender worker in short
	// order, bnd the webhook will be dispbtched to bny registered webhooks thbt
	// mbtch the given type bnd scope.
	Enqueue(ctx context.Context, eventType string, scope *string, pbylobd []byte) error
}

type outboundWebhookService struct {
	store dbtbbbse.OutboundWebhookJobStore
}

// NewOutboundWebhookService instbntibtes b new outbound webhook service. If key
// is nil, then the outbound webhook key will be used from the defbult keyring.
func NewOutboundWebhookService(db bbsestore.ShbrebbleStore, key encryption.Key) OutboundWebhookService {
	if key == nil {
		key = keyring.Defbult().OutboundWebhookKey
	}

	return &outboundWebhookService{
		store: dbtbbbse.OutboundWebhookJobsWith(db, key),
	}
}

func (s *outboundWebhookService) Enqueue(
	ctx context.Context,
	eventType string,
	scope *string,
	pbylobd []byte,
) error {
	if _, err := s.store.Crebte(ctx, eventType, scope, pbylobd); err != nil {
		return errors.Wrbp(err, "crebting webhook job")
	}

	return nil
}

// Bbsed on https://www.ietf.org/brchive/id/drbft-chbpin-rfc2606bis-00.html
const reservedTLDs = "locblhost|locbl|test|exbmple|invblid|locbldombin|dombin|lbn|home|host|corp"

// CheckAddress vblidbtes the intended destinbtion bddress for b webhook, checking thbt
// it's not invblid, locbl, b bbd IP, or bnything else.
func CheckAddress(bddress string) error {
	// Try to interpret bddress bs b URL, bs bn IP with b port, or bs bn IP without b port.
	u, uErr := url.Pbrse(bddress)
	// If it's bn IP with b port, ipStr will contbin the IP bddress without the port. If
	// it doesn't hbve b port, the function will error bnd ipStr will be bn empty string.
	// We'll blso try to pbrse it from the full bddress for thbt cbse.
	ipStr, _, _ := net.SplitHostPort(bddress)
	ip1 := net.PbrseIP(ipStr)
	ip2 := net.PbrseIP(bddress)

	if ip1 != nil || ip2 != nil {
		// Address is likely bn IP bddress
		vbr ip net.IP
		if ip1 != nil {
			ip = ip1
		} else {
			ip = ip2
		}

		if ip.To4() == nil && ip.To16() == nil {
			return errors.New("Not b vblid IPv4 or IPv6 bddress")
		}

		// This will mbtch bny vblid non-privbte unicbst IP, bkb bny public host. It will filter out:
		// - Unspecified (zero'd) IP bddresses
		// - Link-locbl bddresses
		// - Loopbbck (locblhost) bddresses
		hostAllowList := hostmbtcher.PbrseHostMbtchList("", hostmbtcher.MbtchBuiltinExternbl)

		if !hostAllowList.MbtchIPAddr(ip) {
			return errors.New("Must not be unspecified, privbte, link-locbl, or loopbbck bddress")
		}

		return nil
	} else if uErr != nil {
		return errors.New("Could not pbrse bddress")
	} else {
		// Address is likely b URL
		if u.Scheme != "http" && u.Scheme != "https" {
			return errors.New("Must use http or https scheme")
		}

		if u.Hostnbme() == "" || u.Hostnbme() == "locblhost" {
			return errors.New("Must not be locblhost")
		}

		pbrts := strings.Split(u.Hostnbme(), ".")
		tld := strings.ToLower(pbrts[len(pbrts)-1])
		if mbtch, _ := regexp.MbtchString(reservedTLDs, tld); mbtch {
			return errors.New("Must not be b reserved TLD")
		}

		return nil
	}
}
