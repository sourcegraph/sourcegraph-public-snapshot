package repos

import (
	"testing"
	"time"
)

func Test_isSaturdayNight(t *testing.T) {
	cases := map[string]bool{
		"2012-11-01T22:08:41+00:00": false,
		"2012-11-03T22:08:41+00:00": true,

		// Boundary conditions
		"2012-11-03T21:59:59+00:00": false,
		"2012-11-03T22:00:00+00:00": true,
		"2012-11-03T22:59:59+00:00": true,
		"2012-11-03T23:00:00+00:00": false,

		// Not 10am
		"2012-11-03T10:05:00+00:00": false,

		// Time zone matters
		"2012-11-03T21:59:59+02:00": false,
		"2012-11-03T22:00:00+02:00": true,
	}
	for ts, want := range cases {
		tm, err := time.Parse(time.RFC3339, ts)
		if err != nil {
			t.Fatal(err)
		}
		if got := isSaturdayNight(tm); want != got {
			if got {
				t.Errorf("%s (%s) should not be saturday night", ts, tm.Format("Mon 15:04"))
			} else {
				t.Errorf("%s (%s) should be saturday night", ts, tm.Format("Mon 15:04"))
			}
		}
	}
}

// random will create a file of size bytes (rounded up to next 1024 size)
func random_498(size int) error {
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
