import type { Meta, Story } from '@storybook/react'

import { NOOP_TELEMETRY_SERVICE } from '@sourcegraph/shared/src/telemetry/telemetryService'

import { WebStory } from '../components/WebStory'
import { SourcegraphContext } from '../jscontext'

import { SignUpPage } from './SignUpPage'

const config: Meta = {
    title: 'web/auth/SignUpPage',
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

export const Default: Story = () => (
    <WebStory>
        {() => (
            <SignUpPage
                telemetryService={NOOP_TELEMETRY_SERVICE}
                context={{
                    authProviders,
                    xhrHeaders: {},
                    allowSignup: true,
                    authMinPasswordLength: 12,
                    authPasswordPolicy: {},
                    sourcegraphDotComMode: false,
                }}
                authenticatedUser={null}
            />
        )}
    </WebStory>
)
