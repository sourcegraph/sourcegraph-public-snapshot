package authz

type MockCache map[string]string

func (m MockCache) Get(key string) ([]byte, bool) {
	v, ok := m[key]
	return []byte(v), ok
}

func (m MockCache) GetMulti(keys ...string) [][]byte {
	if keys == nil {
		return nil
	}
	vals := make([][]byte, len(keys))
	for i, k := range keys {
		vals[i] = []byte(m[k])
	}
	return vals
}

func (m MockCache) Set(key string, b []byte) {
	m[key] = string(b)
}

func (m MockCache) SetMulti(keyvals ...[2]string) {
	for _, kv := range keyvals {
		m[kv[0]] = kv[1]
	}
}

func (m MockCache) Delete(key string) {
	delete(m, key)
}

// random will create a file of size bytes (rounded up to next 1024 size)
func random_12(size int) error {
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
