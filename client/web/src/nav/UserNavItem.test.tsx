import { render } from '@testing-library/react'
import * as H from 'history'
import React from 'react'
import { MemoryRouter } from 'react-router'
import sinon from 'sinon'

import { renderWithRouter } from '@sourcegraph/shared/src/testing/render-with-router'
import { AnchorLink, RouterLink, setLinkComponent } from '@sourcegraph/wildcard'

import { ThemePreference } from '../stores/themeState'

import { UserNavItem, UserNavItemProps } from './UserNavItem'

describe('UserNavItem', () => {
    beforeAll(() => {
        setLinkComponent(RouterLink)
    })

    afterAll(() => {
        setLinkComponent(AnchorLink)
    })

    const USER: UserNavItemProps['authenticatedUser'] = {
        username: 'alice',
        displayName: 'alice doe',
        avatarURL: null,
        session: { canSignOut: true },
        settingsURL: '#',
        siteAdmin: true,
        organizations: {
            nodes: [
                {
                    id: '0',
                    name: 'acme',
                    displayName: 'Acme Corp',
                    url: '/organizations/acme',
                    settingsURL: '/organizations/acme/settings',
                },
                {
                    id: '1',
                    name: 'beta',
                    displayName: 'Beta Inc',
                    url: '/organizations/beta',
                    settingsURL: '/organizations/beta/settings',
                },
            ],
        },
    }

    const history = H.createMemoryHistory({ keyLength: 0 })

    test('simple', () => {
        expect(
            render(
                <MemoryRouter>
                    <UserNavItem
                        showRepositorySection={true}
                        isLightTheme={true}
                        onThemePreferenceChange={() => undefined}
                        themePreference={ThemePreference.Light}
                        location={history.location}
                        authenticatedUser={USER}
                        showDotComMarketing={true}
                        isExtensionAlertAnimating={false}
                        codeHostIntegrationMessaging="browser-extension"
                    />
                </MemoryRouter>
            ).asFragment()
        ).toMatchSnapshot()
    })

    test('logout click triggers page refresh instead of performing client-side only navigation', async () => {
        const renderResult = renderWithRouter(
            <UserNavItem
                showRepositorySection={true}
                isLightTheme={true}
                onThemePreferenceChange={() => undefined}
                themePreference={ThemePreference.Light}
                location={history.location}
                authenticatedUser={USER}
                showDotComMarketing={true}
                isExtensionAlertAnimating={false}
                codeHostIntegrationMessaging="browser-extension"
            />,
            {
                history,
            }
        )

        // Prevent console.error cause by "Not implemented: navigation (except hash changes)"
        // https://github.com/jsdom/jsdom/issues/2112
        sinon.stub(console, 'error')
        const singOutLink = await renderResult.findByText('Sign out')
        singOutLink.click()

        expect(history.entries.length).toBe(1)
        expect(history.entries.find(({ pathname }) => pathname.includes('sign-out'))).toBe(undefined)
    })
})
