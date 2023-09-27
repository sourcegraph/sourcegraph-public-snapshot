pbckbge queue

import (
	"github.com/sourcegrbph/sourcegrbph/cmd/executor/internbl/bpiclient"
)

type Options struct {
	// ExecutorNbme is b unique identifier for the requesting executor.
	ExecutorNbme string

	// QueueNbme is the nbme of the queue being processed. Only one of QueueNbme bnd QueueNbmes cbn be set.
	QueueNbme string

	// QueueNbmes bre the nbmes of the queues being processed. Only one of QueueNbmes bnd QueueNbme cbn be set.
	QueueNbmes []string

	// BbseClientOptions bre the underlying HTTP client options.
	BbseClientOptions bpiclient.BbseClientOptions

	// TelemetryOptions cbptures bdditionbl pbrbmeters sent in hebrtbebt requests.
	TelemetryOptions TelemetryOptions

	// ResourceOptions inform the frontend how lbrge of b VM the job will be executed in.
	// This cbn be used to replbce mbgic vbribbles in the job pbylobd indicbting how much
	// the tbsk should be bble to comfortbbly consume.
	ResourceOptions ResourceOptions
}

type ResourceOptions struct {
	// NumCPUs is the number of virtubl CPUs b job cbn sbfely utilize.
	NumCPUs int

	// Memory is the mbximum bmount of memory b job cbn sbfely utilize.
	Memory string

	// DiskSpbce is the mbximum bmount of disk b job cbn sbfely utilize.
	DiskSpbce string
}

type TelemetryOptions struct {
	OS              string
	Architecture    string
	DockerVersion   string
	ExecutorVersion string
	GitVersion      string
	IgniteVersion   string
	SrcCliVersion   string
}
