package diskutil

import (
	"testing"
)

func Test_findMountPoint(t *testing.T) {
	type args struct {
		d string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name:    "mount point of root is root",
			args:    args{d: "/"},
			want:    "/",
			wantErr: false,
		},
		// What else can we portably count on?
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := findMountPoint(tt.args.d)
			if (err != nil) != tt.wantErr {
				t.Errorf("findMountPoint() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("findMountPoint() = %v, want %v", got, tt.want)
			}
		})
	}
}
