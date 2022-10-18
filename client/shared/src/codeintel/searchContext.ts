// NOTE(2022-09-08) We store global state at the module level because that was
// the easiest way to inline sourcegraph/code-intel-extensions into the main
// repository. The old extension code imported from the npm package
// 'sourcegraph' and it would have required a large refactoring to pass around
// the state for all methods. It would be nice to refactor the code one day to
// avoid storing state at the module level, but we had to deprecate extensions
// on a tight deadline so we decided not to do this refactoring during the
// initial migration.
let _searchContext: string | undefined
export function setCodeIntelSearchContext(newSearchContext: string | undefined): void {
    _searchContext = newSearchContext
}

export function searchContext(): string | undefined {
    return _searchContext
}
