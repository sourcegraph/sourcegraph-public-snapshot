import type { MockedResponse } from '@apollo/client/testing'
import type { Decorator, StoryFn, Meta } from '@storybook/react'
import { subDays } from 'date-fns'

import { getDocumentNode } from '@sourcegraph/http-client'
import { noOpTelemetryRecorder } from '@sourcegraph/shared/src/telemetry'
import { MockedTestProvider } from '@sourcegraph/shared/src/testing/apollo'

import { WebStory } from '../../../components/WebStory'
import {
    type GlobalExecutorSecretsResult,
    ExecutorSecretScope,
    type UserExecutorSecretsResult,
} from '../../../graphql-operations'

import { GLOBAL_EXECUTOR_SECRETS, USER_EXECUTOR_SECRETS } from './backend'
import { GlobalExecutorSecretsListPage, UserExecutorSecretsListPage } from './ExecutorSecretsListPage'

const decorator: Decorator = story => <div className="p-3 container">{story()}</div>

const config: Meta = {
    title: 'web/executors/secrets/ExecutorSecretsListPage',
    decorators: [decorator],
    parameters: {},
}

export default config

const EXECUTOR_SECRET_LIST_MOCK: MockedResponse<UserExecutorSecretsResult> = {
    request: {
        query: getDocumentNode(USER_EXECUTOR_SECRETS),
        variables: {
            scope: ExecutorSecretScope.BATCHES,
            first: 15,
            after: null,
            user: 'user1',
        },
    },
    result: {
        data: {
            node: {
                __typename: 'User',
                executorSecrets: {
                    __typename: 'ExecutorSecretConnection',
                    pageInfo: {
                        hasNextPage: false,
                        endCursor: null,
                        __typename: 'PageInfo',
                    },
                    totalCount: 5,
                    nodes: [
                        // Global secret.
                        {
                            __typename: 'ExecutorSecret',
                            id: 'secret1',
                            creator: {
                                __typename: 'User',
                                id: 'user1',
                                displayName: 'John Doe',
                                url: '/users/jdoe',
                                username: 'jdoe',
                            },
                            key: 'GITHUB_TOKEN',
                            namespace: null,
                            overwritesGlobalSecret: false,
                            scope: ExecutorSecretScope.BATCHES,
                            createdAt: subDays(new Date(), 1).toISOString(),
                            updatedAt: subDays(new Date(), 1).toISOString(),
                        },
                        // User secret.
                        {
                            __typename: 'ExecutorSecret',
                            id: 'secret2',
                            creator: {
                                __typename: 'User',
                                id: 'user1',
                                displayName: 'John Doe',
                                url: '/users/jdoe',
                                username: 'jdoe',
                            },
                            key: 'NPM_TOKEN',
                            namespace: {
                                __typename: 'User',
                                id: 'user1',
                                namespaceName: 'jdoe',
                                url: '/users/jdoe',
                            },
                            overwritesGlobalSecret: false,
                            scope: ExecutorSecretScope.BATCHES,
                            createdAt: subDays(new Date(), 1).toISOString(),
                            updatedAt: subDays(new Date(), 1).toISOString(),
                        },
                        // Overwritten secret.
                        {
                            __typename: 'ExecutorSecret',
                            id: 'secret3',
                            creator: {
                                __typename: 'User',
                                id: 'user1',
                                displayName: 'John Doe',
                                url: '/users/jdoe',
                                username: 'jdoe',
                            },
                            key: 'DOCKER_PASS',
                            namespace: {
                                __typename: 'User',
                                id: 'user1',
                                namespaceName: 'jdoe',
                                url: '/users/jdoe',
                            },
                            overwritesGlobalSecret: true,
                            scope: ExecutorSecretScope.BATCHES,
                            createdAt: subDays(new Date(), 1).toISOString(),
                            updatedAt: subDays(new Date(), 1).toISOString(),
                        },
                        // Deleted creator.
                        {
                            __typename: 'ExecutorSecret',
                            id: 'secret4',
                            creator: null,
                            key: 'SRC_ACCESS_TOKEN',
                            namespace: {
                                __typename: 'User',
                                id: 'user1',
                                namespaceName: 'jdoe',
                                url: '/users/jdoe',
                            },
                            overwritesGlobalSecret: false,
                            scope: ExecutorSecretScope.BATCHES,
                            createdAt: subDays(new Date(), 1).toISOString(),
                            updatedAt: subDays(new Date(), 1).toISOString(),
                        },
                        // Docker auth secret.
                        {
                            __typename: 'ExecutorSecret',
                            id: 'secret5',
                            creator: {
                                __typename: 'User',
                                id: 'user1',
                                displayName: 'John Doe',
                                url: '/users/jdoe',
                                username: 'jdoe',
                            },
                            key: 'DOCKER_AUTH_CONFIG',
                            namespace: null,
                            overwritesGlobalSecret: false,
                            scope: ExecutorSecretScope.BATCHES,
                            createdAt: subDays(new Date(), 1).toISOString(),
                            updatedAt: subDays(new Date(), 1).toISOString(),
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
                <UserExecutorSecretsListPage {...webProps} userID="user1" telemetryRecorder={noOpTelemetryRecorder} />
            </MockedTestProvider>
        )}
    </WebStory>
)

List.storyName = 'List of secrets'

const EMPTY_EXECUTOR_SECRET_LIST_MOCK: MockedResponse<GlobalExecutorSecretsResult> = {
    request: {
        query: getDocumentNode(GLOBAL_EXECUTOR_SECRETS),
        variables: {
            scope: ExecutorSecretScope.BATCHES,
            first: 15,
            after: null,
        },
    },
    result: {
        data: {
            executorSecrets: {
                __typename: 'ExecutorSecretConnection',
                pageInfo: {
                    hasNextPage: false,
                    endCursor: null,
                    __typename: 'PageInfo',
                },
                totalCount: 0,
                nodes: [],
            },
        },
    },
}

export const EmptyList: StoryFn = () => (
    <WebStory>
        {webProps => (
            <MockedTestProvider mocks={[EMPTY_EXECUTOR_SECRET_LIST_MOCK]}>
                <GlobalExecutorSecretsListPage {...webProps} telemetryRecorder={noOpTelemetryRecorder} />
            </MockedTestProvider>
        )}
    </WebStory>
)

EmptyList.storyName = 'No secrets'
