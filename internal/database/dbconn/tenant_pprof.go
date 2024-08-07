package dbconn

import (
	"context"
	"runtime/pprof"
	"sync/atomic"

	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/tenant"
)

var pprofUniqID atomic.Int64
var pprofTenantlessQueries = func() *pprof.Profile {
	enabled := env.MustGetBool("SRC_PGSQL_PROFILE_TENANTLESS_QUERIES", false, "INTERNAL: Add pprof profile tenantless_queries for all stack traces which did not set tenant")
	if !enabled {
		return nil
	}
	return pprof.NewProfile("tenantless_queries")
}()

func pprofCheckTenantlessQuery(ctx context.Context) {
	if pprofTenantlessQueries == nil {
		return
	}

	if _, ok := tenant.FromContext(ctx); ok {
		return
	}

	// We want to track every stack trace, so need a unique value for the event
	eventValue := pprofUniqID.Add(1)

	// skip stack for Add and this function (2).
	pprofTenantlessQueries.Add(eventValue, 2)
}
