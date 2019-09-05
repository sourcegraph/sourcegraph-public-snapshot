import { getSearchTypeFromQuery, toggleSearchType } from './helpers'

describe('search/helpers', () => {
    describe('queryIndexOfScope()', () => {
        test.skip('should return the index of a scope if contained in the query', () => {
            /* noop */
        })
        test.skip('should return the index of a scope if at the beginning of the query', () => {
            /* noop */
        })
        test.skip('should return the -1 if the scope is not contained in the query', () => {
            /* noop */
        })
        test.skip('should return the -1 if the scope is contained as a substring of another scope', () => {
            /* noop */
        })
    })

    describe('getSearchTypeFromQuery()', () => {
        expect(getSearchTypeFromQuery('type:diff')).toEqual('diff')
        expect(getSearchTypeFromQuery('type:commit')).toEqual('commit')
        expect(getSearchTypeFromQuery('type:symbol')).toEqual('symbol')
        expect(getSearchTypeFromQuery('type:repo')).toEqual('repo')
        expect(getSearchTypeFromQuery('code')).toEqual(null)

        expect(getSearchTypeFromQuery('test type:diff')).toEqual('diff')
        expect(getSearchTypeFromQuery('type:diff test')).toEqual('diff')
        expect(getSearchTypeFromQuery('repo:^github.com/sourcegraph/sourcegraph type:diff test')).toEqual('diff')
        expect(getSearchTypeFromQuery('type:diff repo:^github.com/sourcegraph/sourcegraph test')).toEqual('diff')
        expect(getSearchTypeFromQuery('type:diff type:commit repo:^github.com/sourcegraph/sourcegraph test')).toEqual(
            'commit'
        )
        /** Edge case. If there are multiple type filters and `type:symbol` is one of them, symbol results always get returned. */
        expect(getSearchTypeFromQuery('type:diff type:symbol repo:^github.com/sourcegraph/sourcegraph test')).toEqual(
            'symbol'
        )
    })

    describe('toggleSearchType()', () => {
        expect(toggleSearchType('test', null)).toEqual('test')
        expect(toggleSearchType('test type:diff', 'diff')).toEqual('test type:diff')
        expect(toggleSearchType('test type:commit', 'commit')).toEqual('test type:commit')
        expect(toggleSearchType('test type:symbol', 'symbol')).toEqual('test type:symbol')
        expect(toggleSearchType('test type:repo', 'repo')).toEqual('test type:repo')

        expect(toggleSearchType('test', 'diff')).toEqual('test type:diff')
        expect(toggleSearchType('test', 'commit')).toEqual('test type:commit')
        expect(toggleSearchType('test', 'symbol')).toEqual('test type:symbol')
        expect(toggleSearchType('test', 'repo')).toEqual('test type:repo')

        expect(toggleSearchType('test type:commit', 'diff')).toEqual('test type:diff')
        expect(toggleSearchType('type:diff test', 'commit')).toEqual('type:commit test')
        expect(toggleSearchType('test type:symbol repo:^sourcegraph/test', 'diff')).toEqual(
            'test type:diff repo:^sourcegraph/test'
        )
        expect(toggleSearchType('test type:symbol repo:^sourcegraph/test', null)).toEqual(
            'test  repo:^sourcegraph/test'
        )
        expect(toggleSearchType('test type:symbol repo:^sourcegraph/test', null)).toEqual(
            'test  repo:^sourcegraph/test'
        )
    })
})
