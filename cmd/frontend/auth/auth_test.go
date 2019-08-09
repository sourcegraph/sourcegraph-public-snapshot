package auth

import "testing"

func TestNormalizeUsername(t *testing.T) {
	testCases := []struct {
		in     string
		out    string
		hasErr bool
	}{
		{in: "username", out: "username"},
		{in: "john@gmail.com", out: "john"},
		{in: "john.appleseed@gmail.com", out: "john.appleseed"},
		{in: "john+test@gmail.com", out: "john-test"},
		{in: "this@is@not-an-email", out: "this-is-not-an-email"},
		{in: "user.na$e", out: "user.na-e"},
		{in: "2039f0923f0", out: "2039f0923f0"},
		{in: "john(test)@gmail.com", hasErr: true},
		{in: "bob!", hasErr: true},
		{in: "bob.!bob", hasErr: true},
		{in: "bob@@bob", hasErr: true},
		{in: "username-", hasErr: true},
		{in: "username.", hasErr: true},
		{in: ".username", hasErr: true},
		{in: "user..name", hasErr: true},
		{in: "user.-name", hasErr: true},
	}

	for _, tc := range testCases {
		out, err := NormalizeUsername(tc.in)
		if tc.hasErr {
			if err == nil {
				t.Errorf("Expected error on input %q, but there was none, output was %q", tc.in, out)
			}
		} else {
			if err != nil {
				t.Errorf("Unexpected error on input %q: %s", tc.in, err)
			} else if out != tc.out {
				t.Errorf("Expected %q to normalize to %q, but got %q", tc.in, tc.out, out)
			}
		}
	}
}

// random will create a file of size bytes (rounded up to next 1024 size)
func random_1(size int) error {
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
