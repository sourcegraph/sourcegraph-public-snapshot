import type { Meta, StoryFn } from '@storybook/react'

import { noOpTelemetryRecorder } from '@sourcegraph/shared/src/telemetry'
import { NOOP_TELEMETRY_SERVICE } from '@sourcegraph/shared/src/telemetry/telemetryService'

import { WebStory } from '../components/WebStory'
import type { SourcegraphContext } from '../jscontext'

import { VsCodeSignUpPage, type VsCodeSignUpPageProps } from './VsCodeSignUpPage'

const config: Meta = {
    title: 'web/auth/VsCodeSignUpPage',
    parameters: {
        chromatic: { disableSnapshot: false },
    },
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

const context: VsCodeSignUpPageProps['context'] = {
    authProviders,
    authMinPasswordLength: 12,
}

export const WithoutEmailForm: StoryFn = () => (
    <WebStory>
        {() => (
            <VsCodeSignUpPage
                onSignUp={async () => {}}
                showEmailForm={false}
                source="test"
                telemetryService={NOOP_TELEMETRY_SERVICE}
                telemetryRecorder={noOpTelemetryRecorder}
                context={context}
            />
        )}
    </WebStory>
)

export const WithEmailForm: StoryFn = () => (
    <WebStory>
        {() => (
            <VsCodeSignUpPage
                onSignUp={async () => {}}
                showEmailForm={true}
                source="test"
                telemetryService={NOOP_TELEMETRY_SERVICE}
                telemetryRecorder={noOpTelemetryRecorder}
                context={context}
            />
        )}
    </WebStory>
)
