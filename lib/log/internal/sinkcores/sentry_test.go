package sinkcores_test

import (
	"sync"
	"testing"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/log"
	"github.com/sourcegraph/sourcegraph/lib/log/internal/sinkcores"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zapcore"
)

type TransportMock struct {
	mu        sync.Mutex
	events    []*sentry.Event
	lastEvent *sentry.Event
}

func (t *TransportMock) Configure(options sentry.ClientOptions) {}
func (t *TransportMock) SendEvent(event *sentry.Event) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.events = append(t.events, event)
	t.lastEvent = event
}
func (t *TransportMock) Flush(timeout time.Duration) bool {
	return true
}
func (t *TransportMock) Events() []*sentry.Event {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.events
}

func newTestHub(t *testing.T) (*sentry.Hub, *TransportMock) {
	transport := &TransportMock{}
	c, err := sentry.NewClient(sentry.ClientOptions{Transport: transport})
	assert.NoError(t, err)
	hub := sentry.NewHub(c, sentry.NewScope())
	return hub, transport
}

func TestSentryCore(t *testing.T) {
	t.Run("INFO, no error", func(t *testing.T) {
		hub, tr := newTestHub(t)
		core := sinkcores.NewSentryCore(hub)
		core.Start()
		err := core.Write(zapcore.Entry{Level: zapcore.InfoLevel, Message: "msg"}, nil)
		assert.NoError(t, err)
		core.Sync()
		assert.Len(t, tr.Events(), 0)
	})
	t.Run("WARN, 1 error", func(t *testing.T) {
		hub, tr := newTestHub(t)
		core := sinkcores.NewSentryCore(hub)
		core.Start()
		err := core.Write(zapcore.Entry{Level: zapcore.WarnLevel, Message: "msg"}, []zapcore.Field{log.Error(errors.New("foobar"))})
		assert.NoError(t, err)
		core.Sync()
		assert.Len(t, tr.Events(), 1)
	})
}

// BenchmarkWrite-10         924944              1842 ns/op
func BenchmarkWrite(b *testing.B) {
	c := sinkcores.NewSentryCore(nil)
	err := errors.New("foobar")
	for n := 0; n < b.N; n++ {
		c.With([]zapcore.Field{log.Error(err)}).Write(zapcore.Entry{Message: "msg"}, []zapcore.Field{log.Int("key", 5)})
	}
}

// func init() {
// 	log.Init(log.Resource{Name: "bench"})
// }
//
// // BenchmarkNormal-10        296174              4331 ns/op
// func BenchmarkNormal(b *testing.B) {
// 	logger := globallogger.Get(false)
// 	err := errors.New("foobar")
// 	for n := 0; n < b.N; n++ {
// 		logger.With(log.Error(err), log.Int("key", 5)).Info("msg")
// 	}
// }
