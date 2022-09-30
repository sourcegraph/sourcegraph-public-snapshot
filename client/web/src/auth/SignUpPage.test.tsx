import { Route, Routes } from 'react-router-dom-v5-compat'

import { NOOP_TELEMETRY_SERVICE } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { renderWithBrandedContext } from '@sourcegraph/shared/src/testing'
import { MockedTestProvider } from '@sourcegraph/shared/src/testing/apollo'

import { AuthenticatedUser } from '../auth'
import { SourcegraphContext } from '../jscontext'

import { SignUpPage } from './SignUpPage'

describe('SignUpPage', () => {
    const commonProps = {
        isLightTheme: true,
    }
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

    it('renders sign up page (server)', () => {
        expect(
            renderWithBrandedContext(
                <MockedTestProvider mocks={[]}>
                    <Routes>
                        <Route
                            path="/sign-up"
                            element={
                                <SignUpPage
                                    {...commonProps}
                                    authenticatedUser={null}
                                    context={{
                                        allowSignup: true,
                                        sourcegraphDotComMode: false,
                                        experimentalFeatures: { enablePostSignupFlow: false },
                                        authMinPasswordLength: 12,
                                        authProviders,
                                        xhrHeaders: {},
                                    }}
                                    telemetryService={NOOP_TELEMETRY_SERVICE}
                                />
                            }
                        />
                    </Routes>
                </MockedTestProvider>,
                { route: '/sign-up' }
            ).asFragment()
        ).toMatchSnapshot()
    })

    it('renders sign up page (cloud)', () => {
        expect(
            renderWithBrandedContext(
                <MockedTestProvider mocks={[]}>
                    <Routes>
                        <Route
                            path="/sign-up"
                            element={
                                <SignUpPage
                                    {...commonProps}
                                    authenticatedUser={null}
                                    context={{
                                        allowSignup: true,
                                        sourcegraphDotComMode: true,
                                        experimentalFeatures: { enablePostSignupFlow: false },
                                        authMinPasswordLength: 12,
                                        authProviders,
                                        xhrHeaders: {},
                                    }}
                                    telemetryService={NOOP_TELEMETRY_SERVICE}
                                />
                            }
                        />
                    </Routes>
                </MockedTestProvider>,
                { route: '/sign-up' }
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
                <MockedTestProvider mocks={[]}>
                    <Routes>
                        <Route
                            path="/sign-up"
                            element={
                                <SignUpPage
                                    {...commonProps}
                                    authenticatedUser={mockUser}
                                    context={{
                                        allowSignup: true,
                                        sourcegraphDotComMode: false,
                                        experimentalFeatures: { enablePostSignupFlow: false },
                                        authMinPasswordLength: 12,
                                        authProviders,
                                        xhrHeaders: {},
                                    }}
                                    telemetryService={NOOP_TELEMETRY_SERVICE}
                                />
                            }
                        />
                    </Routes>
                </MockedTestProvider>,
                { route: '/sign-up' }
            ).asFragment()
        ).toMatchSnapshot()
    })
})
