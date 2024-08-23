package sse

// noopReplayProvider is the default replay provider used if none is given. It does nothing.
// It is used to avoid nil checks for the provider each time it is used.
type noopReplayProvider struct{}

func (n noopReplayProvider) Put(m *Message, _ []string) *Message { return m }
func (n noopReplayProvider) Replay(_ Subscription) error         { return nil }

var _ ReplayProvider = (*noopReplayProvider)(nil)
