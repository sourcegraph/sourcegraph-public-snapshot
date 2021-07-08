import { assert } from 'chai'
import { Range, Selection, SelectionDirection } from 'monaco-editor'

import { FILTERS, FilterType } from '@sourcegraph/shared/src/search/query/filters'

import { QueryChangeSource, QueryState } from '../../helpers'

import { parsePlaceholder, updateQueryWithFilter } from './SearchReference'

/**
 * Automatically sets cursor position and selections from example query.
 */
function queryStateFromExample(query: string, showSuggestions = false): QueryState {
    const positions: { [char: string]: number } = {}
    const magicCharacters = {
        cursor: '|',
        selectionStart: '[',
        selectionEnd: ']',
        rangeStart: '{',
        rangeEnd: '}',
    }
    let offset = 0

    const cleanedQuery = query.replaceAll(/[[\]{|}]/g, (match, index: number) => {
        positions[match] = index + offset + 1
        offset -= 1
        return ''
    })

    const { [magicCharacters.selectionStart]: selectionStart, [magicCharacters.cursor]: cursor } = positions

    const selection = Selection.createWithDirection(
        1,
        selectionStart,
        1,
        positions[magicCharacters.selectionEnd],
        cursor !== undefined && cursor <= selectionStart ? SelectionDirection.RTL : SelectionDirection.LTR
    )

    return {
        query: cleanedQuery,
        changeSource: QueryChangeSource.searchReference,
        selection,
        showSuggestions,
        revealRange: new Range(1, positions[magicCharacters.rangeStart], 1, positions[magicCharacters.rangeEnd]),
    }
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
            updateQueryWithFilter({ query: 'foo' }, createSearchReference(FilterType.after, '{test}'), false, FILTERS),
            queryStateFromExample('foo {after:[test]|}')
        )
    })

    it('appends suggestions filter', () => {
        assert.deepStrictEqual(
            updateQueryWithFilter({ query: 'foo' }, createSearchReference(FilterType.lang, '{lang}'), false, FILTERS),
            queryStateFromExample('foo {lang:|[lang]}', true)
        )
    })
})

describe('unique filters', () => {
    it('appends placeholder filter and selects placeholder', () => {
        assert.deepStrictEqual(
            updateQueryWithFilter(
                { query: 'foo' },
                createSearchReference(FilterType.repogroup, '{test}'),
                false,
                FILTERS
            ),
            queryStateFromExample('foo {repogroup:[test]|}')
        )
    })

    it('updates existing placeholder filter and selects value', () => {
        assert.deepStrictEqual(
            updateQueryWithFilter(
                { query: 'repogroup:value foo' },
                createSearchReference(FilterType.repogroup, '{test}'),
                false,
                FILTERS
            ),
            queryStateFromExample('{repogroup:[value]|} foo')
        )
    })

    it('appends suggestions filter', () => {
        assert.deepStrictEqual(
            updateQueryWithFilter({ query: 'foo' }, createSearchReference(FilterType.case, '{test}'), false, FILTERS),
            queryStateFromExample('foo {case:|[test]}', true)
        )
    })

    it('updates existing suggestions filter', () => {
        assert.deepStrictEqual(
            updateQueryWithFilter(
                { query: 'case:yes foo' },
                createSearchReference(FilterType.case, '{test}'),
                false,
                FILTERS
            ),
            queryStateFromExample('{case:|[yes]} foo', true)
        )
    })
})
