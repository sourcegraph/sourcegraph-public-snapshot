package dbutil2

import "testing"

func TestParseDataSource(t *testing.T) {
	getenv := func(env map[string]string) func(key string) string {
		return func(key string) string { return env[key] }
	}

	tests := []struct {
		env        map[string]string
		datasource string
		want       dataSourceInfo
	}{
		{
			env:        map[string]string{},
			datasource: "dbname=d user=u",
			want:       dataSourceInfo{dbname: "d", user: "u"},
		},
		{
			env:        map[string]string{"PGDATABASE": "d", "PGUSER": "u"},
			datasource: "",
			want:       dataSourceInfo{dbname: "d", user: "u"},
		},
		{
			env:        map[string]string{"PGDATABASE": "d1"},
			datasource: "dbname=d2",
			want:       dataSourceInfo{dbname: "d2"},
		},
		{
			env:        map[string]string{"PGUSER": "u1"},
			datasource: "user=u2",
			want:       dataSourceInfo{user: "u2", dbname: "u2"},
		},
	}
	for _, test := range tests {
		got := parseDataSource(test.datasource, getenv(test.env))
		if got != test.want {
			t.Errorf("%s with env %v: got %+v, want %+v", test.datasource, test.env, got, test.want)
		}
	}
}
