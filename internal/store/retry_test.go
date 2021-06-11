package store

import (
	"os"
	"testing"

	"github.com/cockroachdb/errors"
)

func TestGetZipFileWithRetry(t *testing.T) {
	tests := []struct {
		name     string
		errs     []error
		succeeds bool
	}{
		{
			name:     "success first try",
			errs:     []error{nil},
			succeeds: true,
		},
		{
			name:     "success second try",
			errs:     []error{errors.New("not a valid zip file"), nil},
			succeeds: true,
		},
		{
			name:     "error that doesn't get a retry",
			errs:     []error{errors.New("blah")},
			succeeds: false,
		},
		{
			name:     "fails twice",
			errs:     []error{errors.New("not a valid zip file"), errors.New("not a valid zip file")},
			succeeds: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var tmp *os.File
			defer func() {
				if tmp == nil {
					return
				}

				tmp.Close()
			}()

			tries := 0
			get := func() (string, *ZipFile, error) {
				var err error
				tmp, err = os.CreateTemp("", "")
				if err != nil {
					t.Fatalf("TempFile(%v)", err)
				}

				err = test.errs[tries]
				var zf *ZipFile
				if err == nil {
					zf = &ZipFile{}
				}
				tries++
				return tmp.Name(), zf, err
			}

			_, zf, err := GetZipFileWithRetry(get)
			if test.succeeds && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if test.succeeds && zf == nil {
				t.Error("expected a zip file; got nil")
			}
		})
	}
}
