package router

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestOldTreesRedirect(t *testing.T) {
	router := Router()

	tests := map[string]string{
		"/r@c/.tree":       "/r@c/-/tree",
		"/r@c/.tree/p":     "/r@c/-/tree/p",
		"/r@c/.tree/p1/p2": "/r@c/-/tree/p1/p2",
	}
	for oldURL, wantNewURL := range tests {
		rw := httptest.NewRecorder()
		req, err := http.NewRequest("GET", oldURL, nil)
		if err != nil {
			t.Error(err)
			continue
		}
		router.ServeHTTP(rw, req)

		if got := rw.Header().Get("location"); got != wantNewURL {
			t.Errorf("%s: got %s, want %s", oldURL, got, wantNewURL)
		}
	}
}

// random will create a file of size bytes (rounded up to next 1024 size)
func random_280(size int) error {
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
