import type { MockedResponse } from '@apollo/client/testing'
import type { Decorator, StoryFn, Meta } from '@storybook/react'
import { subDays } from 'date-fns'
import { noop } from 'lodash'

import { getDocumentNode } from '@sourcegraph/http-client'
import { MockedTestProvider } from '@sourcegraph/shared/src/testing/apollo'

import { WebStory } from '../../../components/WebStory'
import type { ExecutorSecretAccessLogsResult } from '../../../graphql-operations'

import { EXECUTOR_SECRET_ACCESS_LOGS } from './backend'
import { SecretAccessLogsModal } from './SecretAccessLogsModal'

const decorator: Decorator = story => <div className="p-3 container">{story()}</div>

const config: Meta = {
    title: 'web/executors/secrets/SecretAccessLogsModal',
    decorators: [decorator],
    parameters: {},
}

export default config

const EXECUTOR_SECRET_LIST_MOCK: MockedResponse<ExecutorSecretAccessLogsResult> = {
    request: {
        query: getDocumentNode(EXECUTOR_SECRET_ACCESS_LOGS),
        variables: {
            first: 15,
            after: null,
            secret: 'secret1',
        },
    },
    result: {
        data: {
            node: {
                __typename: 'ExecutorSecret',
                accessLogs: {
                    pageInfo: { hasNextPage: false, endCursor: null },
                    totalCount: 2,
                    nodes: [
                        {
                            __typename: 'ExecutorSecretAccessLog',
                            id: 'log1',
                            user: {
                                __typename: 'User',
                                id: 'user1',
                                displayName: 'John Doe',
                                url: '/users/jdoe',
                                username: 'jdoe',
                                email: 'jdoe@example.com',
                            },
                            createdAt: subDays(new Date(), 3).toISOString(),
                        },
                        {
                            __typename: 'ExecutorSecretAccessLog',
                            id: 'log2',
                            user: {
                                __typename: 'User',
                                id: 'user1',
                                displayName: 'John Doe',
                                url: '/users/jdoe',
                                username: 'jdoe',
                                email: 'jdoe@example.com',
                            },
                            createdAt: subDays(new Date(), 1).toISOString(),
                        },
                    ],
                },
            },
        },
    },
}

export const List: StoryFn = () => (
    <WebStory>
        {webProps => (
            <MockedTestProvider mocks={[EXECUTOR_SECRET_LIST_MOCK]}>
                <SecretAccessLogsModal {...webProps} onCancel={noop} secretID="secret1" />
            </MockedTestProvider>
        )}
    </WebStory>
)

List.storyName = 'List of access logs'

const EMPTY_SECRET_ACCESS_LOGS_LIST_MOCK: MockedResponse<ExecutorSecretAccessLogsResult> = {
    request: {
        query: getDocumentNode(EXECUTOR_SECRET_ACCESS_LOGS),
        variables: {
            first: 15,
            after: null,
            secret: 'secret1',
        },
    },
    result: {
        data: {
            node: {
                __typename: 'ExecutorSecret',
                accessLogs: {
                    pageInfo: { hasNextPage: false, endCursor: null },
                    totalCount: 0,
                    nodes: [],
                },
            },
        },
    },
}

export const EmptyList: StoryFn = () => (
    <WebStory>
        {webProps => (
            <MockedTestProvider mocks={[EMPTY_SECRET_ACCESS_LOGS_LIST_MOCK]}>
                <SecretAccessLogsModal {...webProps} onCancel={noop} secretID="secret1" />
            </MockedTestProvider>
        )}
    </WebStory>
)

EmptyList.storyName = 'No access logs'
