import { MockedResponse } from '@apollo/client/testing'

import { getDocumentNode } from '@sourcegraph/http-client'

import { OutboundWebhookEventTypesResult } from '../../graphql-operations'

import { OUTBOUND_WEBHOOK_EVENT_TYPES } from './backend'

export const eventTypesMock: MockedResponse<OutboundWebhookEventTypesResult> = {
    request: {
        query: getDocumentNode(OUTBOUND_WEBHOOK_EVENT_TYPES),
    },
    result: {
        data: {
            outboundWebhookEventTypes: [
                {
                    __typename: 'OutboundWebhookEventType',
                    key: 'batch_change:apply',
                    description: 'sent when a batch change is applied',
                },
                {
                    __typename: 'OutboundWebhookEventType',
                    key: 'batch_change:close',
                    description: 'sent when a batch change is closed',
                },
            ],
        },
    },
}
