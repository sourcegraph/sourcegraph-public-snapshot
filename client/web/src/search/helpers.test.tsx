import * as H from 'history'
import { getSearchTypeFromQuery, toggleSearchType, toggleSearchFilter, submitSearch } from './helpers'
import { SearchType } from './results/SearchResults'
import { SearchPatternType } from '../../../shared/src/graphql/schema'

jest.mock('../tracking/eventLogger', () => ({ eventLogger: { log: () => undefined } }))

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

    describe('submitSearch()', () => {
        test('should update history', () => {
            const history = H.createMemoryHistory({})
            submitSearch({
                history,
                query: 'querystring',
                patternType: SearchPatternType.literal,
                caseSensitive: false,
                versionContext: undefined,
                selectedSearchContextSpec: 'global',
                activation: undefined,
                source: 'home',
                searchParameters: undefined,
            })
            expect(history.location.search).toEqual('?q=context:global+querystring&patternType=literal')
        })
        test('should keep trace param when updating history', () => {
            const history = H.createMemoryHistory({ initialEntries: ['/?trace=1'] })
            submitSearch({
                history,
                query: 'querystring',
                patternType: SearchPatternType.literal,
                caseSensitive: false,
                versionContext: undefined,
                selectedSearchContextSpec: 'global',
                activation: undefined,
                source: 'home',
                searchParameters: undefined,
            })
            expect(history.location.search).toEqual('?q=context%3Aglobal+querystring&patternType=literal&trace=1')
        })
    })

    const searchTypes: NonNullable<SearchType>[] = ['diff', 'commit', 'symbol', 'repo', 'path']

    describe('getSearchTypeFromQuery()', () => {
        test('parses the search type in simple queries', () => {
            for (const searchType of searchTypes) {
                expect(getSearchTypeFromQuery(`type:${searchType}`)).toEqual(searchType)
            }
        })

        test('returns null when no search type specified', () => {
            expect(getSearchTypeFromQuery('code')).toEqual(null)
        })

        test('parses the search type in complex queries', () => {
            expect(getSearchTypeFromQuery('test type:diff')).toEqual('diff')
            expect(getSearchTypeFromQuery('type:diff test')).toEqual('diff')
            expect(getSearchTypeFromQuery('repo:^github.com/sourcegraph/sourcegraph type:diff test')).toEqual('diff')
            expect(getSearchTypeFromQuery('type:diff repo:^github.com/sourcegraph/sourcegraph test')).toEqual('diff')
        })

        test('returns symbol when multiple search types, including symbol, are specified', () => {
            // Edge case. If there are multiple type filters and `type:symbol` is one of them, symbol results always get returned.
            expect(
                getSearchTypeFromQuery('type:diff type:symbol repo:^github.com/sourcegraph/sourcegraph test')
            ).toEqual('symbol')
        })

        test('returns the first search type specified when multiple search types, not including symbol, are specified', () => {
            expect(
                getSearchTypeFromQuery('type:diff type:commit repo:^github.com/sourcegraph/sourcegraph test')
            ).toEqual('diff')
        })
    })

    describe('toggleSearchType()', () => {
        test('returns the original query when the query already contains the correct type', () => {
            expect(toggleSearchType('test', null)).toEqual('test')

            for (const searchType of searchTypes) {
                expect(toggleSearchType(`test type:${searchType}`, searchType)).toEqual(`test type:${searchType}`)
            }
        })

        test('appends type:$TYPE to the query when no type exists in the query', () => {
            for (const searchType of searchTypes) {
                expect(toggleSearchType('test', searchType)).toEqual(`test type:${searchType}`)
            }
        })

        test('replaces existing type in query with new type in simple queries', () => {
            expect(toggleSearchType('test type:commit', 'diff')).toEqual('test type:diff')
            expect(toggleSearchType('type:diff test', 'commit')).toEqual('type:commit test')
        })

        test('replaces existing type in query with new type in complex queries', () => {
            expect(toggleSearchType('test type:symbol repo:^sourcegraph/test', 'diff')).toEqual(
                'test type:diff repo:^sourcegraph/test'
            )
            expect(toggleSearchType('test type:symbol repo:^sourcegraph/test', null)).toEqual(
                'test  repo:^sourcegraph/test'
            )
        })

        test('replaces the first type in query with new type in queries with multiple type fields', () => {
            expect(toggleSearchType('test type:symbol type:commit repo:^sourcegraph/test', 'diff')).toEqual(
                'test type:diff type:commit repo:^sourcegraph/test'
            )
        })
    })

    describe('toggleSearchFilter', () => {
        it('adds filter if it is not already in query', () => {
            expect(toggleSearchFilter('repo:test ', 'lang:c++')).toStrictEqual('repo:test lang:c++ ')
        })

        it('adds filter if it is not already in query, even if it matches substring for an existing filter', () => {
            expect(toggleSearchFilter('repo:test lang:c++ ', 'lang:c')).toStrictEqual('repo:test lang:c++ lang:c ')
        })

        it('removes filter from query it it exists', () => {
            expect(toggleSearchFilter('repo:test lang:c++ lang:c ', 'lang:c')).toStrictEqual('repo:test lang:c++')
        })
    })
})
