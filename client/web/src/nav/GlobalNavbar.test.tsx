import React from 'react'

import { afterEach, beforeEach, describe, expect, test, vi } from 'vitest'

import type { AuthenticatedUser } from '@sourcegraph/shared/src/auth'
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
    authenticatedUser: { username: 'alice', organizations: { nodes: [] } } as Partial<AuthenticatedUser> as any,
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
    searchJobsEnabled: true,
    showFeedbackModal: () => undefined,
}

describe('GlobalNavbar', () => {
    if (!window.context) {
        window.context = {} as any
    }
    const origCodeSearchEnabledOnInstance = window.context?.codeSearchEnabledOnInstance ?? true
    const origCodyEnabledOnInstance = window.context?.codyEnabledOnInstance ?? true
    const origCodyEnabledForCurrentUser = window.context?.codyEnabledForCurrentUser ?? true
    const reset = () => {
        window.context.codeSearchEnabledOnInstance = origCodeSearchEnabledOnInstance
        window.context.codyEnabledOnInstance = origCodyEnabledOnInstance
        window.context.codyEnabledForCurrentUser = origCodyEnabledForCurrentUser
    }
    beforeEach(reset)
    afterEach(reset)

    test('default', () => {
        window.context.codeSearchEnabledOnInstance = true
        window.context.codyEnabledOnInstance = true
        window.context.codyEnabledForCurrentUser = true

        const { baseElement } = renderWithBrandedContext(
            <MockedTestProvider>
                <GlobalNavbar {...PROPS} />
            </MockedTestProvider>
        )
        expect(describeNavBar(baseElement)).toEqual<NavBarTestDescription>({
            codyItemType: 'link',
            codyItemLink: 'Cody /cody/chat',
        })
    })

    test('dotcom unauthed', () => {
        window.context.codyEnabledForCurrentUser = false
        const { baseElement } = renderWithBrandedContext(
            <MockedTestProvider>
                <GlobalNavbar {...PROPS} isSourcegraphDotCom={true} authenticatedUser={null} />
            </MockedTestProvider>
        )
        expect(describeNavBar(baseElement)).toEqual<NavBarTestDescription>({
            codyItemType: 'link',
            codyItemLink: 'Cody https://sourcegraph.com/cody',
        })
    })

    test('dotcom authed', () => {
        const { baseElement } = renderWithBrandedContext(
            <MockedTestProvider>
                <GlobalNavbar {...PROPS} isSourcegraphDotCom={true} />
            </MockedTestProvider>
        )
        expect(describeNavBar(baseElement)).toEqual<NavBarTestDescription>({
            codyItemType: 'link',
            codyItemLink: 'Cody /cody/chat',
        })
    })

    test('enterprise cody enabled for user', () => {
        window.context.codyEnabledForCurrentUser = true
        const { baseElement } = renderWithBrandedContext(
            <MockedTestProvider>
                <GlobalNavbar {...PROPS} />
            </MockedTestProvider>
        )
        expect(describeNavBar(baseElement)).toEqual<NavBarTestDescription>({
            codyItemType: 'link',
            codyItemLink: 'Cody /cody/chat',
        })
    })

    test('enterprise cody disabled for user', () => {
        window.context.codyEnabledForCurrentUser = false
        const { baseElement } = renderWithBrandedContext(
            <MockedTestProvider>
                <GlobalNavbar {...PROPS} />
            </MockedTestProvider>
        )
        expect(describeNavBar(baseElement)).toEqual<NavBarTestDescription>({
            codyItemType: 'link',
            codyItemLink: 'Cody /cody/dashboard',
        })
    })

    test('code search disabled on instance', () => {
        window.context.codeSearchEnabledOnInstance = false
        window.context.codyEnabledOnInstance = true
        window.context.codyEnabledForCurrentUser = true

        const { baseElement } = renderWithBrandedContext(
            <MockedTestProvider>
                <GlobalNavbar {...PROPS} />
            </MockedTestProvider>
        )
        expect(describeNavBar(baseElement)).toEqual<NavBarTestDescription>({
            codyItemType: 'link',
            codyItemLink: 'BrandLogo /cody/chat',
        })
    })

    test('cody disabled on instance', () => {
        window.context.codeSearchEnabledOnInstance = true
        window.context.codyEnabledOnInstance = false
        window.context.codyEnabledForCurrentUser = false

        const { baseElement } = renderWithBrandedContext(
            <MockedTestProvider>
                <GlobalNavbar {...PROPS} />
            </MockedTestProvider>
        )
        expect(baseElement.querySelector('a[href*="cody"]')).toBeNull()
        expect(describeNavBar(baseElement)).toEqual<NavBarTestDescription>({ codyItemType: 'none' })
    })
})

interface NavBarTestDescription {
    codyItemType: 'none' | 'link'
    codyItemLink?: string
}

function describeNavBar(baseElement: HTMLElement): NavBarTestDescription {
    const item = baseElement.querySelector<HTMLAnchorElement>('a[href*="cody"]')
    return item
        ? {
              codyItemType: 'link',
              codyItemLink: `${item.textContent} ${item.getAttribute('href') ?? ''}`,
          }
        : { codyItemType: 'none' }
}
