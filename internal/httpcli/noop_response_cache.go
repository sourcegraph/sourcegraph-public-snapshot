package httpcli

import "context"

type NoopCache struct{}

func (c NoopCache) Get(ctx context.Context, key string) (responseBytes []byte, ok bool) {
	return nil, false
}
func (c NoopCache) Set(ctx context.Context, key string, responseBytes []byte) {}
func (c NoopCache) Delete(ctx context.Context, key string)                    {}
