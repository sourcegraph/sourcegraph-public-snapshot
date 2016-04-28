// @flow weak

// trackPromise is a no-op. It used to be used to track the
// unresolved promises for server-side rendering, but we no
// longer need to do that.
//
// It should be left in until all calls to trackPromise are
// cleaned up. It was left in as a no-op to avoid huge merge
// conflicts with the open branches that use it.
export function trackPromise(p: Promise): void {}
