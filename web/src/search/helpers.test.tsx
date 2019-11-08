import {
    getSearchTypeFromQuery,
    toggleSearchType,
    filterSearchSuggestions,
    insertSuggestionInQuery,
    getFilterTypedBeforeCursor,
    isFuzzyWordSearch,
} from './helpers'
import { SearchType } from './results/SearchResults'
import { searchFilterSuggestions, filterAliases } from './searchFilterSuggestions'

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

    const searchTypes: SearchType[] = ['diff', 'commit', 'symbol', 'repo']

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

    describe('suggestions', () => {
        const filterQuery = 'test r test'

        const getArchivedSuggestions = () => filterSearchSuggestions('archived:', 9, searchFilterSuggestions)
        const getFilterSuggestionStartingWithR = () => filterSearchSuggestions(filterQuery, 6, searchFilterSuggestions)

        describe('filterSearchSuggestions()', () => {
            test('filters suggestions for filters starting with "r"', () => {
                const filtersStartingWithR = Object.keys(searchFilterSuggestions).filter(filter =>
                    filter.startsWith('r')
                )
                expect(getFilterSuggestionStartingWithR().map(suggestion => suggestion.value)).toEqual(
                    expect.arrayContaining(filtersStartingWithR)
                )
            })

            test('filters suggestions for filter aliases', () => {
                for (const [alias, filter] of Object.entries(filterAliases)) {
                    const [{ value }] = filterSearchSuggestions(alias, alias.length, searchFilterSuggestions)
                    expect(value).toBe(filter)
                }
            })

            test('does not throw for query ":"', () => {
                expect(() => filterSearchSuggestions(':', 1, searchFilterSuggestions)).not.toThrowError()
            })

            test('filters suggestions for word "test"', () => {
                expect(filterSearchSuggestions(filterQuery, 4, searchFilterSuggestions)).toHaveLength(0)
            })

            test('filters suggestions for the "archived:" filter', () => {
                const archivedSuggestions = getArchivedSuggestions()
                expect(archivedSuggestions).toEqual(expect.arrayContaining(searchFilterSuggestions.archived.values))
            })
        })

        describe('insertSuggestionInQuery()', () => {
            describe('inserts suggestions for a filter name', () => {
                const suggestion = getFilterSuggestionStartingWithR().filter(({ value: title }) => title === 'repo')[0]
                const { query: newQuery } = insertSuggestionInQuery('test r test', suggestion, 6)
                expect(newQuery).toBe(`test ${suggestion.value}: test`)
            })
            test('inserts suggestion for a filter value', () => {
                const [suggestion] = getArchivedSuggestions()
                const { query: newQuery } = insertSuggestionInQuery('test archived: test', suggestion, 14)
                expect(newQuery).toBe(`test archived:${suggestion.value} test`)
            })
        })
    })

    describe('getFilterTypedBeforeCursor', () => {
        const query = 'archived:yes QueryInput'
        it('returns values when a filter value is being typed', () => {
            expect(getFilterTypedBeforeCursor({ query, cursorPosition: 10 })).toStrictEqual({
                filterAndValue: 'archived:y',
                filter: 'archived',
            })
        })
        it('returns values when a filter is selected but no value char is typed yet', () => {
            expect(getFilterTypedBeforeCursor({ query, cursorPosition: 9 })).toStrictEqual({
                filterAndValue: 'archived:',
                filter: 'archived',
            })
            getFilterTypedBeforeCursor({ query, cursorPosition: 9 })
        })
        it('does not return a value if typed whitespace char', () => {
            expect(getFilterTypedBeforeCursor({ query, cursorPosition: 13 })).toStrictEqual({})
        })
    })

    describe('isFuzzyWordSearch', () => {
        const query = 'Query lang:g'
        it('returns false if typing a filter value', () =>
            expect(isFuzzyWordSearch({ query, cursorPosition: 12 })).toBe(false))
        it('returns true if typing a non filter type or value', () =>
            expect(isFuzzyWordSearch({ query, cursorPosition: 5 })).toBe(true))
    })
})
