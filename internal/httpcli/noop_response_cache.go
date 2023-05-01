package httpcli

type NoopCache struct{}

func (c NoopCache) Get(key string) (responseBytes []byte, ok bool) { return nil, false }
func (c NoopCache) Set(key string, responseBytes []byte)           {}
func (c NoopCache) Delete(key string)                              {}
