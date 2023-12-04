import { dataOrThrowErrors, gql } from '@sourcegraph/http-client'

import {
    useShowMorePagination,
    type UseShowMorePaginationResult,
} from '../../components/FilteredConnection/hooks/useShowMorePagination'
import type {
    OutboundWebhookFieldsWithStats,
    OutboundWebhooksListResult,
    OutboundWebhooksListVariables,
} from '../../graphql-operations'

const OUTBOUND_WEBHOOK_FIELDS_FRAGMENT = gql`
    fragment OutboundWebhookFields on OutboundWebhook {
        id
        url
        eventTypes {
            eventType
            scope
        }
    }
`

const OUTBOUND_WEBHOOK_FIELDS_WITH_STATS_FRAGMENT = gql`
    ${OUTBOUND_WEBHOOK_FIELDS_FRAGMENT}

    fragment OutboundWebhookFieldsWithStats on OutboundWebhook {
        ...OutboundWebhookFields
        stats {
            total
            errored
        }
    }
`

export const CREATE_OUTBOUND_WEBHOOK = gql`
    ${OUTBOUND_WEBHOOK_FIELDS_FRAGMENT}

    mutation CreateOutboundWebhook($input: OutboundWebhookCreateInput!) {
        createOutboundWebhook(input: $input) {
            ...OutboundWebhookFields
        }
    }
`

export const DELETE_OUTBOUND_WEBHOOK = gql`
    mutation DeleteOutboundWebhook($id: ID!) {
        deleteOutboundWebhook(id: $id) {
            alwaysNil
        }
    }
`

export const OUTBOUND_WEBHOOK_BY_ID = gql`
    ${OUTBOUND_WEBHOOK_FIELDS_FRAGMENT}

    query OutboundWebhookByID($id: ID!) {
        node(id: $id) {
            ... on OutboundWebhook {
                ...OutboundWebhookFields
            }
        }
    }
`

export const OUTBOUND_WEBHOOKS = gql`
    ${OUTBOUND_WEBHOOK_FIELDS_WITH_STATS_FRAGMENT}

    query OutboundWebhooksList($first: Int, $after: String) {
        outboundWebhooks(first: $first, after: $after) {
            nodes {
                ...OutboundWebhookFieldsWithStats
            }
            totalCount
            pageInfo {
                hasNextPage
            }
        }
    }
`

export const UPDATE_OUTBOUND_WEBHOOK = gql`
    ${OUTBOUND_WEBHOOK_FIELDS_FRAGMENT}

    mutation UpdateOutboundWebhook($id: ID!, $input: OutboundWebhookUpdateInput!) {
        updateOutboundWebhook(id: $id, input: $input) {
            ...OutboundWebhookFields
        }
    }
`

export const useOutboundWebhooksConnection = (): UseShowMorePaginationResult<
    OutboundWebhooksListResult,
    OutboundWebhookFieldsWithStats
> =>
    useShowMorePagination<OutboundWebhooksListResult, OutboundWebhooksListVariables, OutboundWebhookFieldsWithStats>({
        query: OUTBOUND_WEBHOOKS,
        variables: {
            first: 20,
            after: null,
        },
        getConnection: result => {
            const { outboundWebhooks } = dataOrThrowErrors(result)
            return outboundWebhooks
        },
        options: {
            pollInterval: 5000,
        },
    })

const OUTBOUND_WEBHOOK_EVENT_TYPE_FRAGMENT = gql`
    fragment OutboundWebhookEventTypeFields on OutboundWebhookEventType {
        key
        description
    }
`

export const OUTBOUND_WEBHOOK_EVENT_TYPES = gql`
    ${OUTBOUND_WEBHOOK_EVENT_TYPE_FRAGMENT}

    query OutboundWebhookEventTypes {
        outboundWebhookEventTypes {
            ...OutboundWebhookEventTypeFields
        }
    }
`
