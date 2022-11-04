import { Route, Routes } from 'react-router-dom-v5-compat'

import { renderWithBrandedContext } from '@sourcegraph/shared/src/testing'

import { AuthenticatedUser } from '../auth'
import { SourcegraphContext } from '../jscontext'

import { SignInPage } from './SignInPage'

describe('SignInPage', () => {
    const authProviders: SourcegraphContext['authProviders'] = [
        {
            displayName: 'Builtin username-password authentication',
            isBuiltin: true,
            serviceType: 'builtin',
            authenticationURL: '',
        },
        {
            serviceType: 'github',
            displayName: 'GitHub',
            isBuiltin: false,
            authenticationURL: '/.auth/github/login?pc=f00bar',
        },
    ]

    it('renders sign in page (server)', () => {
        expect(
            renderWithBrandedContext(
                <Routes>
                    <Route
                        path="/sign-in"
                        element={
                            <SignInPage
                                authenticatedUser={null}
                                context={{
                                    allowSignup: true,
                                    sourcegraphDotComMode: false,
                                    authProviders,
                                    resetPasswordEnabled: true,
                                    xhrHeaders: {},
                                }}
                            />
                        }
                    />
                </Routes>,
                { route: '/sign-in' }
            ).asFragment()
        ).toMatchSnapshot()
    })

    it('renders sign in page (cloud)', () => {
        expect(
            renderWithBrandedContext(
                <Routes>
                    <Route
                        path="/sign-in"
                        element={
                            <SignInPage
                                authenticatedUser={null}
                                context={{
                                    allowSignup: true,
                                    sourcegraphDotComMode: true,
                                    authProviders,
                                    resetPasswordEnabled: true,
                                    xhrHeaders: {},
                                }}
                            />
                        }
                    />
                </Routes>,
                { route: '/sign-in' }
            ).asFragment()
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
            renderWithBrandedContext(
                <Routes>
                    <Route
                        path="/sign-in"
                        element={
                            <SignInPage
                                authenticatedUser={mockUser}
                                context={{
                                    allowSignup: true,
                                    sourcegraphDotComMode: false,
                                    authProviders,
                                    xhrHeaders: {},
                                    resetPasswordEnabled: true,
                                }}
                            />
                        }
                    />
                </Routes>,
                { route: '/sign-in' }
            ).asFragment()
        ).toMatchSnapshot()
    })
})
