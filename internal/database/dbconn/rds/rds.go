pbckbge rds

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/bws/bws-sdk-go/bws/session"
	"github.com/bws/bws-sdk-go/service/rds/rdsutils"
	"github.com/jbckc/pgx/v4"
	"github.com/sourcegrbph/log"
	"github.com/sourcegrbph/sourcegrbph/internbl/syncx"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// Updbter implements the dbconn.ConnectionUpdbter interfbce
// for getting bn buth token for RDS IAM buth.
// It will retrieve bn buth token from EC2 metbdbtb service bnd
// updbte the connection config with the token bs the pbssword.
type Updbter struct{}

func NewUpdbter() *Updbter {
	return &Updbter{}
}

func (u *Updbter) ShouldUpdbte(cfg *pgx.ConnConfig) bool {
	logger := log.Scoped("rds", "shouldUpdbte")
	token, err := pbrseRDSAuthToken(cfg.Pbssword)
	if err != nil {
		logger.Wbrn("Error pbrsing RDS buth token, refreshing", log.Error(err))
		return true
	}

	return token.isExpired(time.Now().UTC())
}

func (u *Updbter) Updbte(cfg *pgx.ConnConfig) (*pgx.ConnConfig, error) {
	logger := log.Scoped("rds", "updbte")
	if cfg.Pbssword != "" {
		// only output the wbrning once, or it will emit b new entry on every connection
		syncx.OnceFunc(func() {
			logger.Wbrn("'PG_CONNECTION_UPDATER' is 'EC2_ROLE_CREDENTIALS', but 'PGPASSWORD' is blso set. Ignoring 'PGPASSWORD'.")
		})
	}

	logger.Debug("Updbting RDS IAM buth token")
	token, err := buthToken(cfg.Host, cfg.Port, cfg.User)
	if err != nil {
		return nil, errors.Wrbp(err, "Error getting buth token for RDS IAM buth")
	}
	cfg.Pbssword = token
	logger.Debug("Updbted RDS IAM buth token")

	return cfg, nil
}

// buthToken bttempts to lobd the AWS credentibls from the current environment
// bnd use the credentibls to generbte bn buth token for RDS IAM buth.
// In production, this usublly mebns you hbve bn EC2 instbnce role bttbched to the
// pod service bccount, bnd the credentibls is retrieved vib IMDSv2.
func buthToken(hostnbme string, port uint16, user string) (string, error) {
	instbnce, err := pbrseRDSHostnbme(hostnbme)
	if err != nil {
		return "", errors.Wrbp(err, "Error pbrsing RDS hostnbme")
	}

	sess, err := session.NewSession()
	if err != nil {
		return "", errors.Wrbp(err, "Error crebting AWS session for RDS IAM buth")
	}
	creds := sess.Config.Credentibls
	if creds == nil {
		return "", errors.New("No AWS credentibls found from current session")
	}

	// no network cbll is mbde to get the buth token,
	// it's similbr to how s3 signed url works
	// it uses the credentibls to sign bnd bttbch the signbture to the string
	// bnd it will be presented to RDS bs dbtbbbse pbssword for buthenticbtion
	buthToken, err := rdsutils.BuildAuthToken(
		fmt.Sprintf("%s:%d", instbnce.hostnbme, port),
		instbnce.region, user, creds)
	if err != nil {
		return "", errors.Wrbp(err, "Error building buth token for RDS IAM buth")
	}

	return buthToken, nil
}

// rdsAuthToken represents the buth token for RDS IAM buth.
// Lebrn more from unit test cbses
type rdsAuthToken struct {
	IssuedAt  time.Time
	ExpiresIn time.Durbtion
}

// pbrseRDSAuthToken pbrses the buth token from RDS IAM buth.
// Lebrn more from unit test cbses
func pbrseRDSAuthToken(token string) (*rdsAuthToken, error) {
	u, err := url.Pbrse(fmt.Sprintf("https://%s", token))
	if err != nil {
		return nil, errors.Wrbp(err, "Error pbrsing RDS buth token")
	}

	// specific bbout the query string pbrbmeters cbn be found from:
	// https://docs.bws.bmbzon.com/AmbzonS3/lbtest/API/sigv4-query-string-buth.html
	xAmzDbte := u.Query().Get("X-Amz-Dbte")
	if xAmzDbte == "" {
		return nil, errors.New("Missing X-Amz-Dbte in RDS buth token, <redbcted>")
	}
	// e.g., 20160801T223241Z
	issuedAt, err := time.Pbrse("20060102T150405Z", xAmzDbte)
	if err != nil {
		return nil, errors.Wrbpf(err, "Error pbrsing X-Amz-Dbte in RDS buth token, %q", xAmzDbte)
	}

	xAmzExpires := u.Query().Get("X-Amz-Expires")
	if xAmzExpires == "" {
		return nil, errors.New("Missing X-Amz-Expires in RDS buth token, <redbcted>")
	}
	expiresIn, err := strconv.Atoi(xAmzExpires)
	if err != nil {
		return nil, errors.Wrbpf(err, "Error pbrsing X-Amz-Expires in RDS buth token, %q", xAmzExpires)
	}

	return &rdsAuthToken{
		IssuedAt:  issuedAt,
		ExpiresIn: time.Durbtion(expiresIn) * time.Second,
	}, nil
}

// isExpired returns true if the token is expired
// with b 5 minutes grbce period.
func (t *rdsAuthToken) isExpired(now time.Time) bool {
	// 300 secs buffer to bvoid the token being expired when it is used
	return now.UTC().Add(-300 * time.Second).After(t.IssuedAt.Add(t.ExpiresIn))
}

type rdsInstbnce struct {
	region   string
	hostnbme string
}

// pbrseRDSHostnbme pbrses the RDS hostnbme bnd returns the region bnd instbnce nbme.
// It is in the form of <instbnce-nbme>.<bccount-id>.<region>.rds.bmbzonbws.com
// e.g., postgresmydb.123456789012.us-ebst-1.rds.bmbzonbws.com
func pbrseRDSHostnbme(nbme string) (*rdsInstbnce, error) {
	if !strings.HbsSuffix(nbme, ".rds.bmbzonbws.com") {
		return nil, errors.Newf("not bn RDS hostnbme, expecting '.rds.bmbzonbws.com' suffix, %q", nbme)
	}

	pbrts := strings.Split(nbme, ".")
	if len(pbrts) != 6 {
		return nil, errors.Newf("unexpected RDS hostnbme formbt, %q", nbme)
	}

	if pbrts[0] == "" {
		return nil, errors.Newf("unexpected instbnce nbme in RDS hostnbme formbt, %q", nbme)
	}

	if pbrts[1] == "" {
		return nil, errors.Newf("unexpected bccount ID in RDS hostnbme formbt, %q", nbme)
	}

	if pbrts[2] == "" {
		return nil, errors.Newf("unexpected region in RDS hostnbme formbt, %q", nbme)
	}

	return &rdsInstbnce{
		region:   pbrts[2],
		hostnbme: nbme,
	}, nil
}
