import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter } from 'react-router-dom'
import sinon from 'sinon'

import { AnchorLink, RouterLink, setLinkComponent } from '@sourcegraph/wildcard'
import { renderWithBrandedContext } from '@sourcegraph/wildcard/src/testing'

import { ThemePreference } from '../theme'

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
                    __typename: 'Org',
                    id: '0',
                    name: 'acme',
                    displayName: 'Acme Corp',
                    url: '/organizations/acme',
                    settingsURL: '/organizations/acme/settings',
                },
                {
                    __typename: 'Org',
                    id: '1',
                    name: 'beta',
                    displayName: 'Beta Inc',
                    url: '/organizations/beta',
                    settingsURL: '/organizations/beta/settings',
                },
            ],
        },
    }

    test('simple', () => {
        expect(
            render(
                <MemoryRouter>
                    <UserNavItem
                        isLightTheme={true}
                        onThemePreferenceChange={() => undefined}
                        showKeyboardShortcutsHelp={() => undefined}
                        themePreference={ThemePreference.Light}
                        authenticatedUser={USER}
                        showDotComMarketing={true}
                        codeHostIntegrationMessaging="browser-extension"
                        showFeedbackModal={() => undefined}
                    />
                </MemoryRouter>
            ).asFragment()
        ).toMatchSnapshot()
    })

    test('logout click triggers page refresh instead of performing client-side only navigation', async () => {
        const result = renderWithBrandedContext(
            <UserNavItem
                isLightTheme={true}
                onThemePreferenceChange={() => undefined}
                showKeyboardShortcutsHelp={() => undefined}
                themePreference={ThemePreference.Light}
                authenticatedUser={USER}
                showDotComMarketing={true}
                codeHostIntegrationMessaging="browser-extension"
                showFeedbackModal={() => undefined}
            />
        )

        // Prevent console.error cause by "Not implemented: navigation (except hash changes)"
        // https://github.com/jsdom/jsdom/issues/2112
        sinon.stub(console, 'error')
        userEvent.click(screen.getByRole('button'))
        userEvent.click(await screen.findByText('Sign out'))

        expect(result.locationRef.entries.length).toBe(1)
        expect(result.locationRef.entries.find(({ pathname }) => pathname.includes('sign-out'))).toBe(undefined)
    })
})
