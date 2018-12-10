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
