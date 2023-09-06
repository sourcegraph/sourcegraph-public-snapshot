import type { MockedResponse } from '@apollo/client/testing'

import { getDocumentNode } from '@sourcegraph/http-client'

import type { StatusAndRepoCountResult } from '../graphql-operations'

import { STATUS_AND_REPO_COUNT } from './StatusMessagesNavItemQueries'

export const allStatusMessages: StatusAndRepoCountResult['statusMessages'] = [
    {
        __typename: 'ExternalServiceSyncError',
        externalService: {
            id: 'RXh0ZXJuYWxTZXJ2aWNlOjE=',
            displayName: 'GitHub PRODUCTION',
            __typename: 'ExternalService',
        },
    },
    {
        __typename: 'SyncError',
        message: '13 repositories failed last attempt to sync content from code host',
    },
    {
        __typename: 'CloningProgress',
        message: '477260 repositories enqueued for cloning. 11 repositories currently cloning...',
    },
    {
        __typename: 'IndexingProgress',
        indexed: 15,
        notIndexed: 23,
    },
]

export const newStatusMessageMock = (
    messages: StatusAndRepoCountResult['statusMessages']
): MockedResponse<StatusAndRepoCountResult> => ({
    request: {
        query: getDocumentNode(STATUS_AND_REPO_COUNT),
    },
    result: {
        data: {
            statusMessages: messages,
            repositoryStats: {
                __typename: 'RepositoryStats',
                total: 7,
            },
        },
    },
})
