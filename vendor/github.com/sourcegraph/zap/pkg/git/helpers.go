package git

import (
	"time"

	"github.com/sourcegraph/zap/internal/pkg/backoff"
)

func GitBackOff() backoff.BackOff {
	p := backoff.NewExponentialBackOff()
	p.InitialInterval = 300 * time.Millisecond
	p.MaxElapsedTime = 3 * time.Second
	return p
}
