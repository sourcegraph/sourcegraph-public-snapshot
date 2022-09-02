import { MockedResponse } from '@apollo/client/testing'

import { getDocumentNode } from '@sourcegraph/http-client'

import { StatusMessagesResult } from '../graphql-operations'

import { STATUS_MESSAGES } from './StatusMessagesNavItemQueries'

export const allStatusMessages: StatusMessagesResult['statusMessages'] = [
    {
        __typename: 'ExternalServiceSyncError',
        message:
            'This is a\nmulti line error message\nthat spans multiple lines and also is a bit long on some lines\nbut lets see',
        externalService: {
            id: 'RXh0ZXJuYWxTZXJ2aWNlOjE=',
            displayName: 'GitHub PRODUCTION',
            __typename: 'ExternalService',
        },
    },
    {
        __typename: 'SyncError',
        message: '27 repositories could not be synced',
    },
    {
        __typename: 'CloningProgress',
        message: '477260 repositories enqueued for cloning. 11 repositories currently cloning...',
    },
]

export const newStatusMessageMock = (
    messages: StatusMessagesResult['statusMessages']
): MockedResponse<StatusMessagesResult> => ({
    request: {
        query: getDocumentNode(STATUS_MESSAGES),
    },
    result: {
        data: {
            statusMessages: messages,
        },
    },
})
