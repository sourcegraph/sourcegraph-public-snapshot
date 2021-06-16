import { storiesOf } from '@storybook/react'
import React from 'react'
import sinon from 'sinon'

import { NOOP_TELEMETRY_SERVICE } from '@sourcegraph/shared/src/telemetry/telemetryService'

import { WebStory } from '../components/WebStory'
import { SourcegraphContext } from '../jscontext'

import { ExperimentalSignUpPage } from './ExperimentalSignUpPage'

const { add } = storiesOf('web/auth/ExperimentalSignUpPage', module)

const context: Pick<SourcegraphContext, 'authProviders'> = {
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
}

add('default', () => (
    <WebStory>
        {({ isLightTheme }) => (
            <ExperimentalSignUpPage
                isLightTheme={isLightTheme}
                source="Monitor"
                onSignUp={sinon.stub()}
                context={context}
                showEmailForm={false}
                telemetryService={NOOP_TELEMETRY_SERVICE}
            />
        )}
    </WebStory>
))

add('email form', () => (
    <WebStory>
        {({ isLightTheme }) => (
            <ExperimentalSignUpPage
                isLightTheme={isLightTheme}
                source="SearchCTA"
                onSignUp={sinon.stub()}
                context={context}
                showEmailForm={true}
                telemetryService={NOOP_TELEMETRY_SERVICE}
            />
        )}
    </WebStory>
))

add('invalid source', () => (
    <WebStory>
        {({ isLightTheme }) => (
            <ExperimentalSignUpPage
                isLightTheme={isLightTheme}
                source="test"
                onSignUp={sinon.stub()}
                context={context}
                showEmailForm={false}
                telemetryService={NOOP_TELEMETRY_SERVICE}
            />
        )}
    </WebStory>
))
