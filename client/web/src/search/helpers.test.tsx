import * as H from 'history'

import { SearchPatternType } from '../graphql-operations'

import { submitSearch } from './helpers'

describe('search/helpers', () => {
    describe('submitSearch()', () => {
        test('should update history', () => {
            const history = H.createMemoryHistory({})
            submitSearch({
                history,
                query: 'querystring',
                patternType: SearchPatternType.standard,
                caseSensitive: false,
                selectedSearchContextSpec: 'global',
                source: 'home',
                searchParameters: undefined,
            })
            expect(history.location.search).toMatchInlineSnapshot(
                '"?q=context:global+querystring&patternType=standard"'
            )
        })
        test('should keep trace param when updating history', () => {
            const history = H.createMemoryHistory({ initialEntries: ['/?trace=1'] })
            submitSearch({
                history,
                query: 'querystring',
                patternType: SearchPatternType.standard,
                caseSensitive: false,
                selectedSearchContextSpec: 'global',
                source: 'home',
                searchParameters: undefined,
            })
            expect(history.location.search).toMatchInlineSnapshot(
                '"?q=context%3Aglobal+querystring&patternType=standard&trace=1"'
            )
        })
    })
})
