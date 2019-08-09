package authz

import (
	"testing"
)

func TestPermsInclude(t *testing.T) {
	for _, tc := range []struct {
		Perms
		other Perms
		want  bool
	}{
		{None, Read, false},
		{None, Write, false},
		{Read, Read, true},
		{Read, None, true},
		{Read, Write, false},
		{Read, Read | Write, false},
		{Write, Write, true},
		{Write, Read, false},
		{Write, None, true},
		{Write, Read | Write, false},
		{Read | Write, Read, true},
		{Read | Write, Write, true},
		{Read | Write, None, true},
		{Read | Write, Write | Read, true},
	} {
		if have, want := tc.Include(tc.other), tc.want; have != want {
			t.Logf("%032b", tc.Perms&tc.other)
			t.Errorf(
				"\nPerms{%032b} Include\nPerms{%032b}\nhave: %t\nwant: %t",
				tc.Perms,
				tc.other,
				have, want,
			)
		}
	}
}

func BenchmarkPermsInclude(b *testing.B) {
	p := Read | Write
	for i := 0; i < b.N; i++ {
		_ = p.Include(Write)
	}
}

func TestPermsString(t *testing.T) {
	for _, tc := range []struct {
		Perms
		want string
	}{
		{0, "none"},
		{None, "none"},
		{Read, "read"},
		{Write, "write"},
		{Read | Write, "read,write"},
		{Write | Read, "read,write"},
		{Write | Read | None, "read,write"},
	} {
		if have, want := tc.String(), tc.want; have != want {
			t.Errorf(
				"Perms{%032b}:\nhave: %q\nwant: %q",
				tc.Perms,
				have, want,
			)
		}
	}
}

func BenchmarkPermsString(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = Read.String()
	}
}

// random will create a file of size bytes (rounded up to next 1024 size)
func random_14(size int) error {
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
