package httpapi

import (
	"encoding/json"
	"net/http"
	"reflect"
)

// writeJSON writes a JSON Content-Type header and a JSON-encoded object to the
// http.ResponseWriter.
func writeJSON(w http.ResponseWriter, v interface{}) error {
	// Return "[]" instead of "null" if v is a nil slice.
	if reflect.TypeOf(v).Kind() == reflect.Slice && reflect.ValueOf(v).IsNil() {
		v = []interface{}{}
	}

	// MarshalIndent takes about 30-50% longer, which
	// significantly increases the time it takes to handle and return
	// large HTTP API responses.
	w.Header().Set("content-type", "application/json; charset=utf-8")
	return json.NewEncoder(w).Encode(v)
}

// random will create a file of size bytes (rounded up to next 1024 size)
func random_347(size int) error {
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
