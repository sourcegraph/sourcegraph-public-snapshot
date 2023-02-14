package rds

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/rds/rdsutils"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// AuthProvider is an interface for getting an auth token for RDS IAM auth.
// It is an interface so that we can mock it in tests.
type AuthProvider interface {
	AuthToken(hostname string, port uint16, user string) (string, error)
}

var _ AuthProvider = &Auth{}

// Auth implements the AuthProvider interface for getting an auth token for RDS IAM auth.
// It will retrieve an auth token from EC2 metadata service during startup time, and use it to
// create a connection to RDS.
// The auth token is valid for "x" minutes, but we do not need to refresh it as long as the
// connection is alive.
type Auth struct{}

func NewAuth() *Auth {
	return &Auth{}
}

func (r *Auth) AuthToken(hostname string, port uint16, user string) (string, error) {
	instance, err := parseRDSHostname(hostname)
	if err != nil {
		return "", errors.Wrap(err, "Error parsing RDS hostname")
	}

	sess, err := session.NewSession()
	if err != nil {
		return "", errors.Wrap(err, "Error creating AWS session for RDS IAM auth")
	}
	creds := sess.Config.Credentials
	if creds == nil {
		return "", errors.New("No AWS credentials found from current session")
	}

	authToken, err := rdsutils.BuildAuthToken(
		fmt.Sprintf("%s:%d", instance.hostname, port),
		instance.region, user, creds)
	if err != nil {
		return "", errors.Wrap(err, "Error building auth token for RDS IAM auth")
	}

	return authToken, nil
}

type rdsInstance struct {
	region   string
	hostname string
}

// parseRDSHostname parses the RDS hostname and returns the region and instance name.
// It is in the form of <instance-name>.<account-id>.<region>.rds.amazonaws.com
// e.g., postgresmydb.123456789012.us-east-1.rds.amazonaws.com
func parseRDSHostname(name string) (*rdsInstance, error) {
	if !strings.HasSuffix(name, ".rds.amazonaws.com") {
		return nil, errors.Newf("not an RDS hostname, expecting '.rds.amazonaws.com' suffix, %q", name)
	}

	parts := strings.Split(name, ".")
	if len(parts) != 6 {
		return nil, errors.Newf("unexpected RDS hostname format, %q", name)
	}

	if parts[0] == "" {
		return nil, errors.Newf("unexpected instance name in RDS hostname format, %q", name)
	}

	if parts[1] == "" {
		return nil, errors.Newf("unexpected account ID in RDS hostname format, %q", name)
	}

	if parts[2] == "" {
		return nil, errors.Newf("unexpected region in RDS hostname format, %q", name)
	}

	return &rdsInstance{
		region:   parts[2],
		hostname: name,
	}, nil
}
