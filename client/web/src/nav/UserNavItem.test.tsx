import { render } from '@testing-library/react'
import * as H from 'history'
import React from 'react'
import { MemoryRouter } from 'react-router'

import { ThemePreference } from '../stores/themeState'

import { UserNavItem, UserNavItemProps } from './UserNavItem'

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

jest.mock('../auth', () => {
    // eslint-disable-next-line @typescript-eslint/no-unsafe-assignment, @typescript-eslint/no-require-imports, @typescript-eslint/no-var-requires
    const Rxjs = require('rxjs')
    return {
        authenticatedUser: Rxjs.of(USER),
    }
})

describe('UserNavItem', () => {
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
})
