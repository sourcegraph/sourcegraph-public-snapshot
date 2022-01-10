import { render } from '@testing-library/react'
import { createLocation, createMemoryHistory } from 'history'
import React from 'react'
import { MemoryRouter } from 'react-router'

import { setLinkComponent } from '@sourcegraph/shared/src/components/Link'
import {
    mockFetchAutoDefinedSearchContexts,
    mockFetchSearchContexts,
    mockGetUserSearchContextNamespaces,
} from '@sourcegraph/shared/src/testing/searchContexts/testHelpers'
import { extensionsController, NOOP_SETTINGS_CASCADE } from '@sourcegraph/shared/src/util/searchTestHelpers'

import { useExperimentalFeatures, useNavbarQueryState } from '../stores'
import { ThemePreference } from '../stores/themeState'

import { GlobalNavbar } from './GlobalNavbar'

jest.mock('../search/input/SearchNavbarItem', () => ({ SearchNavbarItem: 'SearchNavbarItem' }))
jest.mock('../components/branding/BrandLogo', () => ({ BrandLogo: 'BrandLogo' }))

const PROPS: React.ComponentProps<typeof GlobalNavbar> = {
    authenticatedUser: null,
    authRequired: false,
    extensionsController,
    location: createLocation('/'),
    history: createMemoryHistory(),
    keyboardShortcuts: [],
    isSourcegraphDotCom: false,
    onThemePreferenceChange: () => undefined,
    isLightTheme: true,
    themePreference: ThemePreference.Light,
    parsedSearchQuery: 'r:golang/oauth2 test f:travis',
    platformContext: {} as any,
    settingsCascade: NOOP_SETTINGS_CASCADE,
    batchChangesEnabled: false,
    batchChangesExecutionEnabled: false,
    batchChangesWebhookLogsEnabled: false,
    telemetryService: {} as any,
    isExtensionAlertAnimating: false,
    showSearchBox: true,
    selectedSearchContextSpec: '',
    setSelectedSearchContextSpec: () => undefined,
    defaultSearchContextSpec: '',
    variant: 'default',
    globbing: false,
    branding: undefined,
    routes: [],
    searchContextsEnabled: true,
    fetchAutoDefinedSearchContexts: mockFetchAutoDefinedSearchContexts(),
    fetchSearchContexts: mockFetchSearchContexts,
    hasUserAddedRepositories: false,
    hasUserAddedExternalServices: false,
    getUserSearchContextNamespaces: mockGetUserSearchContextNamespaces,
    extensionViews: () => null,
}

describe('GlobalNavbar', () => {
    setLinkComponent(({ children, ...props }) => <a {...props}>{children}</a>)
    afterAll(() => setLinkComponent(() => null)) // reset global env for other tests
    beforeEach(() => {
        useNavbarQueryState.setState({ searchCaseSensitivity: false })
        useExperimentalFeatures.setState({ codeMonitoring: false })
    })

    test('default', () => {
        const { asFragment } = render(
            <MemoryRouter>
                <GlobalNavbar {...PROPS} />
            </MemoryRouter>
        )
        expect(asFragment()).toMatchSnapshot()
    })

    test('low-profile', () => {
        const { asFragment } = render(
            <MemoryRouter>
                <GlobalNavbar {...PROPS} variant="low-profile" />
            </MemoryRouter>
        )
        expect(asFragment()).toMatchSnapshot()
    })
})
