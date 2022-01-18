import { act } from '@testing-library/react-hooks'
import { of as mockOf } from 'rxjs'
import { delay as mockDelay } from 'rxjs/operators'

import { SearchPatternType } from '@sourcegraph/shared/src/graphql-operations'

import { setAct } from '../__mocks__/zustand'
import { parseSearchURL } from '../search'

import { useNavbarQueryState, initQueryState, setQueryStateFromURL } from './navbarSearchQueryState'

let mockSpecAvailable = true
// A delay was added to test the expected behavior before the validation results
// arrive.
jest.mock('../search/backend', () => ({
    isSearchContextAvailable: () => mockOf(mockSpecAvailable).pipe(mockDelay(10)),
}))

describe('global query state store', () => {
    beforeAll(() => {
        setAct(act)
        // @fkling: Somehow I wasn't able to get rxjs timer functions to work
        // without this.
        jest.useFakeTimers()
    })
    beforeEach(() => {
        mockSpecAvailable = true
        window.localStorage.clear()
    })

    describe('initQueryState', () => {
        it('uses `global` as default search context if no query is present', () => {
            initQueryState('', true)
            jest.runAllTimers()

            expect(useNavbarQueryState.getState()).toHaveProperty('selectedSearchContext', 'global')
        })

        it('uses the search context present in the query', () => {
            initQueryState('q=context:test+test', true)
            jest.runAllTimers()

            expect(useNavbarQueryState.getState()).toHaveProperty('selectedSearchContext', 'test')
        })

        it('uses `global` if the search context in the query is not available', () => {
            mockSpecAvailable = false
            initQueryState('q=context:notavailable+test', true)
            jest.runAllTimers()

            expect(useNavbarQueryState.getState()).toHaveProperty('selectedSearchContext', 'global')
        })

        it('restores the previously used search context if no query is present', () => {
            // Choose context
            initQueryState('q=context:mycontext', true)
            jest.runAllTimers()
            expect(useNavbarQueryState.getState()).toHaveProperty('selectedSearchContext', 'mycontext')

            // This part is used to ensure that the state is reset correctly
            useNavbarQueryState.setState({ selectedSearchContext: '' })
            expect(useNavbarQueryState.getState()).toHaveProperty('selectedSearchContext', '')

            // Verify that a new initialization would load the previou context
            initQueryState('', true)
            jest.runAllTimers()
            expect(useNavbarQueryState.getState()).toHaveProperty('selectedSearchContext', 'mycontext')
        })
    })

    describe('setQueryStateFromURL', () => {
        it('should update patternType if different between URL and context', () => {
            useNavbarQueryState.setState({ searchPatternType: SearchPatternType.literal })

            setQueryStateFromURL(parseSearchURL('q=r:golang/oauth2+test+f:travis&patternType=regexp'), false)

            expect(useNavbarQueryState.getState().searchPatternType).toBe(SearchPatternType.regexp)
        })

        it('should not update patternType if query is empty', () => {
            useNavbarQueryState.setState({ searchPatternType: SearchPatternType.literal })

            setQueryStateFromURL(parseSearchURL('q=&patternType=regexp'), false)

            expect(useNavbarQueryState.getState().searchPatternType).toBe(SearchPatternType.literal)
        })

        it('should update caseSensitive if different between URL and context', () => {
            useNavbarQueryState.setState({ searchCaseSensitivity: false })

            setQueryStateFromURL(parseSearchURL('q=r:golang/oauth2+test+f:travis+case:yes'), false)

            expect(useNavbarQueryState.getState().searchCaseSensitivity).toBe(true)
        })

        it('should not update caseSensitive if query is empty', () => {
            useNavbarQueryState.setState({ searchCaseSensitivity: false })

            setQueryStateFromURL(parseSearchURL('q=case:yes'), false)

            expect(useNavbarQueryState.getState().searchCaseSensitivity).toBe(false)
        })

        describe('search context', () => {
            beforeEach(() => {
                useNavbarQueryState.setState({ searchContextsEnabled: true })
            })

            it('should assume the provided context is valid before validation results arrive', () => {
                mockSpecAvailable = false

                setQueryStateFromURL(parseSearchURL('q=context:notvalid+test'), true)
                const state = useNavbarQueryState.getState()
                expect(state.queryState).toHaveProperty('query', 'test')

                jest.runAllTimers()
                expect(useNavbarQueryState.getState().queryState).toHaveProperty('query', 'context:notvalid test')
            })

            it('uses the search context present in the query', () => {
                setQueryStateFromURL(parseSearchURL('q=context:test+test'), true)
                jest.runAllTimers()

                expect(useNavbarQueryState.getState()).toHaveProperty('selectedSearchContext', 'test')
            })

            it('uses `global` if the search context in the query is not available', () => {
                mockSpecAvailable = false
                setQueryStateFromURL(parseSearchURL('q=context:notavailable+test'), true)
                jest.runAllTimers()

                expect(useNavbarQueryState.getState()).toHaveProperty('selectedSearchContext', 'global')
            })
        })
    })
})
