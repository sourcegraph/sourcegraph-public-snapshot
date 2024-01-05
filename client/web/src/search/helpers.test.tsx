import * as H from 'history'
import { describe, expect, test } from 'vitest'

import { SearchPatternType } from '../graphql-operations'

import { submitSearch } from './helpers'

describe('search/helpers', () => {
    describe('submitSearch()', () => {
        test('should update history', () => {
            const history = H.createMemoryHistory({})
            submitSearch({
                historyOrNavigate: history,
                location: history.location,
                query: 'querystring',
                patternType: SearchPatternType.standard,
                caseSensitive: false,
                selectedSearchContextSpec: 'global',
                source: 'home',
            })
            expect(history.location.search).toMatchInlineSnapshot(
                '"?q=context:global+querystring&patternType=standard&sm=0"'
            )
        })
        test('should keep trace param when updating history', () => {
            const history = H.createMemoryHistory({ initialEntries: ['/?trace=1'] })
            submitSearch({
                historyOrNavigate: history,
                location: history.location,
                query: 'querystring',
                patternType: SearchPatternType.standard,
                caseSensitive: false,
                selectedSearchContextSpec: 'global',
                source: 'home',
            })
            expect(history.location.search).toMatchInlineSnapshot(
                '"?q=context:global+querystring&patternType=standard&sm=0&trace=1"'
            )
        })
    })
})
