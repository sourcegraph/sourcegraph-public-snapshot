import { Meta, Story } from '@storybook/react'

import { WebStory } from '../components/WebStory'
import { SourcegraphContext } from '../jscontext'

import { SignInPage, SignInPageProps } from './SignInPage'

const config: Meta = {
    title: 'web/auth/SignInPage',
}

export default config

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
    {
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

export const Default: Story = () => (
    <WebStory>{({ isLightTheme }) => <SignInPage context={context} authenticatedUser={null} />}</WebStory>
)

export const NoBuiltIn: Story = () => (
    <WebStory>
        {({ isLightTheme }) => (
            <SignInPage context={{ ...context, authProviders: noBuiltInAuthProviders }} authenticatedUser={null} />
        )}
    </WebStory>
)

export const NoResetPassword: Story = () => (
    <WebStory>
        {({ isLightTheme }) => (
            <SignInPage context={{ ...context, resetPasswordEnabled: false }} authenticatedUser={null} />
        )}
    </WebStory>
)

export const NoSignUp: Story = () => (
    <WebStory>
        {({ isLightTheme }) => <SignInPage context={{ ...context, allowSignup: false }} authenticatedUser={null} />}
    </WebStory>
)

export const NoAccessRequest: Story = () => (
    <WebStory>
        {({ isLightTheme }) => (
            <SignInPage
                context={{ ...context, allowSignup: false, authAccessRequest: { enabled: false } }}
                authenticatedUser={null}
            />
        )}
    </WebStory>
)

export const DotComSignUp: Story = () => (
    <WebStory>
        {({ isLightTheme }) => (
            <SignInPage context={{ ...context, sourcegraphDotComMode: true }} authenticatedUser={null} />
        )}
    </WebStory>
)

export const OnlyOnePrimaryProvider: Story = () => (
    <WebStory>
        {({ isLightTheme }) => (
            <SignInPage context={{ ...context, primaryLoginProvidersCount: 1 }} authenticatedUser={null} />
        )}
    </WebStory>
)

export const OnlyOnePrimaryProviderWithoutBuiltIn: Story = () => (
    <WebStory>
        {({ isLightTheme }) => (
            <SignInPage
                context={{ ...context, primaryLoginProvidersCount: 1, authProviders: noBuiltInAuthProviders }}
                authenticatedUser={null}
            />
        )}
    </WebStory>
)

export const ShowMoreProviders: Story = () => (
    <WebStory initialEntries={['/sign-in?showMore']}>
        {({ isLightTheme }) => (
            <SignInPage context={{ ...context, primaryLoginProvidersCount: 1 }} authenticatedUser={null} />
        )}
    </WebStory>
)

export const ShowMoreProvidersWithoutBuiltIn: Story = () => (
    <WebStory initialEntries={['/sign-in?showMore']}>
        {({ isLightTheme }) => (
            <SignInPage
                context={{ ...context, authProviders: noBuiltInAuthProviders, primaryLoginProvidersCount: 1 }}
                authenticatedUser={null}
            />
        )}
    </WebStory>
)

export const OnlyBuiltInAuthProvider: Story = () => (
    <WebStory>
        {({ isLightTheme }) => (
            <SignInPage context={{ ...context, authProviders: onlyBuiltInAuthProvider }} authenticatedUser={null} />
        )}
    </WebStory>
)

export const PrefixCanBeChanged: Story = () => {
    const providers = noBuiltInAuthProviders.map(provider => ({ ...provider, displayPrefix: 'Just login with' }))

    return (
        <WebStory>
            {({ isLightTheme }) => (
                <SignInPage context={{ ...context, authProviders: providers }} authenticatedUser={null} />
            )}
        </WebStory>
    )
}
