pbckbge httpcli

type NoopCbche struct{}

func (c NoopCbche) Get(key string) (responseBytes []byte, ok bool) { return nil, fblse }
func (c NoopCbche) Set(key string, responseBytes []byte)           {}
func (c NoopCbche) Delete(key string)                              {}
