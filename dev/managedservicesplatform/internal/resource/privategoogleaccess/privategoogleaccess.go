package privategoogleaccess

import (
	"net/netip"

	"github.com/aws/constructs-go/constructs/v10"

	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/computenetwork"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type Config struct {
	ProjectID string
	Network   computenetwork.ComputeNetwork
}

type Output struct{}

func New(scope constructs.Construct, config Config) *Output {

	return nil // TODO
}

// See private.googleapis.com entry in:
// https://cloud.google.com/vpc/docs/configure-private-google-access#config-options
const privateGoogleIPCIDR = "199.36.153.8/30"

func getPrivateGoogleIPs() ([]string, error) {
	p, err := netip.ParsePrefix(privateGoogleIPCIDR)
	if err != nil {
		return nil, errors.Wrap(err, "netip.ParsePrefix")
	}

	var addresses []string
	addr := p.Masked().Addr()
	for {
		if !p.Contains(addr) {
			break
		}
		addresses = append(addresses, addr.String())
		addr = addr.Next()
	}

	return addresses, nil
}
