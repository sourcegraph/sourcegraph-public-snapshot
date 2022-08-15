import { render } from '@testing-library/react'
import { MemoryRouter } from 'react-router'
import { CompatRouter } from 'react-router-dom-v5-compat'

import { PlatformContext } from '@sourcegraph/shared/src/platform/context'

import { AuthenticatedUser } from '../auth'

import { ExtensionCard } from './ExtensionCard'

describe('ExtensionCard', () => {
    const NOOP_PLATFORM_CONTEXT: PlatformContext = {} as any

    const mockUser = {
        id: 'userID',
        username: 'username',
        email: 'user@me.com',
        siteAdmin: true,
    } as AuthenticatedUser

    test('renders', () => {
        expect(
            render(
                <MemoryRouter>
                    <CompatRouter>
                        <ExtensionCard
                            node={{
                                id: 'x/y',
                                manifest: {
                                    activationEvents: ['*'],
                                    description: 'd',
                                    url: 'https://example.com',
                                    icon: 'data:image/png,abcd',
                                },
                                registryExtension: {
                                    id: 'abcd1234',
                                    extensionIDWithoutRegistry: 'x/y',
                                    url: 'extensions/x/y',
                                    isWorkInProgress: false,
                                    viewerCanAdminister: false,
                                },
                            }}
                            subject={{ id: 'u', viewerCanAdminister: false }}
                            viewerSubject={{
                                __typename: 'User',
                                username: 'u',
                                displayName: 'u',
                                id: 'u',
                                viewerCanAdminister: false,
                            }}
                            siteSubject={{
                                __typename: 'Site',
                                id: 's',
                                viewerCanAdminister: true,
                                allowSiteSettingsEdits: true,
                            }}
                            settingsCascade={{ final: null, subjects: null }}
                            platformContext={NOOP_PLATFORM_CONTEXT}
                            enabled={false}
                            enabledForAllUsers={false}
                            isLightTheme={false}
                            settingsURL="/settings/foobar"
                            authenticatedUser={mockUser}
                        />
                    </CompatRouter>
                </MemoryRouter>
            ).asFragment()
        ).toMatchSnapshot()
    })
})
