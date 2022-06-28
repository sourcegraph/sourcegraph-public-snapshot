import { createMemoryHistory, createLocation } from 'history'
import { MemoryRouter } from 'react-router'

import { NOOP_TELEMETRY_SERVICE } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { renderWithBrandedContext } from '@sourcegraph/shared/src/testing'
import { MockedTestProvider } from '@sourcegraph/shared/src/testing/apollo'

import { AuthenticatedUser } from '../auth'
import { SourcegraphContext } from '../jscontext'

import { SignUpPage } from './SignUpPage'

describe('SignUpPage', () => {
    const commonProps = {
        history: createMemoryHistory(),
        location: createLocation('/'),
        isLightTheme: true,
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

    it('renders sign up page (server)', () => {
        expect(
            renderWithBrandedContext(
                <MockedTestProvider mocks={[]}>
                    <MemoryRouter>
                        <SignUpPage
                            {...commonProps}
                            authenticatedUser={null}
                            context={{
                                allowSignup: true,
                                sourcegraphDotComMode: false,
                                experimentalFeatures: { enablePostSignupFlow: false },
                                authProviders,
                                xhrHeaders: {},
                            }}
                            telemetryService={NOOP_TELEMETRY_SERVICE}
                        />
                    </MemoryRouter>
                </MockedTestProvider>
            ).asFragment()
        ).toMatchSnapshot()
    })

    it('renders sign up page (cloud)', () => {
        expect(
            renderWithBrandedContext(
                <MockedTestProvider mocks={[]}>
                    <MemoryRouter>
                        <SignUpPage
                            {...commonProps}
                            authenticatedUser={null}
                            context={{
                                allowSignup: true,
                                sourcegraphDotComMode: true,
                                experimentalFeatures: { enablePostSignupFlow: false },
                                authProviders,
                                xhrHeaders: {},
                            }}
                            telemetryService={NOOP_TELEMETRY_SERVICE}
                        />
                    </MemoryRouter>
                </MockedTestProvider>
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
                    <MemoryRouter>
                        <SignUpPage
                            {...commonProps}
                            authenticatedUser={mockUser}
                            context={{
                                allowSignup: true,
                                sourcegraphDotComMode: false,
                                experimentalFeatures: { enablePostSignupFlow: false },
                                authProviders,
                                xhrHeaders: {},
                            }}
                            telemetryService={NOOP_TELEMETRY_SERVICE}
                        />
                    </MemoryRouter>
                </MockedTestProvider>
            ).asFragment()
        ).toMatchSnapshot()
    })
})
