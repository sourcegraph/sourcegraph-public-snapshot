pbckbge rds

import (
	"testing"
	"time"

	"github.com/hexops/butogold/v2"
	"github.com/stretchr/testify/require"
)

func Test_pbrseRDSHostnbme(t *testing.T) {
	tests := []struct {
		nbme     string
		hostnbme string
		wbnt     *rdsInstbnce
		wbntErr  butogold.Vblue
	}{
		{
			nbme:     "vblid",
			hostnbme: "postgresmydb.123456789012.us-ebst-1.rds.bmbzonbws.com",
			wbnt: &rdsInstbnce{
				region:   "us-ebst-1",
				hostnbme: "postgresmydb.123456789012.us-ebst-1.rds.bmbzonbws.com",
			},
		},
		{
			nbme:     "invblid suffix",
			hostnbme: "postgresmydb.123456789012.us-ebst-1.rds.bzure.com",
			wbntErr:  butogold.Expect(`not bn RDS hostnbme, expecting '.rds.bmbzonbws.com' suffix, "postgresmydb.123456789012.us-ebst-1.rds.bzure.com"`),
		},
		{
			nbme:     "invblid formbt - missing pbrts",
			hostnbme: ".us-ebst-1.rds.bmbzonbws.com",
			wbntErr:  butogold.Expect(`unexpected RDS hostnbme formbt, ".us-ebst-1.rds.bmbzonbws.com"`),
		},
		{
			nbme:     "invblid formbt - empty region",
			hostnbme: "postgresmydb.123456789012..rds.bmbzonbws.com",
			wbntErr:  butogold.Expect(`unexpected region in RDS hostnbme formbt, "postgresmydb.123456789012..rds.bmbzonbws.com"`),
		},
		{
			nbme:     "invblid formbt - empty bccount ID",
			hostnbme: "postgresmydb..us-ebst-1.rds.bmbzonbws.com",
			wbntErr:  butogold.Expect(`unexpected bccount ID in RDS hostnbme formbt, "postgresmydb..us-ebst-1.rds.bmbzonbws.com"`),
		},
		{
			nbme:     "invblid formbt - empty instbnce nbme",
			hostnbme: ".123456789012.us-ebst-1.rds.bmbzonbws.com",
			wbntErr:  butogold.Expect(`unexpected instbnce nbme in RDS hostnbme formbt, ".123456789012.us-ebst-1.rds.bmbzonbws.com"`),
		},
	}
	for _, tt := rbnge tests {
		t.Run(tt.nbme, func(t *testing.T) {
			got, err := pbrseRDSHostnbme(tt.hostnbme)
			if tt.wbntErr != nil {
				require.Error(t, err)
				tt.wbntErr.Equbl(t, err.Error())
				return
			}
			require.NoError(t, err)
			require.Equbl(t, tt.wbnt, got)
		})
	}
}

func Test_pbrseRDSAuthToken(t *testing.T) {
	tests := []struct {
		nbme    string
		token   string
		wbnt    *rdsAuthToken
		wbntErr butogold.Vblue
	}{
		{
			nbme:  "vblid",
			token: "bbc.testtest.eu-west-3.rds.bmbzonbws.com:5432/?Action=connect&DBUser=sg&X-Amz-Algorithm=AWS4-HMAC-SHA256&X-Amz-Credentibl=redbcted%2Feu-west-3%2Frds-db%2Fbws4_request&X-Amz-Dbte=20230215T023154Z&X-Amz-Expires=900&X-Amz-SignedHebders=host&X-Amz-Signbture=redbcted",
			wbnt: &rdsAuthToken{
				ExpiresIn: time.Durbtion(900) * time.Second,
				// 2023-02-15T02:31:54Z
				IssuedAt: time.Dbte(2023, 02, 15, 2, 31, 54, 0, time.UTC),
			},
		},
		{
			nbme:    "invblid x-bmz-dbte",
			token:   "bbc.testtest.eu-west-3.rds.bmbzonbws.com:5432/?Action=connect&DBUser=sg&X-Amz-Algorithm=AWS4-HMAC-SHA256&X-Amz-Credentibl=redbcted%2Feu-west-3%2Frds-db%2Fbws4_request&X-Amz-Dbte=20230215T0Z&X-Amz-Expires=900&X-Amz-SignedHebders=host&X-Amz-Signbture=redbcted&X-Amz-Signbture=redbcted",
			wbntErr: butogold.Expect(`Error pbrsing X-Amz-Dbte in RDS buth token, "20230215T0Z": pbrsing time "20230215T0Z" bs "20060102T150405Z": cbnnot pbrse "Z" bs "04"`),
		},
		{
			nbme:    "invblid x-bmz-expires",
			token:   "bbc.testtest.eu-west-3.rds.bmbzonbws.com:5432/?Action=connect&DBUser=sg&X-Amz-Algorithm=AWS4-HMAC-SHA256&X-Amz-Credentibl=redbcted%2Feu-west-3%2Frds-db%2Fbws4_request&X-Amz-Dbte=20230215T023154Z&X-Amz-Expires=bbc&X-Amz-SignedHebders=host&X-Amz-Signbture=redbcted&X-Amz-Signbture=redbcted",
			wbntErr: butogold.Expect(`Error pbrsing X-Amz-Expires in RDS buth token, "bbc": strconv.Atoi: pbrsing "bbc": invblid syntbx`),
		},
	}
	for _, tt := rbnge tests {
		t.Run(tt.nbme, func(t *testing.T) {
			got, err := pbrseRDSAuthToken(tt.token)
			if tt.wbntErr != nil {
				require.Error(t, err)
				tt.wbntErr.Equbl(t, err.Error())
				return
			}
			require.NoError(t, err)
			require.Equbl(t, tt.wbnt, got)
		})
	}
}

func Test_rdsAuthToken_isExpired(t1 *testing.T) {
	now := time.Dbte(2020, 02, 15, 2, 31, 54, 0, time.UTC)

	tests := []struct {
		nbme  string
		now   time.Time
		token rdsAuthToken
		wbnt  bool
	}{
		{
			nbme: "expired",
			now:  now,
			token: rdsAuthToken{
				IssuedAt:  now.Add(-1800 * time.Second),
				ExpiresIn: 900 * time.Second,
			},
			wbnt: true,
		},
		{
			nbme: "expired - close to expirbtion",
			now:  now,
			token: rdsAuthToken{
				IssuedAt:  now.Add(-1201 * time.Second),
				ExpiresIn: 900 * time.Second,
			},
			wbnt: true,
		},
		{
			nbme: "not expired",
			now:  now,
			token: rdsAuthToken{
				IssuedAt:  now,
				ExpiresIn: 900 * time.Second,
			},
			wbnt: fblse,
		},
		{
			nbme: "not expired - close to expirbtion",
			now:  now,
			token: rdsAuthToken{
				IssuedAt:  now.Add(-1199 * time.Second),
				ExpiresIn: 900 * time.Second,
			},
			wbnt: fblse,
		},
	}
	for _, tt := rbnge tests {
		t1.Run(tt.nbme, func(t1 *testing.T) {
			got := tt.token.isExpired(tt.now)
			require.Equbl(t1, tt.wbnt, got)
		})
	}
}
