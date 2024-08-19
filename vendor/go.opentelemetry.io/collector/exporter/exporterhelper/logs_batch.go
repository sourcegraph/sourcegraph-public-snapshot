// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package exporterhelper // import "go.opentelemetry.io/collector/exporter/exporterhelper"

import (
	"context"
	"errors"

	"go.opentelemetry.io/collector/exporter/exporterbatcher"
	"go.opentelemetry.io/collector/pdata/plog"
)

// mergeLogs merges two logs requests into one.
func mergeLogs(_ context.Context, r1 Request, r2 Request) (Request, error) {
	lr1, ok1 := r1.(*logsRequest)
	lr2, ok2 := r2.(*logsRequest)
	if !ok1 || !ok2 {
		return nil, errors.New("invalid input type")
	}
	lr2.ld.ResourceLogs().MoveAndAppendTo(lr1.ld.ResourceLogs())
	return lr1, nil
}

// mergeSplitLogs splits and/or merges the logs into multiple requests based on the MaxSizeConfig.
func mergeSplitLogs(_ context.Context, cfg exporterbatcher.MaxSizeConfig, r1 Request, r2 Request) ([]Request, error) {
	var (
		res          []Request
		destReq      *logsRequest
		capacityLeft = cfg.MaxSizeItems
	)
	for _, req := range []Request{r1, r2} {
		if req == nil {
			continue
		}
		srcReq, ok := req.(*logsRequest)
		if !ok {
			return nil, errors.New("invalid input type")
		}
		if srcReq.ld.LogRecordCount() <= capacityLeft {
			if destReq == nil {
				destReq = srcReq
			} else {
				srcReq.ld.ResourceLogs().MoveAndAppendTo(destReq.ld.ResourceLogs())
			}
			capacityLeft -= destReq.ld.LogRecordCount()
			continue
		}

		for {
			extractedLogs := extractLogs(srcReq.ld, capacityLeft)
			if extractedLogs.LogRecordCount() == 0 {
				break
			}
			capacityLeft -= extractedLogs.LogRecordCount()
			if destReq == nil {
				destReq = &logsRequest{ld: extractedLogs, pusher: srcReq.pusher}
			} else {
				extractedLogs.ResourceLogs().MoveAndAppendTo(destReq.ld.ResourceLogs())
			}
			// Create new batch once capacity is reached.
			if capacityLeft == 0 {
				res = append(res, destReq)
				destReq = nil
				capacityLeft = cfg.MaxSizeItems
			}
		}
	}

	if destReq != nil {
		res = append(res, destReq)
	}
	return res, nil
}

// extractLogs extracts logs from the input logs and returns a new logs with the specified number of log records.
func extractLogs(srcLogs plog.Logs, count int) plog.Logs {
	destLogs := plog.NewLogs()
	srcLogs.ResourceLogs().RemoveIf(func(srcRL plog.ResourceLogs) bool {
		if count == 0 {
			return false
		}
		needToExtract := resourceLogsCount(srcRL) > count
		if needToExtract {
			srcRL = extractResourceLogs(srcRL, count)
		}
		count -= resourceLogsCount(srcRL)
		srcRL.MoveTo(destLogs.ResourceLogs().AppendEmpty())
		return !needToExtract
	})
	return destLogs
}

// extractResourceLogs extracts resource logs and returns a new resource logs with the specified number of log records.
func extractResourceLogs(srcRL plog.ResourceLogs, count int) plog.ResourceLogs {
	destRL := plog.NewResourceLogs()
	destRL.SetSchemaUrl(srcRL.SchemaUrl())
	srcRL.Resource().CopyTo(destRL.Resource())
	srcRL.ScopeLogs().RemoveIf(func(srcSL plog.ScopeLogs) bool {
		if count == 0 {
			return false
		}
		needToExtract := srcSL.LogRecords().Len() > count
		if needToExtract {
			srcSL = extractScopeLogs(srcSL, count)
		}
		count -= srcSL.LogRecords().Len()
		srcSL.MoveTo(destRL.ScopeLogs().AppendEmpty())
		return !needToExtract
	})
	return destRL
}

// extractScopeLogs extracts scope logs and returns a new scope logs with the specified number of log records.
func extractScopeLogs(srcSL plog.ScopeLogs, count int) plog.ScopeLogs {
	destSL := plog.NewScopeLogs()
	destSL.SetSchemaUrl(srcSL.SchemaUrl())
	srcSL.Scope().CopyTo(destSL.Scope())
	srcSL.LogRecords().RemoveIf(func(srcLR plog.LogRecord) bool {
		if count == 0 {
			return false
		}
		srcLR.MoveTo(destSL.LogRecords().AppendEmpty())
		count--
		return true
	})
	return destSL
}

// resourceLogsCount calculates the total number of log records in the plog.ResourceLogs.
func resourceLogsCount(rl plog.ResourceLogs) int {
	count := 0
	for k := 0; k < rl.ScopeLogs().Len(); k++ {
		count += rl.ScopeLogs().At(k).LogRecords().Len()
	}
	return count
}
