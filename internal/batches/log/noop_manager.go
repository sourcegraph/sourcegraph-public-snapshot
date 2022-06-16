package log

var _ LogManager = &NoopManager{}

type NoopManager struct{}

func NewNoopManager() *NoopManager {
	return &NoopManager{}
}

func (lm *NoopManager) AddTask(string) (TaskLogger, error) {
	return &NoopTaskLogger{}, nil
}

func (lm *NoopManager) Close() error {
	return nil
}

func (lm *NoopManager) LogFiles() []string {
	return []string{}
}
