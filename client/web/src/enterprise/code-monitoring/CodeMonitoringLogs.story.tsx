import { MockedResponse } from '@apollo/client/testing'
import { storiesOf } from '@storybook/react'
import { parseISO } from 'date-fns'
import React from 'react'

import { getDocumentNode } from '@sourcegraph/http-client'
import { MockedTestProvider } from '@sourcegraph/shared/src/testing/apollo'

import { WebStory } from '../../components/WebStory'

import { CodeMonitoringLogs, CODE_MONITOR_EVENTS } from './CodeMonitoringLogs'
import { mockLogs } from './testing/util'

const { add } = storiesOf('web/enterprise/code-monitoring/CodeMonitoringLogs', module)
    .addDecorator(story => <div className="p-3 container">{story()}</div>)
    .addParameters({
        chromatic: {
            disableSnapshot: false,
        },
    })

const mockedResponse: MockedResponse[] = [
    {
        request: {
            query: getDocumentNode(CODE_MONITOR_EVENTS),
            variables: { first: 20, after: null, triggerEventsFirst: 20, triggerEventsAfter: null },
        },
        result: { data: mockLogs },
    },
]

add('CodeMonitoringLogs', () => (
    <WebStory>
        {() => (
            <MockedTestProvider mocks={mockedResponse}>
                <CodeMonitoringLogs now={() => parseISO('2022-02-14T16:21:00+00:00')} />
            </MockedTestProvider>
        )}
    </WebStory>
))
