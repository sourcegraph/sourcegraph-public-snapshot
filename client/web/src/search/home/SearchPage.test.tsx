import { cleanup } from '@testing-library/react'
import { createMemoryHistory } from 'history'

import { getDocumentNode } from '@sourcegraph/http-client'
import { NOOP_TELEMETRY_SERVICE } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { renderWithBrandedContext } from '@sourcegraph/shared/src/testing'
import { MockedTestProvider } from '@sourcegraph/shared/src/testing/apollo'
import {
    mockFetchAutoDefinedSearchContexts,
    mockFetchSearchContexts,
    mockGetUserSearchContextNamespaces,
} from '@sourcegraph/shared/src/testing/searchContexts/testHelpers'
import { extensionsController } from '@sourcegraph/shared/src/testing/searchTestHelpers'

import { SourcegraphContext } from '../../jscontext'
import { useExperimentalFeatures } from '../../stores'
import { ThemePreference } from '../../stores/themeState'
import {
    HOME_PANELS_QUERY,
    RECENTLY_SEARCHED_REPOSITORIES_TO_LOAD,
    RECENT_FILES_TO_LOAD,
    RECENT_SEARCHES_TO_LOAD,
} from '../panels/HomePanels'
import {
    authUser,
    collaboratorsPayload,
    recentFilesPayload,
    recentSearchesPayload,
    savedSearchesPayload,
} from '../panels/utils'

import { SearchPage, SearchPageProps } from './SearchPage'

// Mock the Monaco input box to make this a shallow test
jest.mock('./SearchPageInput', () => ({
    SearchPageInput: () => null,
}))

// Uses import.meta.url, which is a SyntaxError when used outside of ES Modules (Jest runs tests as
// CommonJS).
jest.mock('./LoggedOutHomepage.constants', () => ({
    fonts: [],
    exampleTripsAndTricks: [],
}))

function getMocks({
    enableSavedSearches,
    enableCollaborators,
}: {
    enableSavedSearches: boolean
    enableCollaborators: boolean
}) {
    return [
        {
            request: {
                query: getDocumentNode(HOME_PANELS_QUERY),
                variables: {
                    userId: '0',
                    firstRecentlySearchedRepositories: RECENTLY_SEARCHED_REPOSITORIES_TO_LOAD,
                    firstRecentSearches: RECENT_SEARCHES_TO_LOAD,
                    firstRecentFiles: RECENT_FILES_TO_LOAD,
                    enableSavedSearches,
                    enableCollaborators,
                },
            },
            result: {
                data: {
                    node: {
                        __typename: 'User',
                        recentlySearchedRepositoriesLogs: recentSearchesPayload(),
                        recentSearchesLogs: recentSearchesPayload(),
                        recentFilesLogs: recentFilesPayload(),
                        collaborators: enableCollaborators ? collaboratorsPayload() : undefined,
                    },
                    savedSearches: enableSavedSearches ? savedSearchesPayload() : undefined,
                },
            },
        },
    ]
}

describe('SearchPage', () => {
    afterAll(cleanup)

    beforeEach(() => {
        // eslint-disable-next-line @typescript-eslint/consistent-type-assertions
        window.context = {} as SourcegraphContext & Mocha.SuiteFunction
    })

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
        extensionsController,
        telemetryService: NOOP_TELEMETRY_SERVICE,
        themePreference: ThemePreference.Light,
        onThemePreferenceChange: () => undefined,
        authenticatedUser: authUser,
        globbing: false,
        platformContext: {} as any,
        keyboardShortcuts: [],
        searchContextsEnabled: true,
        selectedSearchContextSpec: '',
        setSelectedSearchContextSpec: () => {},
        defaultSearchContextSpec: '',
        isLightTheme: true,
        fetchAutoDefinedSearchContexts: mockFetchAutoDefinedSearchContexts(),
        fetchSearchContexts: mockFetchSearchContexts,
        hasUserAddedRepositories: false,
        hasUserAddedExternalServices: false,
        getUserSearchContextNamespaces: mockGetUserSearchContextNamespaces,
    }

    it('should not show home panels if on Sourcegraph.com and showEnterpriseHomePanels disabled', () => {
        container = renderWithBrandedContext(
            <MockedTestProvider
                mocks={getMocks({
                    enableSavedSearches: false,
                    enableCollaborators: false,
                })}
            >
                <SearchPage {...defaultProps} isSourcegraphDotCom={true} />
            </MockedTestProvider>
        ).container
        const homePanels = container.querySelector('[data-testid="home-panels"]')
        expect(homePanels).not.toBeInTheDocument()
    })

    it('should show home panels if on Sourcegraph.com and showEnterpriseHomePanels enabled', () => {
        useExperimentalFeatures.setState({ showEnterpriseHomePanels: true })

        container = renderWithBrandedContext(
            <MockedTestProvider
                mocks={getMocks({
                    enableSavedSearches: false,
                    enableCollaborators: false,
                })}
            >
                <SearchPage {...defaultProps} isSourcegraphDotCom={true} />
            </MockedTestProvider>
        ).container
        const homePanels = container.querySelector('[data-testid="home-panels"]')
        expect(homePanels).toBeVisible()
    })

    it('should show home panels if on Sourcegraph.com and showEnterpriseHomePanels enabled with user logged out', () => {
        useExperimentalFeatures.setState({ showEnterpriseHomePanels: true })

        container = renderWithBrandedContext(
            <MockedTestProvider
                mocks={getMocks({
                    enableSavedSearches: false,
                    enableCollaborators: false,
                })}
            >
                <SearchPage {...defaultProps} isSourcegraphDotCom={true} authenticatedUser={null} />
            </MockedTestProvider>
        ).container
        const homePanels = container.querySelector('[data-testid="home-panels"]')
        expect(homePanels).not.toBeInTheDocument()
    })

    it('should not show home panels if showEnterpriseHomePanels disabled', () => {
        container = renderWithBrandedContext(
            <MockedTestProvider
                mocks={getMocks({
                    enableSavedSearches: false,
                    enableCollaborators: false,
                })}
            >
                <SearchPage {...defaultProps} />
            </MockedTestProvider>
        ).container
        const homePanels = container.querySelector('[data-testid="home-panels"]')
        expect(homePanels).not.toBeInTheDocument()
    })

    it('should show home panels if showEnterpriseHomePanels enabled and not on Sourcegraph.com', () => {
        useExperimentalFeatures.setState({ showEnterpriseHomePanels: true })

        container = renderWithBrandedContext(
            <MockedTestProvider
                mocks={getMocks({
                    enableSavedSearches: false,
                    enableCollaborators: false,
                })}
            >
                <SearchPage {...defaultProps} />
            </MockedTestProvider>
        ).container
        const homePanels = container.querySelector('[data-testid="home-panels"]')
        expect(homePanels).toBeVisible()
    })
})
