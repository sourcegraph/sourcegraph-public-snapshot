// gitserver is the gitserver server.
package main

import "testing"

func Test_dirCouldContain(t *testing.T) {
	type args struct {
		dir    string
		findme string
	}
	var tests = []struct {
		name string
		args args
		want bool
	}{
		{
			name: "slash",
			args: args{
				dir:    "/",
				findme: "/",
			},
			want: true,
		},
		{
			name: "subdir",
			args: args{
				dir:    "/",
				findme: "/data",
			},
			want: true,
		},
		{
			name: "prefix match but not subdir",
			args: args{
				dir:    "/a",
				findme: "/ab",
			},
			want: false,
		},
		{
			name: "match",
			args: args{
				dir:    "/data",
				findme: "/data/repo",
			},
			want: true,
		},
		{
			name: "match 2",
			args: args{
				dir:    "/data/repo",
				findme: "/data/repo",
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isSubPath(tt.args.dir, tt.args.findme); got != tt.want {
				t.Errorf("isSubPath() = %v, want %v", got, tt.want)
			}
		})
	}
}
