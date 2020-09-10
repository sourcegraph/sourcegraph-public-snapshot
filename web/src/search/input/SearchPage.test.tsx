import React from 'react'
import { cleanup, render } from '@testing-library/react'
import { createMemoryHistory } from 'history'
import { NOOP_TELEMETRY_SERVICE } from '../../../../shared/src/telemetry/telemetryService'
import { SearchPage, SearchPageProps } from './SearchPage'
import { SearchPatternType } from '../../graphql-operations'
import { Services } from '../../../../shared/src/api/client/services'
import { ThemePreference } from '../../theme'
import { of } from 'rxjs'

// Mock the Monaco input box to make this a shallow test
jest.mock('./SearchPageInput', () => ({
    SearchPageInput: () => null,
}))

describe('SearchPage', () => {
    afterAll(cleanup)

    let container: HTMLElement

    const history = createMemoryHistory()
    const defaultProps: SearchPageProps = {
        isSourcegraphDotCom: false,
        settingsCascade: {
            final: null,
            subjects: null,
        },
        location: history.location,
        history,
        extensionsController: {
            services: {} as Services,
        } as any,
        telemetryService: NOOP_TELEMETRY_SERVICE,
        themePreference: ThemePreference.Light,
        onThemePreferenceChange: () => undefined,
        authenticatedUser: null,
        setVersionContext: () => undefined,
        availableVersionContexts: [],
        globbing: false,
        patternType: SearchPatternType.literal,
        setPatternType: () => undefined,
        caseSensitive: false,
        setCaseSensitivity: () => undefined,
        platformContext: {} as any,
        keyboardShortcuts: [],
        filtersInQuery: {} as any,
        onFiltersInQueryChange: () => undefined,
        splitSearchModes: false,
        interactiveSearchMode: false,
        toggleSearchMode: () => undefined,
        copyQueryButton: false,
        versionContext: undefined,
        showRepogroupHomepage: false,
        showEnterpriseHomePanels: false,
        showOnboardingTour: false,
        isLightTheme: true,
        fetchSavedSearches: () => of([]),
        fetchRecentSearches: () => of({ nodes: [], totalCount: 0, pageInfo: { hasNextPage: false, endCursor: null } }),
    }

    it('should not show enterprise home panels if on Sourcegraph.com', () => {
        container = render(<SearchPage {...defaultProps} isSourcegraphDotCom={true} />).container
        const enterpriseHomePanels = container.querySelector('.enterprise-home-panels')
        expect(enterpriseHomePanels).toBeFalsy()
    })

    it('should not show enterprise home panels if showEnterpriseHomePanels disabled', () => {
        container = render(<SearchPage {...defaultProps} />).container
        const enterpriseHomePanels = container.querySelector('.enterprise-home-panels')
        expect(enterpriseHomePanels).toBeFalsy()
    })

    it('should show enterprise home panels if showEnterpriseHomePanels enabled and not on Sourcegraph.com', () => {
        container = render(<SearchPage {...defaultProps} showEnterpriseHomePanels={true} />).container
        const enterpriseHomePanels = container.querySelector('.enterprise-home-panels')
        expect(enterpriseHomePanels).toBeTruthy()
    })
})
