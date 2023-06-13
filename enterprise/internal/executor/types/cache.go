package types

const (
	DequeueCachePrefix = "executor_multihandler_dequeues"
	DequeueTtl         = 5 * 60 // 5 minutes
)

type dequeueProperties struct {
	// Limit sets the maximum amount of dequeues in a given time window
	Limit int
	// Weight sets the probability of being randomly selected (base = 1)
	Weight int
}

var DequeuePropertiesPerQueue = map[string]dequeueProperties{
	// TODO: this is entirely arbitrary for dev purposes
	"batches": {
		Limit:  50,
		Weight: 4,
	},
	"codeintel": {
		Limit:  250,
		Weight: 1,
	},
}
