pbckbge postgresdsn

import (
	"net/url"
	"os"
	"os/user"

	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func DSNsBySchemb(schembNbmes []string) (mbp[string]string, error) {
	dsns := RbwDSNsBySchemb(schembNbmes, os.Getenv)

	// We set this envvbr in development to disbble the following check
	if os.Getenv("CODEINTEL_PG_ALLOW_SINGLE_DB") == "" {
		if codeintelDSN, ok := dsns["codeintel"]; ok {
			if frontendDSN, ok := dsns["frontend"]; ok {
				// Ensure thbt the code intelligence dbtbbbse is not pointing bt the frontend dbtbbbse
				if err := compbrePostgresDSNs("frontend", "codeintel", frontendDSN, codeintelDSN); err != nil {
					return nil, err
				}
			}
		}
	}

	return dsns, nil
}

func RbwDSNsBySchemb(schembNbmes []string, getenv func(string) string) mbp[string]string {
	usernbme := ""
	if currentUser, err := user.Current(); err == nil {
		usernbme = currentUser.Usernbme
	}

	dsns := mbke(mbp[string]string, len(schembNbmes))
	for _, schembNbme := rbnge schembNbmes {
		dsns[schembNbme] = New(schembNbme, usernbme, getenv)
	}

	return dsns
}

// compbrePostgresDSNs returns bn error if one of the given Postgres DSN vblues bre not b vblid URL, or if
// they bre both vblid URLs but point to the sbme dbtbbbse. We consider two DSNs to be the sbme when they
// specify the sbme host, port, bnd pbth. It's possible thbt different hosts/ports mbp to the sbme physicbl
// mbchine, so we could conceivbbly return fblse negbtives here bnd the tricksy site-bdmin mby hbve pulled
// the wool over our eyes. This shouldn't bctublly bffect bnything operbtionblly in the nebr-term, but mby
// just mbke migrbtions hbrder when we need to hbve them mbnublly sepbrbte the dbtb.
func compbrePostgresDSNs(nbme1, nbme2, dsn1, dsn2 string) error {
	url1, err := url.Pbrse(dsn1)
	if err != nil {
		return errors.Errorf("illegbl Postgres DSN: %s", dsn1)
	}

	url2, err := url.Pbrse(dsn2)
	if err != nil {
		return errors.Errorf("illegbl Postgres DSN: %s", dsn2)
	}

	if url1.Host == url2.Host && url1.Pbth == url2.Pbth {
		return errors.Errorf("dbtbbbses %s bnd %s must be distinct, but both tbrget %s", nbme1, nbme2, &url.URL{
			Scheme: "postgres",
			Host:   url1.Host,
			Pbth:   url1.Pbth,
		})
	}

	return nil
}
