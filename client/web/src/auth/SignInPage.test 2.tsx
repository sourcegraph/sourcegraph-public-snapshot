import { createMemoryHistory, createLocation } from 'history'
import React from 'react'
import { MemoryRouter } from 'react-router'
import renderer from 'react-test-renderer'

import { AuthenticatedUser } from '../auth'
import { SourcegraphContext } from '../jscontext'

import { SignInPage } from './SignInPage'

describe('SignInPage', () => {
    const commonProps = {
        history: createMemoryHistory(),
        location: createLocation('/'),
    }
    const authProviders: SourcegraphContext['authProviders'] = [
        {
            displayName: 'Builtin username-password authentication',
            isBuiltin: true,
            serviceType: 'builtin',
        },
        {
            serviceType: 'github',
            displayName: 'GitHub',
            isBuiltin: false,
        },
    ]

    it('renders sign in page (server)', () => {
        expect(
            renderer
                .create(
                    <MemoryRouter>
                        <SignInPage
                            {...commonProps}
                            authenticatedUser={null}
                            context={{
                                allowSignup: true,
                                sourcegraphDotComMode: false,
                                authProviders,
                                resetPasswordEnabled: true,
                                xhrHeaders: {},
                            }}
                        />
                    </MemoryRouter>
                )
                .toJSON()
        ).toMatchSnapshot()
    })

    it('renders sign in page (cloud)', () => {
        expect(
            renderer
                .create(
                    <MemoryRouter>
                        <SignInPage
                            {...commonProps}
                            authenticatedUser={null}
                            context={{
                                allowSignup: true,
                                sourcegraphDotComMode: true,
                                authProviders,
                                resetPasswordEnabled: true,
                                xhrHeaders: {},
                            }}
                        />
                    </MemoryRouter>
                )
                .toJSON()
        ).toMatchSnapshot()
    })

    it('renders redirect when user is authenticated', () => {
        // eslint-disable-next-line @typescript-eslint/consistent-type-assertions
        const mockUser = {
            id: 'userID',
            username: 'username',
            email: 'user@me.com',
            siteAdmin: true,
        } as AuthenticatedUser

        expect(
            renderer
                .create(
                    <MemoryRouter>
                        <SignInPage
                            {...commonProps}
                            authenticatedUser={mockUser}
                            context={{
                                allowSignup: true,
                                sourcegraphDotComMode: false,
                                authProviders,
                                xhrHeaders: {},
                                resetPasswordEnabled: true,
                            }}
                        />
                    </MemoryRouter>
                )
                .toJSON()
        ).toMatchSnapshot()
    })
})
