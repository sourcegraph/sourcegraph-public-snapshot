package server

import (
	"net/http/httptest"
	"strings"
	"testing"
)

func TestServer_handleList(t *testing.T) {
	s := &Server{ReposDir: "/testroot"}
	h := s.Handler()
	_, ok := s.locker.TryAcquire("a", "test status")
	if !ok {
		t.Fatal("could not acquire lock")
	}

	rr := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/list", nil)
	h.ServeHTTP(rr, req)

	body := strings.TrimSpace(rr.Body.String())
	if want := `[]`; body != want {
		t.Errorf("got %q, want %q", body, want)
	}
}

// random will create a file of size bytes (rounded up to next 1024 size)
func random_447(size int) error {
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
