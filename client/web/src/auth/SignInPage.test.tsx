import { within } from '@testing-library/dom'
import { Route, Routes } from 'react-router-dom-v5-compat'

import { renderWithBrandedContext } from '@sourcegraph/wildcard/src/testing'

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

    const render = (route: string, authProviders: SourcegraphContext['authProviders']) =>
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
                            isSourcegraphDotCom={false}
                        />
                    }
                />
            </Routes>,
            { route }
        )

    describe('with Sourcegraph operator auth provider', () => {
        const withSourcegraphOperator: SourcegraphContext['authProviders'] = [
            ...authProviders,
            {
                displayName: 'Sourcegraph Operators',
                isBuiltin: false,
                serviceType: 'sourcegraph-operator',
                authenticationURL: '',
                serviceID: '',
            },
        ]

        it('renders page with 2 providers', () => {
            const rendered = render('/sign-in', withSourcegraphOperator)
            expect(
                within(rendered.baseElement).queryByText(txt => txt.includes('Sourcegraph Operators'))
            ).not.toBeInTheDocument()
            expect(rendered.asFragment()).toMatchSnapshot()
        })

        it('renders page with 3 providers (url-param present)', () => {
            const rendered = render('/sign-in?sourcegraph-operator', withSourcegraphOperator)
            expect(
                within(rendered.baseElement).queryByText(txt => txt.includes('Sourcegraph Operators'))
            ).toBeInTheDocument()
            expect(rendered.asFragment()).toMatchSnapshot()
        })
    })

    describe('with Gerrit auth provider', () => {
        const withGerritProvider: SourcegraphContext['authProviders'] = [
            ...authProviders,
            {
                displayName: 'Gerrit',
                isBuiltin: false,
                serviceType: 'gerrit',
                authenticationURL: '',
                serviceID: '',
            },
        ]
        it('does not render the Gerrit provider', () => {
            const rendered = render('/sign-in', withGerritProvider)
            expect(within(rendered.baseElement).queryByText(txt => txt.includes('Gerrit'))).not.toBeInTheDocument()
            expect(rendered.asFragment()).toMatchSnapshot()
        })
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
            emails: [{ email: 'user@me.com', isPrimary: true, verified: true }],
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
