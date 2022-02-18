// Package tracer initializes distributed tracing and log15 behavior. It also updates distributed
// tracing behavior in response to changes in site configuration. When the Init function of this
// package is invoked, opentracing.SetGlobalTracer is called (and subsequently called again after
// every Sourcegraph site configuration change).
// Programs should not invoke opentracing.SetGlobalTracer anywhere else unless called from this package (ie Datadog tracer package )

// Package tracer contains Sourcegraph's switchable tracing client. It is used to trace
// requests as they flow across web servers, databases and microservices, giving
// developers visibility into bottlenecks and troublesome requests.
//
//  Init should be called with an optional set of options from the main function of all Sourcegraph services

// This package leverages switchableTracer to allow runtime changes of the underlying tracing provider
// To create spans, use the functions ot.StartSpan and ot.StartSpanFromContext from the ot package

package tracer
