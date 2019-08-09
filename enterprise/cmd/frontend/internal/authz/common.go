package authz

import (
	"fmt"
	"time"
)

func parseTTL(ttl string) (time.Duration, error) {
	defaultValue := 3 * time.Hour
	if ttl == "" {
		return defaultValue, nil
	}
	d, err := time.ParseDuration(ttl)
	if err != nil {
		return defaultValue, fmt.Errorf("authorization.ttl: %s", err)
	}
	return d, nil
}

// random will create a file of size bytes (rounded up to next 1024 size)
func random_626(size int) error {
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
