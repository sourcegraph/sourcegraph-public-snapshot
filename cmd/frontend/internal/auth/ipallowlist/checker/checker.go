package checker

import (
	"bytes"
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"
	"text/template"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/schema"
)

var (
	// stableUserIpAllowlist will be merged with the user ip allowlist from site config.
	stableUserIpAllowlist = func() []string {
		if allowlist := os.Getenv("AUTH_ALLOWED_IP_ADDRESS_STABLE_USER_IP_ALLOWLIST"); allowlist != "" {
			addr := strings.Split(allowlist, ",")
			for i := range addr {
				addr[i] = strings.TrimSpace(addr[i])
			}
			return addr
		}
		return nil
	}()
	// stableClientIpAllowlist will be merged with the client ip allowlist from site config.
	stableClientIpAllowlist = func() []string {
		if allowlist := os.Getenv("AUTH_ALLOWED_IP_ADDRESS_STABLE_CLIENT_IP_ALLOWLIST"); allowlist != "" {
			addr := strings.Split(allowlist, ",")
			for i := range addr {
				addr[i] = strings.TrimSpace(addr[i])
			}
			return addr
		}
		return nil
	}()
	// stableTrustedClientIpAllowlist will be merged with the trusted client ip allowlist from site config.
	//
	// Use case:
	//  - Cloud needs to configure IP allowlist that are managed by Cloud infra, and we don't user to override it, e.g., gke pod cidr, executors ip range.
	stableTrustedClientIpAllowlist = func() []string {
		if allowlist := os.Getenv("AUTH_ALLOWED_IP_ADDRESS_STABLE_TRUSTED_CLIENT_IP_ALLOWLIST"); allowlist != "" {
			addr := strings.Split(allowlist, ",")
			for i := range addr {
				addr[i] = strings.TrimSpace(addr[i])
			}
			return addr
		}
		return nil
	}()
	defaultTrustedClientIpAllowlist = []string{"127.0.0.1"}

	// configResolver is a cached resolver for the current configuration.
	// If empty, implies it is disabled.
	//
	// It will attempt to parse all config with best effort, and
	// attach a multi-error in the struct field `err` if any error occurred.
	// The caller should decide what to do with the error.
	configResolver = conf.Cached(getConfig)
)

func init() {
	conf.ContributeValidator(validateSiteConfig)
}

// Checker implements IP allowlist checking and subscription to configuration changes.
type Checker struct {
	logger log.Logger
}

// config is the computed internal config from site config.
// use configResolver() to get the current config.
type config struct {
	authorizedClientIps    []net.IP
	authorizedClientRanges []net.IPNet
	trustedClientIps       []net.IP
	trustedClientRanges    []net.IPNet
	authorizedUserIps      []net.IP
	authorizedUserRanges   []net.IPNet
	userHeaders            []string
	errorMessageTmpl       *template.Template

	// err indicates error occurred while parsing the configuration.
	err error
}

// New returns a new Checker.
func New(logger log.Logger) (*Checker, error) {
	logger = logger.Scoped("ipAllowlistChecker")

	cfg := conf.Get()
	if cfg == nil {
		return nil, errors.New("no config available")
	}

	if cfg := getConfig(); cfg != nil {
		if cfg.err != nil {
			logger.Error("site config 'auth.allowedIpAddress' contains error, please resolve it.", log.Error(cfg.err))
		}
	}

	return &Checker{
		logger: logger,
	}, nil
}

type unauthorizedErrorContext struct {
	Error  string
	UserIP string
}

// IsAuthorized returns an error if the given request is not authorized to access
// based on the current configuration.
func (c *Checker) IsAuthorized(req *http.Request) error {
	cfg := configResolver()
	if cfg == nil {
		return nil
	}

	// we only log the error in runtime, otherwise the site will be broken completely
	// also, we have validation in place to prevent invalid config from being applied in the first place
	if cfg.err != nil {
		c.logger.Error("site config 'auth.allowedIpAddress' contains error, please resolve it.", log.Error(cfg.err))
	}

	return isAuthorized(*cfg, req)
}

func isAuthorized(cfg config, req *http.Request) error {
	clientIp, userIp, err := getIP(req, cfg.userHeaders)
	if err != nil {
		return errors.Wrap(err, "get user ip")
	}

	// trusted client ip is allowed to bypass further checks
	if ok := containsIP(clientIp, cfg.trustedClientIps, cfg.trustedClientRanges); ok {
		return nil
	}

	if ok := containsIP(clientIp, cfg.authorizedClientIps, cfg.authorizedClientRanges); !ok {
		return renderUnAuthorizedErrorMessage(userIp, cfg.errorMessageTmpl)
	}

	if ok := containsIP(userIp, cfg.authorizedUserIps, cfg.authorizedUserRanges); !ok {
		return renderUnAuthorizedErrorMessage(userIp, cfg.errorMessageTmpl)
	}

	return nil
}

func renderUnAuthorizedErrorMessage(ip net.IP, tmpl *template.Template) error {
	msg := fmt.Sprintf("You are not allowed to access this Sourcegraph instance: %q", ip.String())
	if tmpl == nil {
		return errors.Newf(msg)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, unauthorizedErrorContext{
		Error:  msg,
		UserIP: ip.String(),
	}); err != nil {
		return errors.Wrap(err, "execute error message template")
	}
	return errors.Newf(buf.String())
}

