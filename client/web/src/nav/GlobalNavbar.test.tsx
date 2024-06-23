import React from 'react'

import { afterAll, describe, expect, test, vi } from 'vitest'

import { MockedTestProvider } from '@sourcegraph/shared/src/testing/apollo'
import {
    mockFetchSearchContexts,
    mockGetUserSearchContextNamespaces,
} from '@sourcegraph/shared/src/testing/searchContexts/testHelpers'
import { NOOP_SETTINGS_CASCADE } from '@sourcegraph/shared/src/testing/searchTestHelpers'
import { renderWithBrandedContext } from '@sourcegraph/wildcard/src/testing'

import { GlobalNavbar } from './GlobalNavbar'

vi.mock('../search/input/SearchNavbarItem', () => ({ SearchNavbarItem: () => 'SearchNavbarItem' }))
vi.mock('../components/branding/BrandLogo', () => ({ BrandLogo: () => 'BrandLogo' }))

const PROPS: React.ComponentProps<typeof GlobalNavbar> = {
    authenticatedUser: null,
    isSourcegraphDotCom: false,
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
    if (!window.context) {
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        window.context = {} as any
    }
    const origCodeSearchEnabledOnInstance = window.context?.codeSearchEnabledOnInstance
    const origCodyEnabledOnInstance = window.context?.codyEnabledOnInstance
    const origCodyEnabledForCurrentUser = window.context?.codyEnabledForCurrentUser
    afterAll(() => {
        vi.restoreAllMocks()
        window.context.codeSearchEnabledOnInstance = origCodeSearchEnabledOnInstance
        window.context.codyEnabledOnInstance = origCodyEnabledOnInstance
        window.context.codyEnabledForCurrentUser = origCodyEnabledForCurrentUser
    })

    test('default', () => {
        window.context.codeSearchEnabledOnInstance = true
        window.context.codyEnabledOnInstance = true
        window.context.codyEnabledForCurrentUser = true

        const { asFragment } = renderWithBrandedContext(
            <MockedTestProvider>
                <GlobalNavbar {...PROPS} />
            </MockedTestProvider>
        )
        expect(asFragment()).toMatchSnapshot()
    })

    test('cody only license', () => {
        window.context.codeSearchEnabledOnInstance = false
        window.context.codyEnabledOnInstance = true
        window.context.codyEnabledForCurrentUser = true

        const { asFragment } = renderWithBrandedContext(
            <MockedTestProvider>
                <GlobalNavbar {...PROPS} />
            </MockedTestProvider>
        )
        expect(asFragment()).toMatchSnapshot()
    })

    test('code search only license', () => {
        window.context.codeSearchEnabledOnInstance = true
        window.context.codyEnabledOnInstance = false
        window.context.codyEnabledForCurrentUser = false
        vi.mock('../util/features', () => ({
            isOnlyCodyEnabledOnInstance: () => false,
        }))

        const { asFragment } = renderWithBrandedContext(
            <MockedTestProvider>
                <GlobalNavbar {...PROPS} />
            </MockedTestProvider>
        )
        expect(asFragment()).toMatchSnapshot()
    })
})
