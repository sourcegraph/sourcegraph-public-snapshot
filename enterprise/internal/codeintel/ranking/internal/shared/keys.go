package shared

import (
	"fmt"
	"strings"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/conf"
)

// NewGraphKey creates a new root graph key. This key identifies all work related to ranking,
// including the SCIP export tasks.
func NewGraphKey(graphKey string) string {
	return encode(graphKey)
}

// NewDerivativeGraphKeyKey creates a new derivative graph key. This key identifies work related
// to ranking, excluding the SCIP export tasks, which are identified by the same root graph key
// but with different bucket or derivative graph key prefix values.
func NewDerivativeGraphKeyKey(graphKey, derivativeGraphKeyPrefix string, bucket int64) string {
	return fmt.Sprintf("%s.%s-%d",
		encode(graphKey),
		encode(derivativeGraphKeyPrefix),
		bucket,
	)
}

// GraphKey returns a graph key from the configured root.
func GraphKey() string {
	return NewGraphKey(conf.CodeIntelRankingDocumentReferenceCountsGraphKey())
}

// DerivativeGraphKeyFromTime returns a derivative key from the configured root used for exports
// as well as the current "bucket" of time containing the current instant. Each bucket of time is
// the same configurable length, packed end-to-end since the Unix epoch.
//
// Constructing a graph key for the mapper and reducer jobs in this way ensures that begin a fresh
// map/reduce job on a periodic cadence (equal to the bucket length). Changing the root graph key
// will also create a new map/reduce job (without switching buckets).
func DerivativeGraphKeyFromTime(now time.Time) string {
	graphKey := conf.CodeIntelRankingDocumentReferenceCountsGraphKey()
	derivativeGraphKeyPrefix := conf.CodeIntelRankingDocumentReferenceCountsDerivativeGraphKeyPrefix()
	bucket := now.UTC().Unix() / int64(conf.CodeIntelRankingStaleResultAge().Seconds())

	return NewDerivativeGraphKeyKey(graphKey, derivativeGraphKeyPrefix, bucket)
}

// GraphKeyFromDerivativeGraphKey returns the root of the given derivative graph key.
func GraphKeyFromDerivativeGraphKey(derivativeGraphKey string) (string, bool) {
	parts := strings.Split(derivativeGraphKey, ".")
	if len(parts) != 2 {
		return "", false
	}

	return parts[0], true
}

var replacer = strings.NewReplacer(
	".", "_",
	"-", "_",
)

func encode(s string) string {
	return replacer.Replace(s)
}
