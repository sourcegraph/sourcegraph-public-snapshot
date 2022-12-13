package goroutine

import (
	"time"

	"github.com/inconshreveable/log15"
)

type Monitorable interface {
	BackgroundRoutine
	Name() string
	Description() string
	JobName() string
	SetJobName(string)
	RegisterMonitor(monitor *Monitor)
}

type Monitor interface {
	Register(r Monitorable)
	LogRun(r Monitorable, shutdown bool, err error, duration time.Duration)
	LogStart(r Monitorable)
	LogStop(r Monitorable)
}

type RedisMonitor struct {
	Monitor
	routines []Monitorable
}

func NewRedisMonitor() *RedisMonitor {
	return &RedisMonitor{}
}

func (m *RedisMonitor) Register(r Monitorable) {
	m.routines = append(m.routines, r)

}

func (m *RedisMonitor) LogRun(r Monitorable, shutdown bool, err error, duration time.Duration) {
	println("LogRun" + r.Name())
}

func (m *RedisMonitor) LogStart(r Monitorable, err error) {
	log15.Error("LogStart" + r.Name())
}

func (m *RedisMonitor) LogStop(r Monitorable) {
	log15.Error("LogStop" + r.Name())
}
