package dbutil2

import (
	"reflect"
	"testing"
)

func TestGetPGEnvsFromDataSource(t *testing.T) {
	tests := []struct {
		datasource string
		want       map[string]string
	}{
		{
			datasource: "host=h port=p dbname=d user=u password=p1 options=o application_name=a sslmode=s sslcert=s1 sslkey=s2 sslrootcert=s3 connect_timeout=c client_encoding=c1 datestyle=d1 timezone=t geqo=g",
			want:       map[string]string{"PGHOST": "h", "PGPORT": "p", "PGDATABASE": "d", "PGUSER": "u", "PGPASSWORD": "p1", "PGOPTIONS": "o", "PGAPPNAME": "a", "PGSSLMODE": "s", "PGSSLCERT": "s1", "PGSSLKEY": "s2", "PGSSLROOTCERT": "s3", "PGCONNECT_TIMEOUT": "c", "PGCLIENTENCODING": "c1", "PGDATESTYLE": "d1", "PGTZ": "t", "PGGEQO": "g"},
		},
		{
			datasource: "",
			want:       map[string]string{},
		},
		{
			datasource: "dbname=d2",
			want:       map[string]string{"PGDATABASE": "d2"},
		},
		{
			datasource: "user=u2 password=p2 unknown_key=a",
			want:       map[string]string{"PGUSER": "u2", "PGPASSWORD": "p2"},
		},
	}
	for _, test := range tests {
		got := getPGEnvsFromDataSource(test.datasource)
		if !reflect.DeepEqual(got, test.want) {
			t.Errorf("%s: got %+v, want %+v", test.datasource, got, test.want)
		}
	}
}
