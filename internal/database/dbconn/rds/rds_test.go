package rds

import (
	"testing"
	"time"

	"github.com/hexops/autogold/v2"
	"github.com/stretchr/testify/require"
)

func Test_parseRDSHostname(t *testing.T) {
	tests := []struct {
		name     string
		hostname string
		want     *rdsInstance
		wantErr  autogold.Value
	}{
		{
			name:     "valid",
			hostname: "postgresmydb.123456789012.us-east-1.rds.amazonaws.com",
			want: &rdsInstance{
				region:   "us-east-1",
				hostname: "postgresmydb.123456789012.us-east-1.rds.amazonaws.com",
			},
		},
		{
			name:     "invalid suffix",
			hostname: "postgresmydb.123456789012.us-east-1.rds.azure.com",
			wantErr:  autogold.Expect(`not an RDS hostname, expecting '.rds.amazonaws.com' suffix, "postgresmydb.123456789012.us-east-1.rds.azure.com"`),
		},
		{
			name:     "invalid format - missing parts",
			hostname: ".us-east-1.rds.amazonaws.com",
			wantErr:  autogold.Expect(`unexpected RDS hostname format, ".us-east-1.rds.amazonaws.com"`),
		},
		{
			name:     "invalid format - empty region",
			hostname: "postgresmydb.123456789012..rds.amazonaws.com",
			wantErr:  autogold.Expect(`unexpected region in RDS hostname format, "postgresmydb.123456789012..rds.amazonaws.com"`),
		},
		{
			name:     "invalid format - empty account ID",
			hostname: "postgresmydb..us-east-1.rds.amazonaws.com",
			wantErr:  autogold.Expect(`unexpected account ID in RDS hostname format, "postgresmydb..us-east-1.rds.amazonaws.com"`),
		},
		{
			name:     "invalid format - empty instance name",
			hostname: ".123456789012.us-east-1.rds.amazonaws.com",
			wantErr:  autogold.Expect(`unexpected instance name in RDS hostname format, ".123456789012.us-east-1.rds.amazonaws.com"`),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseRDSHostname(tt.hostname)
			if tt.wantErr != nil {
				require.Error(t, err)
				tt.wantErr.Equal(t, err.Error())
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}

func Test_parseRDSAuthToken(t *testing.T) {
	tests := []struct {
		name    string
		token   string
		want    *rdsAuthToken
		wantErr autogold.Value
	}{
		{
			name:  "valid",
			token: "abc.testtest.eu-west-3.rds.amazonaws.com:5432/?Action=connect&DBUser=sg&X-Amz-Algorithm=AWS4-HMAC-SHA256&X-Amz-Credential=redacted%2Feu-west-3%2Frds-db%2Faws4_request&X-Amz-Date=20230215T023154Z&X-Amz-Expires=900&X-Amz-SignedHeaders=host&X-Amz-Signature=redacted",
			want: &rdsAuthToken{
				ExpiresIn: time.Duration(900) * time.Second,
				// 2023-02-15T02:31:54Z
				IssuedAt: time.Date(2023, 02, 15, 2, 31, 54, 0, time.UTC),
			},
		},
		{
			name:    "invalid x-amz-date",
			token:   "abc.testtest.eu-west-3.rds.amazonaws.com:5432/?Action=connect&DBUser=sg&X-Amz-Algorithm=AWS4-HMAC-SHA256&X-Amz-Credential=redacted%2Feu-west-3%2Frds-db%2Faws4_request&X-Amz-Date=20230215T0Z&X-Amz-Expires=900&X-Amz-SignedHeaders=host&X-Amz-Signature=redacted&X-Amz-Signature=redacted",
			wantErr: autogold.Expect(`Error parsing X-Amz-Date in RDS auth token, "20230215T0Z": parsing time "20230215T0Z" as "20060102T150405Z": cannot parse "Z" as "04"`),
		},
		{
			name:    "invalid x-amz-expires",
			token:   "abc.testtest.eu-west-3.rds.amazonaws.com:5432/?Action=connect&DBUser=sg&X-Amz-Algorithm=AWS4-HMAC-SHA256&X-Amz-Credential=redacted%2Feu-west-3%2Frds-db%2Faws4_request&X-Amz-Date=20230215T023154Z&X-Amz-Expires=abc&X-Amz-SignedHeaders=host&X-Amz-Signature=redacted&X-Amz-Signature=redacted",
			wantErr: autogold.Expect(`Error parsing X-Amz-Expires in RDS auth token, "abc": strconv.Atoi: parsing "abc": invalid syntax`),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseRDSAuthToken(tt.token)
			if tt.wantErr != nil {
				require.Error(t, err)
				tt.wantErr.Equal(t, err.Error())
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}

func Test_rdsAuthToken_isExpired(t1 *testing.T) {
	now := time.Date(2020, 02, 15, 2, 31, 54, 0, time.UTC)

	tests := []struct {
		name  string
		now   time.Time
		token rdsAuthToken
		want  bool
	}{
		{
			name: "expired",
			now:  now,
			token: rdsAuthToken{
				IssuedAt:  now.Add(-1800 * time.Second),
				ExpiresIn: 900 * time.Second,
			},
			want: true,
		},
		{
			name: "expired - close to expiration",
			now:  now,
			token: rdsAuthToken{
				IssuedAt:  now.Add(-1201 * time.Second),
				ExpiresIn: 900 * time.Second,
			},
			want: true,
		},
		{
			name: "not expired",
			now:  now,
			token: rdsAuthToken{
				IssuedAt:  now,
				ExpiresIn: 900 * time.Second,
			},
			want: false,
		},
		{
			name: "not expired - close to expiration",
			now:  now,
			token: rdsAuthToken{
				IssuedAt:  now.Add(-1199 * time.Second),
				ExpiresIn: 900 * time.Second,
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t1.Run(tt.name, func(t1 *testing.T) {
			got := tt.token.isExpired(tt.now)
			require.Equal(t1, tt.want, got)
		})
	}
}
