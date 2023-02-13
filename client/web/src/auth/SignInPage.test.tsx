import { within } from '@testing-library/dom'
import { Route, Routes } from 'react-router-dom-v5-compat'

import { renderWithBrandedContext } from '@sourcegraph/wildcard/src/testing'

import { AuthenticatedUser } from '../auth'
import { AuthProvider, SourcegraphContext } from '../jscontext'

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

    const render = (
        route: string,
        props: {
            authProviders?: AuthProvider[]
            authenticatedUser?: AuthenticatedUser
            sourcegraphDotComMode?: boolean
        }
    ) =>
        renderWithBrandedContext(
            <Routes>
                <Route
                    path="/sign-in"
                    element={
                        <SignInPage
                            authenticatedUser={props.authenticatedUser ?? null}
                            context={{
                                allowSignup: true,
                                sourcegraphDotComMode: props.sourcegraphDotComMode ?? false,
                                authProviders: props.authProviders ?? authProviders,
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

    it('renders sign in page (server)', () => {
        const rendered = render('/sign-in', {})
        expect(
            within(rendered.baseElement)
                .queryByText(txt => txt.includes('Continue with Email'))
                ?.closest('a')
        ).toHaveAttribute('href', '/sign-in?email=1&')

        expect(rendered.asFragment()).toMatchSnapshot()
    })

    it('renders sign in page (server) with email form expanded', () => {
        const rendered = render('/sign-in?email=1', {})
        expect(
            within(rendered.baseElement).queryByText(txt => txt.includes('Continue with Email'))
        ).not.toBeInTheDocument()
        expect(rendered.asFragment()).toMatchSnapshot()
    })

    it('renders sign in page (server) with only builtin authProvider', () => {
        const rendered = render('/sign-in', {
            authProviders: authProviders.filter(authProvider => authProvider.serviceType === 'builtin'),
        })
        expect(
            within(rendered.baseElement).queryByText(txt => txt.includes('Continue with Email'))
        ).not.toBeInTheDocument()

        expect(rendered.asFragment()).toMatchSnapshot()
    })

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
            const rendered = render('/sign-in', { authProviders: withSourcegraphOperator })
            expect(
                within(rendered.baseElement).queryByText(txt => txt.includes('Sourcegraph Operators'))
            ).not.toBeInTheDocument()
            expect(rendered.asFragment()).toMatchSnapshot()
        })

        it('renders page with 3 providers (url-param present)', () => {
            const rendered = render('/sign-in?sourcegraph-operator', { authProviders: withSourcegraphOperator })
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
            const rendered = render('/sign-in', { authProviders: withGerritProvider })
            expect(within(rendered.baseElement).queryByText(txt => txt.includes('Gerrit'))).not.toBeInTheDocument()
            expect(rendered.asFragment()).toMatchSnapshot()
        })
    })

    it('renders sign in page (cloud)', () => {
        expect(render('/sign-in', { sourcegraphDotComMode: true }).asFragment()).toMatchSnapshot()
    })

    it('renders redirect when user is authenticated', () => {
        // eslint-disable-next-line @typescript-eslint/consistent-type-assertions
        const mockUser = {
            id: 'userID',
            username: 'username',
            emails: [{ email: 'user@me.com', isPrimary: true, verified: true }],
            siteAdmin: true,
        } as AuthenticatedUser

        expect(render('/sign-in', { authenticatedUser: mockUser }).asFragment()).toMatchSnapshot()
    })
})
