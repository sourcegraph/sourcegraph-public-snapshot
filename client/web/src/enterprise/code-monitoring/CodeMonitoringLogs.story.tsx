import type { MockedResponse } from '@apollo/client/testing'
import type { Decorator, Meta, StoryFn } from '@storybook/react'
import { parseISO } from 'date-fns'

import { getDocumentNode } from '@sourcegraph/http-client'
import { MockedTestProvider } from '@sourcegraph/shared/src/testing/apollo'

import { WebStory } from '../../components/WebStory'

import { CodeMonitoringLogs, CODE_MONITOR_EVENTS } from './CodeMonitoringLogs'
import { mockLogs } from './testing/util'

const decorator: Decorator = story => <div className="p-3 container">{story()}</div>

const config: Meta = {
    title: 'web/enterprise/code-monitoring/CodeMonitoringLogs',
    decorators: [decorator],
    parameters: {
        chromatic: {
            disableSnapshot: false,
        },
    },
}

const mockedResponse: MockedResponse[] = [
    {
        request: {
            query: getDocumentNode(CODE_MONITOR_EVENTS),
            variables: { first: 20, after: null, triggerEventsFirst: 20, triggerEventsAfter: null },
        },
        result: { data: mockLogs },
    },
]

export default config

export const Default: StoryFn = () => (
    <WebStory>
        {() => (
            <MockedTestProvider mocks={mockedResponse}>
                <CodeMonitoringLogs now={() => parseISO('2022-02-14T16:21:00+00:00')} />
            </MockedTestProvider>
        )}
    </WebStory>
)

export const Open: StoryFn = () => (
    <WebStory>
        {() => (
            <MockedTestProvider mocks={mockedResponse}>
                <CodeMonitoringLogs now={() => parseISO('2022-02-14T16:21:00+00:00')} _testStartOpen={true} />
            </MockedTestProvider>
        )}
    </WebStory>
)

export const Empty: StoryFn = () => {
    const emptyMockedResponse: MockedResponse[] = [
        {
            request: {
                query: getDocumentNode(CODE_MONITOR_EVENTS),
                variables: { first: 20, after: null, triggerEventsFirst: 20, triggerEventsAfter: null },
            },
            result: {
                data: {
                    currentUser: {
                        monitors: { nodes: [], pageInfo: { hasNextPage: false, endCursor: null }, totalCount: 0 },
                    },
                },
            },
        },
    ]

    return (
        <WebStory>
            {() => (
                <MockedTestProvider mocks={emptyMockedResponse}>
                    <CodeMonitoringLogs now={() => parseISO('2022-02-14T16:21:00+00:00')} _testStartOpen={true} />
                </MockedTestProvider>
            )}
        </WebStory>
    )
}
