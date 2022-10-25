import { stub } from 'sinon'

import { Token } from '@sourcegraph/shared/src/search/query/token'

import { searchHistorySource } from './searchHistorySource'
import { buildMockTempSettings } from './useRecentSearches.test'

describe('searchHistorySource', () => {
    const suggestionContext = { position: 0, onAbort: stub() }

    test('returns null if no recent searches', () => {
        const recentSearches = buildMockTempSettings(0)
        const onSelection = stub()

        const source = searchHistorySource({ recentSearches, onSelection })
        expect(source(suggestionContext, [])).toBeNull()
    })

    test('returns null if there are any tokens', () => {
        const recentSearches = buildMockTempSettings(5)
        const onSelection = stub()

        const tokens: Token[] = [{ type: 'literal', value: 'test', quoted: false, range: { start: 0, end: 4 } }]

        const source = searchHistorySource({ recentSearches, onSelection })
        expect(source(suggestionContext, tokens)).toBeNull()
    })

    test('returns recent searches in the correct format', () => {
        const recentSearches = buildMockTempSettings(5)
        const onSelection = stub()

        const source = searchHistorySource({ recentSearches, onSelection })
        expect(source(suggestionContext, [])).toMatchInlineSnapshot(
            `
            Object {
              "filter": false,
              "from": 0,
              "options": Array [
                Object {
                  "apply": [Function],
                  "label": "test0",
                  "type": "searchhistory",
                },
                Object {
                  "apply": [Function],
                  "label": "test1",
                  "type": "searchhistory",
                },
                Object {
                  "apply": [Function],
                  "label": "test2",
                  "type": "searchhistory",
                },
                Object {
                  "apply": [Function],
                  "label": "test3",
                  "type": "searchhistory",
                },
                Object {
                  "apply": [Function],
                  "label": "test4",
                  "type": "searchhistory",
                },
              ],
            }
        `
        )
    })
})
