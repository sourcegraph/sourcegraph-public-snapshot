import type { Observable } from 'rxjs'
import { map } from 'rxjs/operators'

import { dataOrThrowErrors, gql } from '@sourcegraph/http-client'

import { requestGraphQL } from '../../backend/graphql'
import type {
    Scalars,
    ServiceWebhookLogsResult,
    ServiceWebhookLogsVariables,
    WebhookLogConnectionFields,
    WebhookLogsByWebhookIDResult,
    WebhookLogsByWebhookIDVariables,
    WebhookLogsResult,
    WebhookLogsVariables,
} from '../../graphql-operations'

export type SelectedExternalService = 'unmatched' | 'all' | Scalars['ID']

export const WEBHOOK_LOG_REQUEST_FIELDS_FRAGMENT = gql`
    fragment WebhookLogRequestFields on WebhookLogRequest {
        headers {
            name
            values
        }
        body
        method
        url
        version
    }
`

export const WEBHOOK_LOG_RESPONSE_FIELDS_FRAGMENT = gql`
    fragment WebhookLogResponseFields on WebhookLogResponse {
        headers {
            name
            values
        }
        body
    }
`

const WEBHOOK_LOG_FIELDS_FRAGMENT = gql`
    ${WEBHOOK_LOG_REQUEST_FIELDS_FRAGMENT}
    ${WEBHOOK_LOG_RESPONSE_FIELDS_FRAGMENT}

    fragment WebhookLogFields on WebhookLog {
        id
        receivedAt
        externalService {
            displayName
        }
        statusCode
        request {
            ...WebhookLogRequestFields
        }
        response {
            ...WebhookLogResponseFields
        }
    }
`

const WEBHOOK_LOG_CONNECTION_FIELDS_FRAGMENT = gql`
    ${WEBHOOK_LOG_FIELDS_FRAGMENT}
    fragment WebhookLogConnectionFields on WebhookLogConnection {
        nodes {
            ...WebhookLogFields
        }
        pageInfo {
            hasNextPage
            endCursor
        }
        totalCount
    }
`

export const queryWebhookLogs = (
    { first, after }: Pick<WebhookLogsVariables, 'first' | 'after'>,
    externalService: SelectedExternalService,
    onlyErrors: boolean,
    webhookID?: string
): Observable<WebhookLogConnectionFields> => {
    // If webhook ID is provided, then we search for this webhook's logs
    if (webhookID) {
        return requestGraphQL<WebhookLogsByWebhookIDResult, WebhookLogsByWebhookIDVariables>(WEBHOOK_LOGS_BY_ID, {
            first: first ?? 20,
            after: null,
            onlyErrors: false,
            onlyUnmatched: false,
            webhookID,
        }).pipe(
            map(dataOrThrowErrors),
            map((result: WebhookLogsResult) => result.webhookLogs)
        )
    }

    if (externalService === 'all' || externalService === 'unmatched') {
        return requestGraphQL<WebhookLogsResult, WebhookLogsVariables>(
            gql`
                query WebhookLogs($first: Int, $after: String, $onlyErrors: Boolean!, $onlyUnmatched: Boolean!) {
                    webhookLogs(
                        first: $first
                        after: $after
                        onlyErrors: $onlyErrors
                        onlyUnmatched: $onlyUnmatched
                        legacyOnly: true
                    ) {
                        ...WebhookLogConnectionFields
                    }
                }

                ${WEBHOOK_LOG_CONNECTION_FIELDS_FRAGMENT}
            `,
            {
                first,
                after,
                onlyErrors,
                onlyUnmatched: externalService === 'unmatched',
            }
        ).pipe(
            map(dataOrThrowErrors),
            map((result: WebhookLogsResult) => result.webhookLogs)
        )
    }

    return requestGraphQL<ServiceWebhookLogsResult, ServiceWebhookLogsVariables>(
        gql`
            query ServiceWebhookLogs($first: Int, $after: String, $id: ID!, $onlyErrors: Boolean!) {
                node(id: $id) {
                    ... on ExternalService {
                        __typename
                        webhookLogs(first: $first, after: $after, onlyErrors: $onlyErrors) {
                            ...WebhookLogConnectionFields
                        }
                    }
                }
            }

            ${WEBHOOK_LOG_CONNECTION_FIELDS_FRAGMENT}
        `,
        {
            first: first ?? null,
            after: after ?? null,
            onlyErrors,
            id: externalService,
        }
    ).pipe(
        map(dataOrThrowErrors),
        map(result => {
            if (result.node?.__typename === 'ExternalService') {
                return result.node.webhookLogs
            }
            throw new Error('unexpected non ExternalService node')
        })
    )
}

export const WEBHOOK_LOG_PAGE_HEADER = gql`
    query WebhookLogPageHeader {
        externalServices {
            nodes {
                ...WebhookLogPageHeaderExternalService
            }
            totalCount
        }

        webhookLogs(onlyErrors: true, legacyOnly: true) {
            totalCount
        }
    }

    fragment WebhookLogPageHeaderExternalService on ExternalService {
        id
        displayName
    }
`

export const WEBHOOK_LOGS_BY_ID = gql`
    ${WEBHOOK_LOG_FIELDS_FRAGMENT}

    query WebhookLogsByWebhookID(
        $first: Int
        $after: String
        $onlyErrors: Boolean!
        $onlyUnmatched: Boolean!
        $webhookID: ID!
    ) {
        webhookLogs(
            first: $first
            after: $after
            onlyErrors: $onlyErrors
            onlyUnmatched: $onlyUnmatched
            webhookID: $webhookID
        ) {
            nodes {
                ...WebhookLogFields
            }
            pageInfo {
                hasNextPage
                endCursor
            }
            totalCount
        }
    }
`

export const WEBHOOK_BY_ID_LOG_PAGE_HEADER = gql`
    query WebhookByIDLogPageHeader($webhookID: ID!) {
        webhookLogs(webhookID: $webhookID, onlyErrors: true) {
            totalCount
        }
    }
`
