import {
    getSearchTypeFromQuery,
    toggleSearchType,
    filterStaticSuggestions,
    insertSuggestionInQuery,
    validFilterAndValueBeforeCursor,
    isFuzzyWordSearch,
    formatQueryForFuzzySearch,
    filterAliasForSearch,
    highlightInvalidFilters,
    toInvalidFilterHtml,
    QueryState,
} from './helpers'
import { SearchType } from './results/SearchResults'
import { searchFilterSuggestions } from './searchFilterSuggestions'
import { filterAliases, isolatedFuzzySearchFilters } from './input/Suggestion'
import { ContentEditableState } from './input/ContentEditableInput'

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

    const searchTypes: SearchType[] = ['diff', 'commit', 'symbol', 'repo', 'path']

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

        const getArchivedSuggestions = () =>
            filterStaticSuggestions({ query: 'archived:', cursorPosition: 9 }, searchFilterSuggestions)
        const getFilterSuggestionStartingWithR = () =>
            filterStaticSuggestions({ query: filterQuery, cursorPosition: 6 }, searchFilterSuggestions)

        describe(`${filterStaticSuggestions.name}()`, () => {
            test('filters suggestions for filters starting with "r"', () => {
                const filtersStartingWithR = searchFilterSuggestions.filters.values
                    .map(filter => filter.value)
                    .filter(filter => filter.startsWith('r'))
                expect(getFilterSuggestionStartingWithR().map(suggestion => suggestion.value)).toEqual(
                    expect.arrayContaining(filtersStartingWithR)
                )
            })

            test('filters suggestions for filter aliases', () => {
                const [{ value }] = filterStaticSuggestions({ query: 'l', cursorPosition: 1 }, searchFilterSuggestions)
                expect(value).toBe(filterAliases.l + ':')
            })

            test('does not throw for query ":"', () => {
                expect(() =>
                    filterStaticSuggestions({ query: ':', cursorPosition: 1 }, searchFilterSuggestions)
                ).not.toThrowError()
            })

            test('filters suggestions for word "test"', () => {
                expect(
                    filterStaticSuggestions({ query: filterQuery, cursorPosition: 4 }, searchFilterSuggestions)
                ).toHaveLength(0)
            })

            test('filters suggestions for the "archived:" filter', () => {
                const archivedSuggestions = getArchivedSuggestions()
                expect(archivedSuggestions).toEqual(expect.arrayContaining(searchFilterSuggestions.archived.values))
            })
        })

        describe(`${insertSuggestionInQuery.name}()`, () => {
            describe('inserts suggestions for a filter name', () => {
                const [suggestion] = getFilterSuggestionStartingWithR().filter(({ value }) => value === 'repo:')
                const { query: newQuery } = insertSuggestionInQuery('test r test', suggestion, 6)
                expect(newQuery).toBe(`test ${suggestion.value} test`)
            })
            test('inserts suggestion for a filter value', () => {
                const [suggestion] = getArchivedSuggestions()
                const { query: newQuery } = insertSuggestionInQuery('test archived: test', suggestion, 14)
                expect(newQuery).toBe(`test archived:${suggestion.value} test`)
            })
        })
    })

    describe(`${validFilterAndValueBeforeCursor.name}()`, () => {
        const query = 'archived:yes QueryInput'
        it('returns values when a filter value is being typed', () => {
            expect(validFilterAndValueBeforeCursor({ query, cursorPosition: 10 })).toEqual({
                filterIndex: 0,
                filterAndValue: 'archived:y',
                matchedFilter: 'archived',
                resolvedFilterType: 'archived',
                value: 'y',
            })
        })
        it('returns values when a filter is selected but no value char is typed yet', () => {
            expect(validFilterAndValueBeforeCursor({ query, cursorPosition: 9 })).toEqual({
                filterIndex: 0,
                filterAndValue: 'archived:',
                matchedFilter: 'archived',
                resolvedFilterType: 'archived',
                value: '',
            })
        })
        it('does not return a value if typed whitespace char', () => {
            expect(validFilterAndValueBeforeCursor({ query, cursorPosition: 13 })).toStrictEqual(null)
        })
        it('correctly resolves filter aliases', () => {
            const query = 'l:go'
            expect(validFilterAndValueBeforeCursor({ query, cursorPosition: query.length })).toEqual({
                filterIndex: 0,
                filterAndValue: query,
                matchedFilter: 'l',
                resolvedFilterType: 'lang',
                value: 'go',
            })
        })
        it('correctly formats negated filters', () => {
            const query = '-f:package.json'
            expect(validFilterAndValueBeforeCursor({ query, cursorPosition: query.length })).toEqual({
                filterIndex: 0,
                filterAndValue: query,
                matchedFilter: '-f',
                resolvedFilterType: 'file',
                value: 'package.json',
            })
        })
        it('returns correct values for filter query without a value', () => {
            const query = '-f'
            expect(validFilterAndValueBeforeCursor({ query, cursorPosition: query.length })).toEqual({
                filterIndex: 0,
                filterAndValue: query,
                matchedFilter: query,
                resolvedFilterType: 'file',
                value: '',
            })
        })
    })

    describe('isFuzzyWordSearch', () => {
        const query = 'Query lang:g'
        it('returns false if typing a filter value', () =>
            expect(isFuzzyWordSearch({ query, cursorPosition: 12 })).toBe(false))
        it('returns true if typing a non filter type or value', () =>
            expect(isFuzzyWordSearch({ query, cursorPosition: 5 })).toBe(true))
    })

    describe('formatQueryForFuzzySearch', () => {
        const formatForSearchWithFilter = (filter: string) =>
            formatQueryForFuzzySearch({
                query: `archived:Yes ${filter}:value Props`,
                // 19 is position until after ':value'
                cursorPosition: 19 + filter.length,
            })

        it('isolates filters that are in isolatedFuzzySearchFilters', () => {
            expect(isolatedFuzzySearchFilters.map(formatForSearchWithFilter)).toStrictEqual(
                isolatedFuzzySearchFilters.map(filterType => filterType + ':value')
            )
        })
        it('replaces filter being typed with its `filterAliasForSearch`', () => {
            expect(Object.keys(filterAliasForSearch).map(formatForSearchWithFilter)).toStrictEqual(
                Object.values(filterAliasForSearch).map(alias => `archived:Yes ${alias}:value Props`)
            )
        })
        it('return absolute filter if filter being typed is negated (e.g: `-file`)', () => {
            expect(
                formatQueryForFuzzySearch({
                    query: 'l:javascript -file:index.js archived:No',
                    cursorPosition: 27,
                })
            ).toBe('l:javascript file:index.js archived:No')
        })
    })

    describe(`${highlightInvalidFilters.name}()`, () => {
        const invalidKeyword = 'a1b2c3_'
        const invalidKeyword2 = 'd4e5f6_'

        it('returns a ContentEditableState when no invalid keywords are found', () => {
            const query = 'a valid query archived:Yes test'
            const cursorPosition = 2
            expect(highlightInvalidFilters({ query, cursorPosition })).toStrictEqual(
                new ContentEditableState({
                    content: query,
                    cursor: { nodeIndex: 0, index: cursorPosition },
                })
            )
        })
        describe('calculates new cursor position after matching an invalid filter keyword', () => {
            type HighlightInvalidFiltersMock = QueryState & { cursor: ContentEditableState['cursor'] }

            const queries = {
                invalidAtMiddle: {
                    title: 'With invalid keyword at the middle',
                    query: `test ${invalidKeyword}:value repo:github.com test`,
                },
                invalidAtStart: {
                    title: 'With invalid keyword at the start',
                    query: `${invalidKeyword}:value repo:github.com test`,
                },
                invalidAtEnd: {
                    title: 'With invalid keyword at the end',
                    query: `test repo:github.com ${invalidKeyword2}:value ${invalidKeyword}:value`,
                },
            }

            const run = ({ cursor, ...mock }: HighlightInvalidFiltersMock): void => {
                const state = highlightInvalidFilters(mock)
                expect(state.cursor).toStrictEqual(cursor)
            }

            describe('cursor is before the matched node', () => {
                test(queries.invalidAtMiddle.title, () =>
                    run({
                        query: queries.invalidAtMiddle.query,
                        cursorPosition: 4,
                        cursor: { nodeIndex: 0, index: 4 },
                    })
                )
                test(queries.invalidAtEnd.title, () =>
                    run({
                        query: queries.invalidAtEnd.query,
                        cursorPosition: 25,
                        cursor: { nodeIndex: 1, index: 4 },
                    })
                )
            })
            describe('cursor is in the matched node', () => {
                test(queries.invalidAtStart.title, () =>
                    run({
                        query: queries.invalidAtStart.query,
                        cursorPosition: 4,
                        cursor: { nodeIndex: 0, index: 4 },
                    })
                )
                test(queries.invalidAtMiddle.title, () =>
                    run({
                        query: queries.invalidAtMiddle.query,
                        cursorPosition: 7,
                        cursor: { nodeIndex: 1, index: 2 },
                    })
                )
                test(queries.invalidAtEnd.title, () =>
                    run({
                        query: queries.invalidAtEnd.query,
                        cursorPosition: 38,
                        cursor: { nodeIndex: 3, index: 3 },
                    })
                )
            })
            describe('cursor is after the matched node', () => {
                test(queries.invalidAtStart.title, () =>
                    run({
                        query: queries.invalidAtStart.query,
                        cursorPosition: 10,
                        cursor: { nodeIndex: 1, index: 3 },
                    })
                )
                test(queries.invalidAtMiddle.title, () =>
                    run({
                        query: queries.invalidAtMiddle.query,
                        cursorPosition: 18,
                        cursor: { nodeIndex: 2, index: 6 },
                    })
                )
            })
        })
        it('only highlights invalid filter keywords', () => {
            const { content: query } = highlightInvalidFilters({
                query: `${invalidKeyword}:value repo:git Props ${invalidKeyword2}:value count:1`,
                cursorPosition: 0,
            })
            const invalidHtml = toInvalidFilterHtml(invalidKeyword)
            const invalid2Html = toInvalidFilterHtml(invalidKeyword2)
            expect(query).toBe(`${invalidHtml}:value repo:git Props ${invalid2Html}:value count:1`)
        })
    })
})
