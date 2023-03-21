import { QueryResult } from '@apollo/client'
import { Observable } from 'rxjs'
import { map } from 'rxjs/operators'

import { dataOrThrowErrors, gql, useQuery } from '@sourcegraph/http-client'

import { queryGraphQL } from '../../backend/graphql'
import {
    useShowMorePagination,
    UseShowMorePaginationResult,
} from '../../components/FilteredConnection/hooks/useShowMorePagination'
import {
    SiteAdminPreciseIndexResult,
    WebhookByIdResult,
    WebhookByIdVariables,
    WebhookFields,
    WebhookLogFields,
    WebhookLogsByWebhookIDResult,
    WebhookLogsByWebhookIDVariables,
    WebhookPageHeaderResult,
    WebhookPageHeaderVariables,
    WebhooksListResult,
    WebhooksListVariables,
} from '../../graphql-operations'

import { WEBHOOK_LOGS_BY_ID } from './webhooks/backend'

/**
 * Fetch a single precise index by id.
 */
export function fetchPreciseIndex({
    id,
}: {
    id: string
}): Observable<Extract<SiteAdminPreciseIndexResult['node'], { __typename: 'PreciseIndex' }> | null> {
    return queryGraphQL<SiteAdminPreciseIndexResult>(
        gql`
            query SiteAdminPreciseIndex($id: ID!) {
                node(id: $id) {
                    __typename
                    ... on PreciseIndex {
                        projectRoot {
                            commit {
                                repository {
                                    name
                                    url
                                }
                            }
                        }
                    }
                }
            }
        `,
        { id }
    ).pipe(
        map(dataOrThrowErrors),
        map(({ node }) => {
            if (!node) {
                return null
            }
            if (node.__typename !== 'PreciseIndex') {
                throw new Error(`The given ID is a ${node.__typename}, not a PreciseIndex`)
            }

            return node
        })
    )
}

const WEBHOOK_FIELDS_FRAGMENT = gql`
    fragment WebhookFields on Webhook {
        id
        uuid
        url
        name
        codeHostKind
        codeHostURN
        secret
        updatedAt
        createdAt
        createdBy {
            username
            url
        }
        updatedBy {
            username
            url
        }
    }
`

export const WEBHOOKS = gql`
    ${WEBHOOK_FIELDS_FRAGMENT}

    query WebhooksList {
        webhooks {
            nodes {
                ...WebhookFields
            }
            totalCount
            pageInfo {
                hasNextPage
            }
        }
    }
`

export const WEBHOOK_BY_ID = gql`
    ${WEBHOOK_FIELDS_FRAGMENT}

    query WebhookById($id: ID!) {
        node(id: $id) {
            __typename
            ...WebhookFields
        }
    }
`

export const DELETE_WEBHOOK = gql`
    mutation DeleteWebhook($hookID: ID!) {
        deleteWebhook(id: $hookID) {
            alwaysNil
        }
    }
`

export const WEBHOOK_PAGE_HEADER = gql`
    query WebhookPageHeader {
        webhooks {
            nodes {
                webhookLogs {
                    totalCount
                }
            }
        }

        errorsOnly: webhooks {
            nodes {
                webhookLogs(onlyErrors: true) {
                    totalCount
                }
            }
        }
    }
`

export const useWebhookPageHeader = (): { loading: boolean; totalErrors: number; totalNoEvents: number } => {
    const { data, loading } = useQuery<WebhookPageHeaderResult, WebhookPageHeaderVariables>(WEBHOOK_PAGE_HEADER, {})
    const totalNoEvents = data?.webhooks.nodes.filter(webhook => webhook.webhookLogs?.totalCount === 0).length || 0
    const totalErrors =
        data?.errorsOnly.nodes.reduce((sum, webhook) => sum + (webhook.webhookLogs?.totalCount || 0), 0) || 0
    return { loading, totalErrors, totalNoEvents }
}

export const useWebhooksConnection = (): UseShowMorePaginationResult<WebhooksListResult, WebhookFields> =>
    useShowMorePagination<WebhooksListResult, WebhooksListVariables, WebhookFields>({
        query: WEBHOOKS,
        variables: {},
        getConnection: result => {
            const { webhooks } = dataOrThrowErrors(result)
            return webhooks
        },
    })

export const useWebhookQuery = (id: string): QueryResult<WebhookByIdResult, WebhookByIdVariables> =>
    useQuery<WebhookByIdResult, WebhookByIdVariables>(WEBHOOK_BY_ID, {
        variables: { id },
    })

export const useWebhookLogsConnection = (
    webhookID: string,
    first: number,
    onlyErrors: boolean
): UseShowMorePaginationResult<WebhookLogsByWebhookIDResult, WebhookLogFields> =>
    useShowMorePagination<WebhookLogsByWebhookIDResult, WebhookLogsByWebhookIDVariables, WebhookLogFields>({
        query: WEBHOOK_LOGS_BY_ID,
        variables: {
            first: first ?? 20,
            after: null,
            onlyErrors,
            onlyUnmatched: false,
            webhookID,
        },
        getConnection: result => {
            const { webhookLogs } = dataOrThrowErrors(result)
            return webhookLogs
        },
        options: {
            fetchPolicy: 'cache-first',
        },
    })

export const CREATE_WEBHOOK_QUERY = gql`
    mutation CreateWebhook($name: String!, $codeHostKind: String!, $codeHostURN: String!, $secret: String) {
        createWebhook(name: $name, codeHostKind: $codeHostKind, codeHostURN: $codeHostURN, secret: $secret) {
            id
        }
    }
`

export const UPDATE_WEBHOOK_QUERY = gql`
    mutation UpdateWebhook($id: ID!, $name: String!, $codeHostKind: String!, $codeHostURN: String!, $secret: String) {
        updateWebhook(id: $id, name: $name, codeHostKind: $codeHostKind, codeHostURN: $codeHostURN, secret: $secret) {
            id
        }
    }
`
