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
    const origContext = window.context
    beforeEach(() => {
        window.context = {} as any
    })
    afterEach(() => {
        window.context = origContext
    })

    const renderPage = (): DocumentFragment => {
        const { asFragment } = renderWithBrandedContext(
            <MockedTestProvider>
                <GlobalNavbar {...PROPS} />
            </MockedTestProvider>
        )
        return asFragment()
    }

    test('no cody, no code search', () => {
        expect(renderPage()).toMatchSnapshot()
    })

    test('with code search', () => {
        window.context = {
            codeSearchEnabled: true,
        } as any
        expect(renderPage()).toMatchSnapshot()
    })

    test('with cody', () => {
        window.context = {
            codyEnabled: true,
        } as any
        expect(renderPage()).toMatchSnapshot()
    })

    test('both code search and cody', () => {
        window.context = {
            codyEnabled: true,
            codeSearchEnabled: true,
        } as any
        expect(renderPage()).toMatchSnapshot()
    })
})
