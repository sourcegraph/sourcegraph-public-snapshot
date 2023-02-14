package rds

import (
	"testing"

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
