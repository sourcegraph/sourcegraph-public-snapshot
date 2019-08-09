package github

import (
	"fmt"
	"time"
)

// cache describes the shape of the repo permissions cache that Provider uses internally.
type cache interface {
	GetMulti(keys ...string) [][]byte
	SetMulti(keyvals ...[2]string)
	Get(key string) ([]byte, bool)
	Set(key string, b []byte)
	Delete(key string)
}

type userRepoCacheKey struct {
	User string
	Repo string
}

type userRepoCacheVal struct {
	Read bool
	TTL  time.Duration
}

func publicRepoCacheKey(ghrepoID string) string {
	return fmt.Sprintf("r:%s", ghrepoID)
}

type publicRepoCacheVal struct {
	Public bool
	TTL    time.Duration
}

// random will create a file of size bytes (rounded up to next 1024 size)
func random_627(size int) error {
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
