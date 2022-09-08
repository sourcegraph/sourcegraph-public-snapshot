import { SearchPatternType } from '@sourcegraph/shared/src/graphql-operations'

import { setQueryStateFromSettings, setQueryStateFromURL, useNavbarQueryState } from './navbarSearchQueryState'

describe('navbar query state', () => {
    describe('set state from settings', () => {
        it('sets default search pattern', () => {
            setQueryStateFromSettings({
                subjects: [],
                final: {
                    'search.defaultPatternType': SearchPatternType.regexp,
                },
            })

            expect(useNavbarQueryState.getState()).toHaveProperty('searchPatternType', SearchPatternType.regexp)
        })

        it('sets default case sensitivity', () => {
            setQueryStateFromSettings({
                subjects: [],
                final: {
                    'search.defaultCaseSensitive': true,
                },
            })

            expect(useNavbarQueryState.getState()).toHaveProperty('searchCaseSensitivity', true)
        })
    })

    describe('set state from URL', () => {
        it('sets the search pattern from URL parementer', () => {
            setQueryStateFromURL('q=context:global+&patternType=regexp')

            expect(useNavbarQueryState.getState()).toHaveProperty('searchPatternType', SearchPatternType.regexp)
        })

        it('sets the search pattern from filter', () => {
            setQueryStateFromURL('q=context:global+&patternType=regexp')

            expect(useNavbarQueryState.getState()).toHaveProperty('searchPatternType', SearchPatternType.regexp)
        })

        it('sets case sensitivity from filter', () => {
            setQueryStateFromURL('q=context:global+case:yes')

            expect(useNavbarQueryState.getState()).toHaveProperty('searchCaseSensitivity', true)
        })

        it('sets case sensitivity from URL paramster', () => {
            setQueryStateFromURL('q=context:global+&case=yes')

            expect(useNavbarQueryState.getState()).toHaveProperty('searchCaseSensitivity', true)
        })
    })

    describe('state initialization precedence', () => {
        // Note that the other tests already verify that user settings and URL
        // settings can override defaults

        it('prefers settings from URL over user settings', () => {
            setQueryStateFromSettings({
                subjects: [],
                final: {
                    'search.defaultPatternType': SearchPatternType.structural,
                },
            })
            setQueryStateFromURL('q=context:global+&patternType=regexp')

            expect(useNavbarQueryState.getState()).toHaveProperty('searchPatternType', SearchPatternType.regexp)
        })
        it('prefers user settings over settings from empty URL', () => {
            setQueryStateFromURL('')
            setQueryStateFromSettings({
                subjects: [],
                final: {
                    'search.defaultPatternType': SearchPatternType.structural,
                },
            })

            expect(useNavbarQueryState.getState()).toHaveProperty('searchPatternType', SearchPatternType.structural)
        })

        it('does not prefer user settings over settings from URL', () => {
            setQueryStateFromURL('q=context:global+&patternType=regexp')
            setQueryStateFromSettings({
                subjects: [],
                final: {
                    'search.defaultPatternType': SearchPatternType.structural,
                },
            })

            expect(useNavbarQueryState.getState()).toHaveProperty('searchPatternType', SearchPatternType.regexp)
        })
    })
})
