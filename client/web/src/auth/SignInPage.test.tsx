import { within } from '@testing-library/dom'
import { Route, Routes } from 'react-router-dom-v5-compat'

import { renderWithBrandedContext } from '@sourcegraph/wildcard'

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
            serviceID: '',
        },
        {
            serviceType: 'github',
            displayName: 'GitHub',
            isBuiltin: false,
            authenticationURL: '/.auth/github/login?pc=f00bar',
            serviceID: 'https://github.com',
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
                                    experimentalFeatures: {},
                                }}
                                isSourcegraphDotCom={false}
                            />
                        }
                    />
                </Routes>,
                { route: '/sign-in' }
            ).asFragment()
        ).toMatchSnapshot()
    })

    describe('with Sourcegraph auth provider', () => {
        it('renders page with 3 providers (experimentalFeature disabled)', () => {
            const rendered = render(false, '/sign-in')

            expect(
                within(rendered.baseElement).queryByText(txt => txt.includes('Sourcegraph Employee'))
            ).toBeInTheDocument()
            expect(rendered.asFragment()).toMatchSnapshot()
        })

        it('renders page with 2 providers (experimentalFeature enabled)', () => {
            const rendered = render(true, '/sign-in')
            expect(
                within(rendered.baseElement).queryByText(txt => txt.includes('Sourcegraph Employee'))
            ).not.toBeInTheDocument()
            expect(rendered.asFragment()).toMatchSnapshot()
        })

        it('renders page with 3 providers (experimentalFeature enabled & url-param present)', () => {
            const rendered = render(true, '/sign-in?sourcegraph-operator')
            expect(
                within(rendered.baseElement).queryByText(txt => txt.includes('Sourcegraph Employee'))
            ).toBeInTheDocument()
            expect(rendered.asFragment()).toMatchSnapshot()
        })

        function render(hideSourcegraphOperatorLogin: boolean, route: string) {
            const withSourcegraphOperator: SourcegraphContext['authProviders'] = [
                ...authProviders,
                {
                    displayName: 'Sourcegraph Employee',
                    isBuiltin: false,
                    serviceType: 'openidconnect',
                    authenticationURL: '',
                    serviceID: '',
                },
            ]

            return renderWithBrandedContext(
                <Routes>
                    <Route
                        path="/sign-in"
                        element={
                            <SignInPage
                                authenticatedUser={null}
                                context={{
                                    allowSignup: true,
                                    sourcegraphDotComMode: false,
                                    authProviders: withSourcegraphOperator,
                                    resetPasswordEnabled: true,
                                    xhrHeaders: {},
                                    experimentalFeatures: { hideSourcegraphOperatorLogin },
                                }}
                                isSourcegraphDotCom={false}
                            />
                        }
                    />
                </Routes>,
                { route }
            )
        }
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
                                    experimentalFeatures: {},
                                }}
                                isSourcegraphDotCom={false}
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
                                    experimentalFeatures: {},
                                }}
                                isSourcegraphDotCom={false}
                            />
                        }
                    />
                </Routes>,
                { route: '/sign-in' }
            ).asFragment()
        ).toMatchSnapshot()
    })
})