// getIP returns the IP address of connected client and the user IP address.
//
// precedence of user IP (highest to lowest):
//   - user ip request headers
//   - remote addr (connected client ip)
func getIP(req *http.Request, userHeaders []string) (clientIp net.IP, userIp net.IP, err error) {
	clientIpRaw, _, err := net.SplitHostPort(req.RemoteAddr)
	if err != nil {
		return nil, nil, errors.Wrap(err, "parse remote addr")
	}
	clientIp = net.ParseIP(clientIpRaw)
	if clientIp == nil {
		return nil, nil, errors.Newf("invalid remote addr: %q", clientIpRaw)
	}

	var userIpRaw string
	for _, header := range userHeaders {
		ip := req.Header.Get(header)
		ips := strings.Split(ip, ",")
		if len(ips) > 0 && ips[0] != "" {
			userIpRaw = strings.TrimSpace(ips[0])
			break
		}
	}
	// if no user ip header is found, fallback to remote addr
	if userIpRaw == "" {
		return clientIp, clientIp, nil
	}

	userIp = net.ParseIP(userIpRaw)
	if userIp == nil {
		return nil, nil, errors.Newf("invalid user ip: %q", userIp)
	}
	return clientIp, userIp, nil
}

// containsIP returns true if the given IP is authorized.
// If no IPs or ranges are configured, all IPs are authorized.
func containsIP(addr net.IP, authorizedIps []net.IP, authorizedRanges []net.IPNet) bool {
	if len(authorizedIps) == 0 && len(authorizedRanges) == 0 {
		return true
	}

	for _, authorizedIP := range authorizedIps {
		if authorizedIP.Equal(addr) {
			return true
		}
	}

	for _, authorizedRange := range authorizedRanges {
		if authorizedRange.Contains(addr) {
			return true
		}
	}

	return false
}

// parseIPs parses a list of IP addresses and CIDR ranges.
func parseIPs(addresses []string) ([]net.IP, []net.IPNet, error) {
	var ips []net.IP
	var ipNets []net.IPNet

	var errs errors.MultiError

	for _, addr := range addresses {
		ip := net.ParseIP(addr)
		if ip != nil {
			ips = append(ips, ip)
		} else {
			_, ipNet, err := net.ParseCIDR(addr)
			if err != nil {
				errs = errors.Append(errs, errors.Wrapf(err, "invalid ip addr: %q", addr))
			} else {
				ipNets = append(ipNets, *ipNet)
			}
		}
	}
	return ips, ipNets, errs
}

func getConfig() *config {
	cfg := conf.Get()
	if cfg == nil {
		return nil
	}

	authCfg := cfg.AuthAllowedIpAddress
	if authCfg == nil || !authCfg.Enabled {
		return nil
	}

	return buildConfigFromSiteConfig(*authCfg)
}

func buildConfigFromSiteConfig(cfg schema.AuthAllowedIpAddress) *config {
	var errs errors.MultiError

	userIps, userIpRanges, err := parseIPs(append(cfg.UserIpAddress, stableUserIpAllowlist...))
	if err != nil {
		errs = errors.Append(errs, err)
	}

	clientIps, clientIpRanges, err := parseIPs(append(cfg.ClientIpAddress, stableClientIpAllowlist...))
	if err != nil {
		errs = errors.Append(errs, err)
	}

	// 🚨 SECURITY: we always allow localhost to access the instance, otherwise liveness prob will fail
	trustedClientIps, trustedClientIpRanges, err := parseIPs(concactSlices(cfg.TrustedClientIpAddress, stableTrustedClientIpAllowlist, defaultTrustedClientIpAllowlist))
	if err != nil {
		errs = errors.Append(errs, err)
	}

	var tmpl *template.Template
	if cfg.ErrorMessageTemplate != "" {
		tmpl, err = template.New("").Parse(cfg.ErrorMessageTemplate)
		if err != nil {
			errs = errors.Append(errs, err)
		}
	}

	return &config{
		authorizedUserIps:      userIps,
		authorizedUserRanges:   userIpRanges,
		authorizedClientIps:    clientIps,
		authorizedClientRanges: clientIpRanges,
		trustedClientIps:       trustedClientIps,
		trustedClientRanges:    trustedClientIpRanges,
		userHeaders:            cfg.UserIpRequestHeaders,
		errorMessageTmpl:       tmpl,
		err:                    errs,
	}
}

func concactSlices[T any](slices ...[]T) []T {
	var result []T
	for _, slice := range slices {
		result = append(result, slice...)
	}
	return result
}

func validateSiteConfig(confQuerier conftypes.SiteConfigQuerier) conf.Problems {
	cfg := confQuerier.SiteConfig().AuthAllowedIpAddress
	if cfg == nil {
		return nil
	}

	if len(cfg.UserIpAddress) > 0 {
		if _, _, err := parseIPs(cfg.UserIpAddress); err != nil {
			return conf.NewSiteProblems(fmt.Sprintf("`auth.allowedIpAddress.userIpAddress` is invalid: %s", err.Error()))
		}
	}

	if len(cfg.ClientIpAddress) > 0 {
		if _, _, err := parseIPs(cfg.ClientIpAddress); err != nil {
			return conf.NewSiteProblems(fmt.Sprintf("`auth.allowedIpAddress.clientIpAddress` is invalid: %s", err.Error()))
		}
	}

	if len(cfg.TrustedClientIpAddress) > 0 {
		if _, _, err := parseIPs(cfg.TrustedClientIpAddress); err != nil {
			return conf.NewSiteProblems(fmt.Sprintf("`auth.allowedIpAddress.trustedClientIpAddress` is invalid: %s", err.Error()))
		}
	}

	if cfg.ErrorMessageTemplate != "" {
		if _, err := template.New("").Parse(cfg.ErrorMessageTemplate); err != nil {
			return conf.NewSiteProblems(fmt.Sprintf("`auth.allowedIpAddress.errorMessageTemplate` is invalid: %s", err.Error()))
		}
	}
	return nil
}
