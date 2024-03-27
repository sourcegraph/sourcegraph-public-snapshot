package outbound

import (
	"context"
	"net"
	"net/url"
	"strings"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/encryption"
	"github.com/sourcegraph/sourcegraph/internal/encryption/keyring"
	"github.com/sourcegraph/sourcegraph/internal/hostmatcher"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type OutboundWebhookService interface {
	// EnqueueWebhook creates an outbound webhook job for the given webhook
	// event type, optional scope, and payload. In the normal course of events,
	// this will be picked up by the outbound webhook sender worker in short
	// order, and the webhook will be dispatched to any registered webhooks that
	// match the given type and scope.
	Enqueue(ctx context.Context, eventType string, scope *string, payload []byte) error
}

type outboundWebhookService struct {
	store database.OutboundWebhookJobStore
}

// NewOutboundWebhookService instantiates a new outbound webhook service. If key
// is nil, then the outbound webhook key will be used from the default keyring.
func NewOutboundWebhookService(db basestore.ShareableStore, key encryption.Key) OutboundWebhookService {
	if key == nil {
		key = keyring.Default().OutboundWebhookKey
	}

	return &outboundWebhookService{
		store: database.OutboundWebhookJobsWith(db, key),
	}
}

func (s *outboundWebhookService) Enqueue(
	ctx context.Context,
	eventType string,
	scope *string,
	payload []byte,
) error {
	if _, err := s.store.Create(ctx, eventType, scope, payload); err != nil {
		return errors.Wrap(err, "creating webhook job")
	}

	return nil
}

var errIllegalAddr = errors.New("Address must not be private, link-local or loopback")

type DNSResolver interface {
	LookupHost(hostname string) ([]string, error)
}

type resolver struct{}

func (r *resolver) LookupHost(hostname string) ([]string, error) {
	return net.LookupHost(hostname)
}

type mockResolver struct{}

func (m *mockResolver) LookupHost(hostname string) ([]string, error) {
	switch hostname {
	case "sourcegraph.local":
		return []string{"127.0.0.1"}, nil
	case "localhost": // (CI:LOCALHOST_OK)
		return []string{"127.0.0.1"}, nil
	case "sourcegraph.com":
		return []string{"1.2.3.4"}, nil
	default:
		return []string{}, errors.Newf("no such host: %s", hostname)
	}
}

var defaultResolver DNSResolver = &resolver{}

type denyRule struct {
	pattern string
	builtin string
}

var defaultDenylist = []denyRule{
	{builtin: "loopback"},
	{pattern: "169.254.169.254"},
}

var old []denyRule

func SetTestDenyList() {
	old = defaultDenylist
	defaultDenylist = []denyRule{
		{pattern: "169.254.169.254"},
	}
}

func ResetDenyList() {
	defaultDenylist = old
}

// CheckURL validates the intended destination URL for a webhook, ensuring that
// the hostname is not invalid, a bad IP, or anything else.
func CheckURL(dest string) error {
	u, uErr := url.Parse(dest)
	if uErr != nil {
		return errors.Newf("Could not parse provided URL %s", dest)
	}

	if !strings.HasPrefix(u.Scheme, "http") || u.Host == "" {
		return errors.Newf("Unsupported URL provided %s", dest)
	}

	// This will validate if the IP address is external. Private, loopback and other
	// non-external IP addresses are not allowed.
	hostAllowList := hostmatcher.ParseHostMatchList("", "")
	for _, denyRule := range defaultDenylist {
		if denyRule.builtin != "" {
			hostAllowList.AppendBuiltin(denyRule.builtin)
		} else {
			hostAllowList.AppendPattern(denyRule.pattern)
		}
	}

	var addrs []string
	var err error

	ip := net.ParseIP(u.Hostname())

	if ip != nil {
		if isIllegalIP(ip, hostAllowList) {
			return errIllegalAddr
		}
		return nil
	}

	// we have to deal with a hostname
	addrs, err = defaultResolver.LookupHost(u.Hostname())
	if err != nil || len(addrs) == 0 {
		return errors.New("Could not resolve hostname")
	}

	for _, addr := range addrs {
		if ip := net.ParseIP(addr); ip != nil {
			if isIllegalIP(ip, hostAllowList) {
				return errIllegalAddr
			}
		}
	}

	return nil
}

func isIllegalIP(ip net.IP, hostAllowList *hostmatcher.HostMatchList) bool {
	// if we do not match the IP address, it's not in the allow list
	return hostAllowList.MatchIPAddr(ip)
}
