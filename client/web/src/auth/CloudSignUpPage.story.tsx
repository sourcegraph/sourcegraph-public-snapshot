import { Meta, Story } from '@storybook/react'
import sinon from 'sinon'

import { NOOP_TELEMETRY_SERVICE } from '@sourcegraph/shared/src/telemetry/telemetryService'

import { WebStory } from '../components/WebStory'
import { SourcegraphContext } from '../jscontext'

import { CloudSignUpPage } from './CloudSignUpPage'

const config: Meta = {
    title: 'web/auth/CloudSignUpPage',
}

export default config

const context: Pick<SourcegraphContext, 'authProviders' | 'experimentalFeatures' | 'authMinPasswordLength'> = {
    authProviders: [
        {
            serviceType: 'github',
            displayName: 'GitHub.com',
            isBuiltin: false,
            authenticationURL: '/.auth/github/login?pc=https%3A%2F%2Fgithub.com%2F',
            serviceID: 'https://github.com',
        },
        {
            serviceType: 'gitlab',
            displayName: 'GitLab.com',
            isBuiltin: false,
            authenticationURL: '/.auth/gitlab/login?pc=https%3A%2F%2Fgitlab.com%2F',
            serviceID: 'https://gitlab.com',
        },
    ],
    experimentalFeatures: {},
    authMinPasswordLength: 0,
}

export const Default: Story = () => (
    <WebStory>
        {({ isLightTheme }) => (
            <CloudSignUpPage
                isLightTheme={isLightTheme}
                source="Monitor"
                onSignUp={sinon.stub()}
                context={context}
                showEmailForm={false}
                telemetryService={NOOP_TELEMETRY_SERVICE}
            />
        )}
    </WebStory>
)

export const EmailForm: Story = () => (
    <WebStory>
        {({ isLightTheme }) => (
            <CloudSignUpPage
                isLightTheme={isLightTheme}
                source="SearchCTA"
                onSignUp={sinon.stub()}
                context={context}
                showEmailForm={true}
                telemetryService={NOOP_TELEMETRY_SERVICE}
            />
        )}
    </WebStory>
)

export const InvalidSource: Story = () => (
    <WebStory>
        {({ isLightTheme }) => (
            <CloudSignUpPage
                isLightTheme={isLightTheme}
                source="test"
                onSignUp={sinon.stub()}
                context={context}
                showEmailForm={false}
                telemetryService={NOOP_TELEMETRY_SERVICE}
            />
        )}
    </WebStory>
)

export const OptimizationSignup: Story = () => (
    <WebStory>
        {({ isLightTheme }) => (
            <CloudSignUpPage
                isLightTheme={isLightTheme}
                source="test"
                onSignUp={sinon.stub()}
                context={context}
                showEmailForm={false}
                telemetryService={NOOP_TELEMETRY_SERVICE}
            />
        )}
    </WebStory>
)
