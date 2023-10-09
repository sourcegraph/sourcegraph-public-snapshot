import type { Meta, Story } from '@storybook/react'

import { WebStory } from '../components/WebStory'
import type { SourcegraphContext } from '../jscontext'

import { SignInPage, type SignInPageProps } from './SignInPage'

const config: Meta = {
    title: 'web/auth/SignInPage',
}

export default config

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

export const Default: Story = () => (
    <WebStory>{() => <SignInPage context={context} authenticatedUser={null} />}</WebStory>
)

export const ShowMore: Story = () => (
    <WebStory initialEntries={[{ pathname: '/sign-in', search: '?showMore' }]}>
        {() => <SignInPage context={{ ...context, primaryLoginProvidersCount: 1 }} authenticatedUser={null} />}
    </WebStory>
)

export const Dotcom: Story = () => (
    <WebStory>
        {() => <SignInPage context={{ ...context, sourcegraphDotComMode: true }} authenticatedUser={null} />}
    </WebStory>
)

export const NoProviders: Story = () => (
    <WebStory>{() => <SignInPage context={{ ...context, authProviders: [] }} authenticatedUser={null} />}</WebStory>
)

export const NoBuiltIn: Story = () => (
    <WebStory>
        {() => <SignInPage context={{ ...context, authProviders: noBuiltInAuthProviders }} authenticatedUser={null} />}
    </WebStory>
)

export const NoResetPassword: Story = () => (
    <WebStory>
        {() => <SignInPage context={{ ...context, resetPasswordEnabled: false }} authenticatedUser={null} />}
    </WebStory>
)

export const NoSignUp: Story = () => (
    <WebStory>{() => <SignInPage context={{ ...context, allowSignup: false }} authenticatedUser={null} />}</WebStory>
)

export const NoAccessRequest: Story = () => (
    <WebStory>
        {() => (
            <SignInPage
                context={{ ...context, allowSignup: false, authAccessRequest: { enabled: false } }}
                authenticatedUser={null}
            />
        )}
    </WebStory>
)

export const DotComSignUp: Story = () => (
    <WebStory>
        {() => <SignInPage context={{ ...context, sourcegraphDotComMode: true }} authenticatedUser={null} />}
    </WebStory>
)

export const OnlyOnePrimaryProvider: Story = () => (
    <WebStory>
        {() => <SignInPage context={{ ...context, primaryLoginProvidersCount: 1 }} authenticatedUser={null} />}
    </WebStory>
)

export const OnlyOnePrimaryProviderWithoutBuiltIn: Story = () => (
    <WebStory>
        {() => (
            <SignInPage
                context={{ ...context, primaryLoginProvidersCount: 1, authProviders: noBuiltInAuthProviders }}
                authenticatedUser={null}
            />
        )}
    </WebStory>
)

export const ShowMoreProviders: Story = () => (
    <WebStory initialEntries={['/sign-in?showMore']}>
        {() => <SignInPage context={{ ...context, primaryLoginProvidersCount: 1 }} authenticatedUser={null} />}
    </WebStory>
)

export const ShowMoreProvidersWithoutBuiltIn: Story = () => (
    <WebStory initialEntries={['/sign-in?showMore']}>
        {() => (
            <SignInPage
                context={{ ...context, authProviders: noBuiltInAuthProviders, primaryLoginProvidersCount: 1 }}
                authenticatedUser={null}
            />
        )}
    </WebStory>
)

export const OnlyBuiltInAuthProvider: Story = () => (
    <WebStory>
        {() => <SignInPage context={{ ...context, authProviders: onlyBuiltInAuthProvider }} authenticatedUser={null} />}
    </WebStory>
)

export const PrefixCanBeChanged: Story = () => {
    const providers = noBuiltInAuthProviders.map(provider => ({ ...provider, displayPrefix: 'Just login with' }))

    return (
        <WebStory>
            {() => <SignInPage context={{ ...context, authProviders: providers }} authenticatedUser={null} />}
        </WebStory>
    )
}
