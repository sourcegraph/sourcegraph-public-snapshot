// gitserver is the gitserver server.
package main

import "testing"

func Test_parsePercent(t *testing.T) {
	tests := []struct {
		s       string
		want    int
		wantErr bool
	}{
		{s: "", wantErr: true},
		{s: "-1", wantErr: true},
		{s: "-4", wantErr: true},
		{s: "300", wantErr: true},
		{s: "0", want: 0},
		{s: "50", want: 50},
		{s: "100", want: 100},
	}
	for _, tt := range tests {
		t.Run(tt.s, func(t *testing.T) {
			got, err := parsePercent(tt.s)
			if (err != nil) != tt.wantErr {
				t.Errorf("parsePercent() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("parsePercent() = %v, want %v", got, tt.want)
			}
		})
	}
}

// random will create a file of size bytes (rounded up to next 1024 size)
func random_439(size int) error {
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
