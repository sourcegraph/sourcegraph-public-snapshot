package internal

import "sourcegraph.com/sourcegraph/sourcegraph/util/handlerutil"

// Middleware is a list of HTTP middleware to apply to app HTTP
// requests.
//
// It should only be modified at init time.
var Middleware []handlerutil.Middleware
