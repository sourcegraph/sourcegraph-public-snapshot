// Package bg defines init and background tasks.
//
// Because there can be multiple frontend processes running, these tasks
// must be idempotent and not racy amid concurrent frontend processes.
package bg
