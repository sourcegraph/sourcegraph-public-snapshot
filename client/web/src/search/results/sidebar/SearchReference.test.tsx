import { FILTERS, FilterType } from '@sourcegraph/shared/src/search/query/filters'
import { assert } from 'chai'
import { Selection } from 'monaco-editor'
import { QueryChangeSource, QueryState } from '../../../helpers'
import { parsePlaceholder, SearchReferenceInfo, updateQueryWithFilter } from './SearchReference'

/**
 * Automatically sets cursor position and selections from example query.
 */
function queryStateFromExample(query: string, showSuggestions = false): QueryState {
    let cursorPosition
    let selectionStart: number
    let selection: Selection
    let offset = 0

    const cleanedQuery = query.replaceAll(/[\[\]|]/g, (match, index) => {
        switch (match) {
            case '[':
                selectionStart = index
                offset -= 1
                break
            case ']':
                selection = new Selection(1, selectionStart + 1, 1, index + offset + 1)
                offset -= 1
                break
            case '|':
                cursorPosition = index + offset + 1
                break
        }
        return ''
    })

    return {
        query: cleanedQuery,
        changeSource: QueryChangeSource.searchReference,
        cursorPosition,
        showSuggestions,
        selection,
    }
}

type TestInput = [string, string, string, SearchReferenceInfo, typeof FILTERS]

function runTest([name, expectedResult, currentQuery, searchReference, filters]: TestInput): void {
    it(name, () => {
        assert.deepStrictEqual(
            updateQueryWithFilter({ query: currentQuery }, searchReference, filters),
            queryStateFromExample(expectedResult)
        )
    })
}

function createSearchReference(type: FilterType, placeholder: string) {
    return {
        type,
        placeholder: parsePlaceholder(placeholder),
        description: '',
    }
}

describe('repeatable filters', () => {
    it('appends placeholder filter and selects placeholder', () => {
        assert.deepStrictEqual(
            updateQueryWithFilter({ query: 'foo' }, createSearchReference(FilterType.after, '{test}'), FILTERS),
            queryStateFromExample('foo after:[test]')
        )
    })

    it('appends suggestions filter', () => {
        assert.deepStrictEqual(
            updateQueryWithFilter({ query: 'foo' }, createSearchReference(FilterType.lang, '{lang}'), FILTERS),
            queryStateFromExample('foo lang:', true)
        )
    })
})

describe('unique filters', () => {
    it('appends placeholder filter and selects placeholder', () => {
        assert.deepStrictEqual(
            updateQueryWithFilter({ query: 'foo' }, createSearchReference(FilterType.repogroup, '{test}'), FILTERS),
            queryStateFromExample('foo repogroup:[test]')
        )
    })

    it('updates existing placeholder filter and selects placeholder', () => {
        assert.deepStrictEqual(
            updateQueryWithFilter(
                { query: 'repogroup:value foo' },
                createSearchReference(FilterType.repogroup, '{test}'),
                FILTERS
            ),
            queryStateFromExample('repogroup:[value]| foo')
        )
    })

    it('appends suggestions filter', () => {
        assert.deepStrictEqual(
            updateQueryWithFilter({ query: 'foo' }, createSearchReference(FilterType.case, '{test}'), FILTERS),
            queryStateFromExample('foo case:', true)
        )
    })

    it('updates existing suggestions filter', () => {
        assert.deepStrictEqual(
            updateQueryWithFilter({ query: 'case:yes foo' }, createSearchReference(FilterType.case, '{test}'), FILTERS),
            queryStateFromExample('case:| foo', true)
        )
    })
})
