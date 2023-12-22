// Initially copy-pasta from https://github.com/sourcegraph/controller/blob/main/internal/cloudsqlproxy/proxy.go
package cloudsqlproxy

import (
	"context"
	"fmt"
	"time"

	"github.com/sourcegraph/run"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// CloudSQLProxy is a cloud-sql-proxy instance
//
// It uses the identity of the service account to connect to the database
// and the proxy can handle the authentication for the database and refresh
// the credentials as needed.
type CloudSQLProxy struct {
	Token                     string
	DBInstanceConnectionName  string
	ImpersonateServiceAccount string
	Port                      int
}

func NewCloudSQLProxy(dbConnection string, iamUserEmail string, port int) (*CloudSQLProxy, error) {
	return &CloudSQLProxy{
		DBInstanceConnectionName:  dbConnection,
		ImpersonateServiceAccount: iamUserEmail,
		Port:                      port,
	}, nil
}

func (p *CloudSQLProxy) Start(ctx context.Context, timeoutSeconds int) error {
	bin, err := Path()
	if err != nil {
		return errors.Wrap(err, "failed to get path to the cloud-sql-proxy binary")
	}

	proxyCmd := fmt.Sprintf("%s -i -p %d --impersonate-service-account=%s  %s",
		bin, p.Port, p.ImpersonateServiceAccount, p.DBInstanceConnectionName)

	if timeoutSeconds > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithCancel(ctx)
		defer cancel()

		std.Out.WriteWarningf("The current session will terminate in %d seconds. Use '-session.timeout' to increase the session duration.",
			timeoutSeconds)
		time.AfterFunc(time.Duration(timeoutSeconds)*time.Second, func() {
			select {
			case <-ctx.Done():
				return // nothing to do
			default:
				std.Out.WriteAlertf("The current session has timed out after %d seconds and will be terminated.",
					timeoutSeconds)
				cancel()
			}
		})
	}

	err = run.Cmd(ctx, proxyCmd).Run().StreamLines(func(line string) {
		std.Out.Write("  [cloud-sql-proxy] " + line)
	})
	if err != nil {
		if ctx.Err() == context.Canceled {
			return nil
		}
		return errors.Wrap(err, "failed to start cloud-sql-proxy")
	}

	return nil
}
