package shared

import (
	"fmt"
	"strings"

	"github.com/sourcegraph/sourcegraph/internal/conf"
)

// NewGraphKey creates a new root graph key. This key identifies all work related to ranking,
// including the SCIP export tasks.
func NewGraphKey(graphKey string) string {
	return encode(graphKey)
}

// NewDerivativeGraphKey creates a new derivative graph key. This key identifies work related
// to ranking, excluding the SCIP export tasks, which are identified by the same root graph key
// but with different derivative graph key prefix values.
func NewDerivativeGraphKey(graphKey, derivativeGraphKeyPrefix string) string {
	return fmt.Sprintf("%s.%s",
		encode(graphKey),
		encode(derivativeGraphKeyPrefix),
	)
}

// GraphKey returns a graph key from the configured root.
func GraphKey() string {
	return NewGraphKey(conf.CodeIntelRankingDocumentReferenceCountsGraphKey())
}

// DerivativeGraphKeyFromPrefix returns a derivative key from the configured root used for exports
// as well as the current "bucket" of time containing the current instant identified by the given
// prefix.
//
// Constructing a graph key for the mapper and reducer jobs in this way ensures that begin a fresh
// map/reduce job on a periodic cadence (determined by a cron-like site config setting). Changing
// the root graph key will also create a new map/reduce job.
func DerivativeGraphKeyFromPrefix(derivativeGraphKeyPrefix string) string {
	return NewDerivativeGraphKey(conf.CodeIntelRankingDocumentReferenceCountsGraphKey(), derivativeGraphKeyPrefix)
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
