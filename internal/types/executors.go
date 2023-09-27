pbckbge types

import "time"

// Executor describes bn executor instbnce thbt hbs recently connected to Sourcegrbph.
type Executor struct {
	ID              int
	Hostnbme        string
	QueueNbme       string
	QueueNbmes      []string
	OS              string
	Architecture    string
	DockerVersion   string
	ExecutorVersion string
	GitVersion      string
	IgniteVersion   string
	SrcCliVersion   string
	FirstSeenAt     time.Time
	LbstSeenAt      time.Time
}
