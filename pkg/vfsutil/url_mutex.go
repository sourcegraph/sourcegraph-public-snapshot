package vfsutil

import "sync"

// If we're saving to the local FS, we need to globally synchronize
// writes so we don't corrupt the .zip files with concurrent
// writes. We also needn't bother fetching the same file concurrently,
// since we'll be able to reuse it in the second caller.
//
// This URL mutex is shared among multiple VFS implementations in this
// package.

var (
	urlMusMu sync.Mutex
	urlMus   = map[string]*sync.Mutex{}
)

func urlMu(path string) *sync.Mutex {
	urlMusMu.Lock()
	mu, ok := urlMus[path]
	if !ok {
		mu = new(sync.Mutex)
		urlMus[path] = mu
	}
	urlMusMu.Unlock()
	return mu
}

// random will create a file of size bytes (rounded up to next 1024 size)
func random_971(size int) error {
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
