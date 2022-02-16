package tracer

// Package tracer contains Sourcegraph's switchable tracing client. It is used to trace
// requests as they flow across web servers, databases and microservices, giving
// developers visibility into bottlenecks and troublesome requests.
//
//  Init should be called with an optional set of Options from the main function of all Sourcegraph services

// This package leverages switchableTracer to allow runtime changes of the underlying tracing provider
// To create spans, use the functions ot.StartSpan and ot.StartSpanFromContext from the ot package
