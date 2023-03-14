import { MockedResponse } from '@apollo/client/testing'

import { getDocumentNode } from '@sourcegraph/http-client'

import { StatusAndRepoStatsResult } from '../graphql-operations'

import { STATUS_AND_REPO_STATS } from './StatusMessagesNavItemQueries'

export const allStatusMessages: StatusAndRepoStatsResult['statusMessages'] = [
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
    messages: StatusAndRepoStatsResult['statusMessages']
): MockedResponse<StatusAndRepoStatsResult> => ({
    request: {
        query: getDocumentNode(STATUS_AND_REPO_STATS),
    },
    result: {
        data: {
            statusMessages: messages,
        },
    },
})
