import React from 'react'

import { createLocation, createMemoryHistory } from 'history'

import { renderWithBrandedContext } from '@sourcegraph/shared/src/testing'
import { MockedTestProvider } from '@sourcegraph/shared/src/testing/apollo'
import {
    mockFetchAutoDefinedSearchContexts,
    mockFetchSearchContexts,
    mockGetUserSearchContextNamespaces,
} from '@sourcegraph/shared/src/testing/searchContexts/testHelpers'
import { extensionsController, NOOP_SETTINGS_CASCADE } from '@sourcegraph/shared/src/testing/searchTestHelpers'

import { useExperimentalFeatures } from '../stores'
import { ThemePreference } from '../stores/themeState'

import { GlobalNavbar } from './GlobalNavbar'

jest.mock('../search/input/SearchNavbarItem', () => ({ SearchNavbarItem: 'SearchNavbarItem' }))
jest.mock('../components/branding/BrandLogo', () => ({ BrandLogo: 'BrandLogo' }))

const history = createMemoryHistory()
const PROPS: React.ComponentProps<typeof GlobalNavbar> = {
    authenticatedUser: null,
    authRequired: false,
    extensionsController,
    location: createLocation('/'),
    history,
    keyboardShortcuts: [],
    isSourcegraphDotCom: false,
    onThemePreferenceChange: () => undefined,
    isLightTheme: true,
    themePreference: ThemePreference.Light,
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
}

describe('GlobalNavbar', () => {
    beforeEach(() => {
        useExperimentalFeatures.setState({ codeMonitoring: false, showSearchContext: true })
    })

    test('default', () => {
        const { asFragment } = renderWithBrandedContext(
            <MockedTestProvider>
                <GlobalNavbar {...PROPS} />
            </MockedTestProvider>
        )
        expect(asFragment()).toMatchSnapshot()
    })

    test('low-profile', () => {
        const { asFragment } = renderWithBrandedContext(
            <MockedTestProvider>
                <GlobalNavbar {...PROPS} variant="low-profile" />
            </MockedTestProvider>
        )
        expect(asFragment()).toMatchSnapshot()
    })
})
