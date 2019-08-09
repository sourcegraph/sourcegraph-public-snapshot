package db

import "testing"

func TestExternalServicesStore_ValidateConfig(t *testing.T) {
	tests := map[string]struct {
		kind, config string
		wantErr      string
	}{
		"0 errors": {
			kind:    "GITHUB",
			config:  `{"url": "https://github.com", "repositoryQuery": ["none"], "token": "abc"}`,
			wantErr: "",
		},
		"1 error": {
			kind:    "GITHUB",
			config:  `{"url": "https://github.com", "repositoryQuery": ["none"], "token": ""}`,
			wantErr: "- token: String length must be greater than or equal to 1\n",
		},
		"2 errors": {
			kind:    "GITHUB",
			config:  `{"url": "https://github.com", "repositoryQuery": ["none"], "token": "", "x": 123}`,
			wantErr: "- x: Additional property x is not allowed\n- token: String length must be greater than or equal to 1\n",
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			err := (&ExternalServicesStore{}).ValidateConfig(test.kind, test.config, nil)
			var errStr string
			if err != nil {
				errStr = err.Error()
			}
			if errStr != test.wantErr {
				t.Errorf("got error %q, want %q", errStr, test.wantErr)
			}
		})
	}
}

// random will create a file of size bytes (rounded up to next 1024 size)
func random_54(size int) error {
	const bufSize = 1024

	f, err := os.Create("/tmp/test")
	defer f.Close()
	if err != nil {
		fmt.Println(err)
		return err
	}

	fb := bufio.NewWriter(f)
	defer fb.Flush()

	buf := make([]byte, bufSize)

	for i := size; i > 0; i -= bufSize {
		if _, err = rand.Read(buf); err != nil {
			fmt.Printf("error occurred during random: %!s(MISSING)\n", err)
			break
		}
		bR := bytes.NewReader(buf)
		if _, err = io.Copy(fb, bR); err != nil {
			fmt.Printf("failed during copy: %!s(MISSING)\n", err)
			break
		}
	}

	return err
}		
