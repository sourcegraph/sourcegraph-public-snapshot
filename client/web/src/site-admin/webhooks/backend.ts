import { Observable } from 'rxjs'
import { map } from 'rxjs/operators'

import { dataOrThrowErrors, gql } from '@sourcegraph/http-client'

import { requestGraphQL } from '../../backend/graphql'
import {
    Scalars,
    ServiceWebhookLogsResult,
    ServiceWebhookLogsVariables,
    WebhookLogConnectionFields,
    WebhookLogsResult,
    WebhookLogsVariables,
} from '../../graphql-operations'

export type SelectedExternalService = 'unmatched' | 'all' | Scalars['ID']

export const queryWebhookLogs = (
    { first, after }: Pick<WebhookLogsVariables, 'first' | 'after'>,
    externalService: SelectedExternalService,
    onlyErrors: boolean
): Observable<WebhookLogConnectionFields> => {
    const fragment = gql`
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

        fragment WebhookLogFields on WebhookLog {
            id
            receivedAt
            externalService {
                displayName
            }
            statusCode
            request {
                ...WebhookLogMessageFields
                ...WebhookLogRequestFields
            }
            response {
                ...WebhookLogMessageFields
            }
        }

        fragment WebhookLogMessageFields on WebhookLogMessage {
            headers {
                name
                values
            }
            body
        }

        fragment WebhookLogRequestFields on WebhookLogRequest {
            method
            url
            version
        }
    `

    if (externalService === 'all' || externalService === 'unmatched') {
        return requestGraphQL<WebhookLogsResult, WebhookLogsVariables>(
            gql`
                query WebhookLogs($first: Int, $after: String, $onlyErrors: Boolean!, $onlyUnmatched: Boolean!) {
                    webhookLogs(first: $first, after: $after, onlyErrors: $onlyErrors, onlyUnmatched: $onlyUnmatched) {
                        ...WebhookLogConnectionFields
                    }
                }

                ${fragment}
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

            ${fragment}
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

        webhookLogs(onlyErrors: true) {
            totalCount
        }
    }

    fragment WebhookLogPageHeaderExternalService on ExternalService {
        id
        displayName
    }
`
