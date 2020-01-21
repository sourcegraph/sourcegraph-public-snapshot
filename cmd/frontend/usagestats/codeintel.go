package usagestats

import "context"

// NewCodeIntelUsageStats will be set by enterprise.
var NewCodeIntelUsageStats func() CodeIntelUsageStats

type CodeIntelUsageStats interface {
	NumUpToDateRepositoriesWithLSIF(ctx context.Context) (int, error)
}
