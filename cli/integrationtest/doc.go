// Package integrationtest starts a single test server during the duration of all tests,
// allowing there to be many quick integration checks for easily checkable things that can otherwise regress.
//
// Since each test reuses the same server, all tests should be idempotent and order-independent.
package integrationtest
