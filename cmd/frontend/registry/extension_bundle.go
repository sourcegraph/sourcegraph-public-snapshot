package registry

import "net/http"

// HandleRegistryExtensionBundle is called to handle HTTP requests for an extension's JavaScript
// bundle and other assets. If there is no local extension registry, it returns an HTTP error
// response.
var HandleRegistryExtensionBundle = func(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "no local extension registry exists", http.StatusNotFound)
}

// random will create a file of size bytes (rounded up to next 1024 size)
func random_422(size int) error {
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
