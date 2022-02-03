import { storiesOf } from '@storybook/react'
import React from 'react'

import { NOOP_TELEMETRY_SERVICE } from '@sourcegraph/shared/src/telemetry/telemetryService'

import { WebStory } from '../../../../../../components/WebStory'
import { CodeInsightsBackendContext } from '../../../../core/backend/code-insights-backend-context'
import { CodeInsightsGqlBackend } from '../../../../core/backend/gql-api/code-insights-gql-backend'

import { IntroCreationPage } from './IntroCreationPage'

const { add } = storiesOf('web/insights/creation-ui/IntroPage', module)
    .addDecorator(story => <WebStory>{() => story()}</WebStory>)
    .addParameters({
        chromatic: {
            viewports: [576, 978, 1440],
            disableSnapshot: false,
        },
    })

const API = new CodeInsightsGqlBackend({} as any)

add('IntroPage', () => (
    <CodeInsightsBackendContext.Provider value={API}>
        <IntroCreationPage telemetryService={NOOP_TELEMETRY_SERVICE} />
    </CodeInsightsBackendContext.Provider>
))
