import { storiesOf } from '@storybook/react'
import React from 'react'
import sinon from 'sinon'

import { NOOP_TELEMETRY_SERVICE } from '@sourcegraph/shared/src/telemetry/telemetryService'

import { WebStory } from '../components/WebStory'
import { EMPTY_FEATURE_FLAGS } from '../featureFlags/featureFlags'
import { SourcegraphContext } from '../jscontext'

import { CloudSignUpPage } from './CloudSignUpPage'

const { add } = storiesOf('web/auth/CloudSignUpPage', module)

const context: Pick<SourcegraphContext, 'authProviders' | 'experimentalFeatures'> = {
    authProviders: [
        {
            serviceType: 'github',
            displayName: 'GitHub.com',
            isBuiltin: false,
            authenticationURL: '/.auth/github/login?pc=https%3A%2F%2Fgithub.com%2F',
        },
        {
            serviceType: 'gitlab',
            displayName: 'GitLab.com',
            isBuiltin: false,
            authenticationURL: '/.auth/gitlab/login?pc=https%3A%2F%2Fgitlab.com%2F',
        },
    ],
    experimentalFeatures: {},
}

add('default', () => (
    <WebStory>
        {({ isLightTheme }) => (
            <CloudSignUpPage
                isLightTheme={isLightTheme}
                source="Monitor"
                onSignUp={sinon.stub()}
                context={context}
                showEmailForm={false}
                telemetryService={NOOP_TELEMETRY_SERVICE}
                featureFlags={EMPTY_FEATURE_FLAGS}
            />
        )}
    </WebStory>
))

add('email form', () => (
    <WebStory>
        {({ isLightTheme }) => (
            <CloudSignUpPage
                isLightTheme={isLightTheme}
                source="SearchCTA"
                onSignUp={sinon.stub()}
                context={context}
                showEmailForm={true}
                telemetryService={NOOP_TELEMETRY_SERVICE}
                featureFlags={EMPTY_FEATURE_FLAGS}
            />
        )}
    </WebStory>
))

add('invalid source', () => (
    <WebStory>
        {({ isLightTheme }) => (
            <CloudSignUpPage
                isLightTheme={isLightTheme}
                source="test"
                onSignUp={sinon.stub()}
                context={context}
                showEmailForm={false}
                telemetryService={NOOP_TELEMETRY_SERVICE}
                featureFlags={EMPTY_FEATURE_FLAGS}
            />
        )}
    </WebStory>
))

add('Optimization signup', () => (
    <WebStory>
        {({ isLightTheme }) => (
            <CloudSignUpPage
                isLightTheme={isLightTheme}
                source="test"
                onSignUp={sinon.stub()}
                context={context}
                showEmailForm={false}
                telemetryService={NOOP_TELEMETRY_SERVICE}
                featureFlags={new Map([['signup-optimization', true]])}
            />
        )}
    </WebStory>
))
