import type { Meta, StoryFn } from '@storybook/react'

import { noOpTelemetryRecorder } from '@sourcegraph/shared/src/telemetry'

import { WebStory } from '../components/WebStory'
import type { SourcegraphContext } from '../jscontext'

import { SignInPage, type SignInPageProps } from './SignInPage'

const config: Meta = {
    title: 'web/auth/SignInPage',
    parameters: {
        chromatic: { disableSnapshot: false },
    },
}

export default config

const telemetryRecorder: SourcegraphContext['telemetryRecorder'] = noOpTelemetryRecorder

const authProviders: SourcegraphContext['authProviders'] = [
    {
        clientID: '001',
        displayName: 'Builtin username-password authentication',
        isBuiltin: true,
        serviceType: 'builtin',
        authenticationURL: '',
        serviceID: '',
    },
    {
        clientID: '002',
        serviceType: 'github',
        displayName: 'GitHub',
        isBuiltin: false,
        authenticationURL: '/.auth/github/login?pc=f00bar',
        serviceID: 'https://github.com',
    },
    {
        clientID: '003',
        serviceType: 'gitlab',
        displayName: 'GitLab',
        isBuiltin: false,
        authenticationURL: '/.auth/gitlab/login?pc=f00bar',
        serviceID: 'https://gitlab.com',
    },
]

const noBuiltInAuthProviders = authProviders.filter(provider => !provider.isBuiltin)
const onlyBuiltInAuthProvider = authProviders.filter(provider => provider.isBuiltin)

const context: SignInPageProps['context'] = {
    allowSignup: true,
    authProviders,
    sourcegraphDotComMode: false,
    primaryLoginProvidersCount: 5,
    authAccessRequest: { enabled: true },
    xhrHeaders: {},
    resetPasswordEnabled: true,
}

export const Default: StoryFn = () => (
    <WebStory>
        {() => <SignInPage telemetryRecorder={telemetryRecorder} context={context} authenticatedUser={null} />}
    </WebStory>
)

export const ShowMore: StoryFn = () => (
    <WebStory initialEntries={[{ pathname: '/sign-in', search: '?showMore' }]}>
        {() => (
            <SignInPage
                telemetryRecorder={telemetryRecorder}
                context={{ ...context, primaryLoginProvidersCount: 1 }}
                authenticatedUser={null}
            />
        )}
    </WebStory>
)

export const Dotcom: StoryFn = () => (
    <WebStory>
        {() => (
            <SignInPage
                telemetryRecorder={telemetryRecorder}
                context={{ ...context, sourcegraphDotComMode: true }}
                authenticatedUser={null}
            />
        )}
    </WebStory>
)

export const NoProviders: StoryFn = () => (
    <WebStory>
        {() => (
            <SignInPage
                telemetryRecorder={telemetryRecorder}
                context={{ ...context, authProviders: [] }}
                authenticatedUser={null}
            />
        )}
    </WebStory>
)

export const NoBuiltIn: StoryFn = () => (
    <WebStory>
        {() => (
            <SignInPage
                telemetryRecorder={telemetryRecorder}
                context={{ ...context, authProviders: noBuiltInAuthProviders }}
                authenticatedUser={null}
            />
        )}
    </WebStory>
)

export const NoResetPassword: StoryFn = () => (
    <WebStory>
        {() => (
            <SignInPage
                telemetryRecorder={telemetryRecorder}
                context={{ ...context, resetPasswordEnabled: false }}
                authenticatedUser={null}
            />
        )}
    </WebStory>
)

export const NoSignUp: StoryFn = () => (
    <WebStory>
        {() => (
            <SignInPage
                telemetryRecorder={telemetryRecorder}
                context={{ ...context, allowSignup: false }}
                authenticatedUser={null}
            />
        )}
    </WebStory>
)

export const NoAccessRequest: StoryFn = () => (
    <WebStory>
        {() => (
            <SignInPage
                telemetryRecorder={telemetryRecorder}
                context={{ ...context, allowSignup: false, authAccessRequest: { enabled: false } }}
                authenticatedUser={null}
            />
        )}
    </WebStory>
)

export const DotComSignUp: StoryFn = () => (
    <WebStory>
        {() => (
            <SignInPage
                telemetryRecorder={telemetryRecorder}
                context={{ ...context, sourcegraphDotComMode: true }}
                authenticatedUser={null}
            />
        )}
    </WebStory>
)

export const OnlyOnePrimaryProvider: StoryFn = () => (
    <WebStory>
        {() => (
            <SignInPage
                telemetryRecorder={telemetryRecorder}
                context={{ ...context, primaryLoginProvidersCount: 1 }}
                authenticatedUser={null}
            />
        )}
    </WebStory>
)

export const OnlyOnePrimaryProviderWithoutBuiltIn: StoryFn = () => (
    <WebStory>
        {() => (
            <SignInPage
                telemetryRecorder={telemetryRecorder}
                context={{ ...context, primaryLoginProvidersCount: 1, authProviders: noBuiltInAuthProviders }}
                authenticatedUser={null}
            />
        )}
    </WebStory>
)

export const ShowMoreProviders: StoryFn = () => (
    <WebStory initialEntries={['/sign-in?showMore']}>
        {() => (
            <SignInPage
                telemetryRecorder={telemetryRecorder}
                context={{ ...context, primaryLoginProvidersCount: 1 }}
                authenticatedUser={null}
            />
        )}
    </WebStory>
)

export const ShowMoreProvidersWithoutBuiltIn: StoryFn = () => (
    <WebStory initialEntries={['/sign-in?showMore']}>
        {() => (
            <SignInPage
                telemetryRecorder={telemetryRecorder}
                context={{ ...context, authProviders: noBuiltInAuthProviders, primaryLoginProvidersCount: 1 }}
                authenticatedUser={null}
            />
        )}
    </WebStory>
)

export const OnlyBuiltInAuthProvider: StoryFn = () => (
    <WebStory>
        {() => (
            <SignInPage
                telemetryRecorder={telemetryRecorder}
                context={{ ...context, authProviders: onlyBuiltInAuthProvider }}
                authenticatedUser={null}
            />
        )}
    </WebStory>
)

export const PrefixCanBeChanged: StoryFn = () => {
    const providers = noBuiltInAuthProviders.map(provider => ({ ...provider, displayPrefix: 'Just login with' }))

    return (
        <WebStory>
            {() => (
                <SignInPage
                    telemetryRecorder={telemetryRecorder}
                    context={{ ...context, authProviders: providers }}
                    authenticatedUser={null}
                />
            )}
        </WebStory>
    )
}
