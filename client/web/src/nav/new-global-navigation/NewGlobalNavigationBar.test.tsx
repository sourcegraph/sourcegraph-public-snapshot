import React from 'react'

import { afterEach, beforeEach, describe, expect, test, vi } from 'vitest'

import type { AuthenticatedUser } from '@sourcegraph/shared/src/auth'
import { MockedTestProvider } from '@sourcegraph/shared/src/testing/apollo'
import { renderWithBrandedContext } from '@sourcegraph/wildcard/src/testing'

import { NewGlobalNavigationBar } from './NewGlobalNavigationBar'

vi.mock('../search/input/SearchNavbarItem', () => ({ SearchNavbarItem: () => 'SearchNavbarItem' }))
vi.mock('../components/branding/BrandLogo', () => ({ BrandLogo: () => 'BrandLogo' }))

const PROPS: React.ComponentProps<typeof NewGlobalNavigationBar> = {
    authenticatedUser: { username: 'alice', organizations: { nodes: [] } } as Partial<AuthenticatedUser> as any,
    isSourcegraphDotCom: false,
    notebooksEnabled: false,
    searchContextsEnabled: true,
    codeMonitoringEnabled: true,
    batchChangesEnabled: true,
    codeInsightsEnabled: true,
    searchJobsEnabled: true,
    showSearchBox: true,
    showFeedbackModal: () => {},
    routes: [],
    __testing__initialSideMenuOpen: true,
    telemetryService: {} as any,
    telemetryRecorder: {} as any,
}

describe('NewGlobalNavigationBar', () => {
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
                <NewGlobalNavigationBar {...PROPS} />
            </MockedTestProvider>
        )
        const sidebarElement = baseElement.querySelector<HTMLElement>('[data-reach-dialog-overlay]')!
        expect(describeNavBarSideMenu(sidebarElement)).toEqual<NavBarTestDescription>({
            codyItems: ['Cody /cody/chat'],
        })
    })

    test('dotcom unauthed', () => {
        window.context.codyEnabledForCurrentUser = false
        const { baseElement } = renderWithBrandedContext(
            <MockedTestProvider>
                <NewGlobalNavigationBar {...PROPS} isSourcegraphDotCom={true} authenticatedUser={null} />
            </MockedTestProvider>
        )
        const sidebarElement = baseElement.querySelector<HTMLElement>('[data-reach-dialog-overlay]')!
        expect(describeNavBarSideMenu(sidebarElement)).toEqual<NavBarTestDescription>({
            codyItems: ['Cody https://sourcegraph.com/cody'],
        })
    })

    test('dotcom authed', () => {
        const { baseElement } = renderWithBrandedContext(
            <MockedTestProvider>
                <NewGlobalNavigationBar {...PROPS} isSourcegraphDotCom={true} />
            </MockedTestProvider>
        )
        const sidebarElement = baseElement.querySelector<HTMLElement>('[data-reach-dialog-overlay]')!
        expect(describeNavBarSideMenu(sidebarElement)).toEqual<NavBarTestDescription>({
            codyItems: ['Cody /cody/chat'],
        })
    })

    test('enterprise cody enabled for user', () => {
        window.context.codyEnabledForCurrentUser = true
        const { baseElement } = renderWithBrandedContext(
            <MockedTestProvider>
                <NewGlobalNavigationBar {...PROPS} />
            </MockedTestProvider>
        )
        const sidebarElement = baseElement.querySelector<HTMLElement>('[data-reach-dialog-overlay]')!
        expect(describeNavBarSideMenu(sidebarElement)).toEqual<NavBarTestDescription>({
            codyItems: ['Cody /cody/chat'],
        })
    })

    test('enterprise cody disabled for user', () => {
        window.context.codyEnabledForCurrentUser = false
        const { baseElement } = renderWithBrandedContext(
            <MockedTestProvider>
                <NewGlobalNavigationBar {...PROPS} />
            </MockedTestProvider>
        )
        const sidebarElement = baseElement.querySelector<HTMLElement>('[data-reach-dialog-overlay]')!
        expect(describeNavBarSideMenu(sidebarElement)).toEqual<NavBarTestDescription>({
            codyItems: ['Cody /cody/dashboard'],
        })
    })

    test('code search disabled on instance', () => {
        window.context.codeSearchEnabledOnInstance = false
        window.context.codyEnabledOnInstance = true
        window.context.codyEnabledForCurrentUser = true

        const { baseElement } = renderWithBrandedContext(
            <MockedTestProvider>
                <NewGlobalNavigationBar {...PROPS} />
            </MockedTestProvider>
        )
        const sidebarElement = baseElement.querySelector<HTMLElement>('[data-reach-dialog-overlay]')!
        expect(describeNavBarSideMenu(sidebarElement)).toEqual<NavBarTestDescription>({
            codyItems: ['Cody /cody/chat'],
        })
    })

    test('cody disabled on instance', () => {
        window.context.codeSearchEnabledOnInstance = true
        window.context.codyEnabledOnInstance = false
        window.context.codyEnabledForCurrentUser = false

        const { baseElement } = renderWithBrandedContext(
            <MockedTestProvider>
                <NewGlobalNavigationBar {...PROPS} />
            </MockedTestProvider>
        )
        const sidebarElement = baseElement.querySelector<HTMLElement>('[data-reach-dialog-overlay]')!
        expect(sidebarElement.querySelector('a[href*="cody"]')).toBeNull()
        expect(describeNavBarSideMenu(sidebarElement)).toEqual<NavBarTestDescription>({ codyItems: [] })
    })
})

interface NavBarTestDescription {
    codyItems: string[]
}

function describeNavBarSideMenu(sidebarElement: HTMLElement): NavBarTestDescription {
    return {
        codyItems: Array.from(sidebarElement.querySelectorAll<HTMLAnchorElement>('a[href*="cody"]')).map(
            a => `${a.textContent?.trim() ?? ''} ${a.getAttribute('href') ?? ''}`
        ),
    }
}
