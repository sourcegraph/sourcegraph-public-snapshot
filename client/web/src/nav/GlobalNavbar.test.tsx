import React from 'react'

import { MockedTestProvider } from '@sourcegraph/shared/src/testing/apollo'
import {
    mockFetchSearchContexts,
    mockGetUserSearchContextNamespaces,
} from '@sourcegraph/shared/src/testing/searchContexts/testHelpers'
import { NOOP_SETTINGS_CASCADE } from '@sourcegraph/shared/src/testing/searchTestHelpers'
import { renderWithBrandedContext } from '@sourcegraph/wildcard/src/testing'

import { GlobalNavbar } from './GlobalNavbar'

jest.mock('../search/input/SearchNavbarItem', () => ({ SearchNavbarItem: 'SearchNavbarItem' }))
jest.mock('../components/branding/BrandLogo', () => ({ BrandLogo: 'BrandLogo' }))

const PROPS: React.ComponentProps<typeof GlobalNavbar> = {
    authenticatedUser: null,
    isSourcegraphDotCom: false,
    isSourcegraphApp: false,
    platformContext: {} as any,
    settingsCascade: NOOP_SETTINGS_CASCADE,
    batchChangesEnabled: false,
    batchChangesExecutionEnabled: false,
    batchChangesWebhookLogsEnabled: false,
    telemetryService: {} as any,
    showSearchBox: true,
    selectedSearchContextSpec: '',
    setSelectedSearchContextSpec: () => undefined,
    branding: undefined,
    routes: [],
    searchContextsEnabled: true,
    fetchSearchContexts: mockFetchSearchContexts,
    getUserSearchContextNamespaces: mockGetUserSearchContextNamespaces,
    showKeyboardShortcutsHelp: () => undefined,
    setFuzzyFinderIsVisible: () => undefined,
    notebooksEnabled: true,
    codeMonitoringEnabled: true,
    ownEnabled: true,
    showFeedbackModal: () => undefined,
}

describe('GlobalNavbar', () => {
    describe('default', () => {
        test('anonymous', () => {
            const { asFragment } = renderWithBrandedContext(
                <MockedTestProvider>
                    <GlobalNavbar {...PROPS} />
                </MockedTestProvider>
            )
            expect(asFragment()).toMatchSnapshot()
        })

        test('signed in', () => {
            const { asFragment } = renderWithBrandedContext(
                <MockedTestProvider>
                    <GlobalNavbar {...PROPS} />
                </MockedTestProvider>
            )
            expect(asFragment()).toMatchSnapshot()
        })
    })

    describe('dotcom', () => {
        test('anonymous', () => {
            const { asFragment } = renderWithBrandedContext(
                <MockedTestProvider>
                    <GlobalNavbar {...PROPS} />
                </MockedTestProvider>
            )
            expect(asFragment()).toMatchSnapshot()
        })

        test('signed in', () => {
            const { asFragment } = renderWithBrandedContext(
                <MockedTestProvider>
                    <GlobalNavbar {...PROPS} />
                </MockedTestProvider>
            )
            expect(asFragment()).toMatchSnapshot()
        })
    })

    describe('app', () => {
        // There is no anonymous mode for app that shows the global navbar.

        test('signed in', () => {
            const { asFragment } = renderWithBrandedContext(
                <MockedTestProvider>
                    <GlobalNavbar {...PROPS} />
                </MockedTestProvider>
            )
            expect(asFragment()).toMatchSnapshot()
        })
    })
})
