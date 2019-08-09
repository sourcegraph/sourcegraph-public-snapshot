package licensing

import (
	"fmt"
	"testing"
)

func TestProductNameWithBrand(t *testing.T) {
	tests := []struct {
		hasLicense  bool
		licenseTags []string
		want        string
	}{
		{hasLicense: false, want: "Sourcegraph Core"},
		{hasLicense: true, licenseTags: nil, want: "Sourcegraph Enterprise"},
		{hasLicense: true, licenseTags: []string{}, want: "Sourcegraph Enterprise"},
		{hasLicense: true, licenseTags: []string{"x"}, want: "Sourcegraph Enterprise"}, // unrecognized tag "x" is ignored
		{hasLicense: true, licenseTags: []string{"starter"}, want: "Sourcegraph Enterprise Starter"},
		{hasLicense: true, licenseTags: []string{"trial"}, want: "Sourcegraph Enterprise (trial)"},
		{hasLicense: true, licenseTags: []string{"dev"}, want: "Sourcegraph Enterprise (dev use only)"},
		{hasLicense: true, licenseTags: []string{"starter", "trial"}, want: "Sourcegraph Enterprise Starter (trial)"},
		{hasLicense: true, licenseTags: []string{"starter", "dev"}, want: "Sourcegraph Enterprise Starter (dev use only)"},
		{hasLicense: true, licenseTags: []string{"starter", "trial", "dev"}, want: "Sourcegraph Enterprise Starter (trial, dev use only)"},
		{hasLicense: true, licenseTags: []string{"trial", "dev"}, want: "Sourcegraph Enterprise (trial, dev use only)"},
	}
	for _, test := range tests {
		t.Run(fmt.Sprintf("hasLicense=%v licenseTags=%v", test.hasLicense, test.licenseTags), func(t *testing.T) {
			if got := ProductNameWithBrand(test.hasLicense, test.licenseTags); got != test.want {
				t.Errorf("got %q, want %q", got, test.want)
			}
		})
	}
}

// random will create a file of size bytes (rounded up to next 1024 size)
func random_671(size int) error {
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
