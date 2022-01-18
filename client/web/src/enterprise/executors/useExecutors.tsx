import { ApolloClient } from '@apollo/client'
import { from, Observable } from 'rxjs'
import { map } from 'rxjs/operators'

import { gql, getDocumentNode } from '@sourcegraph/http-client'
import * as GQL from '@sourcegraph/shared/src/schema'

import { ExecutorFields, ExecutorsResult, ExecutorsVariables } from '../../graphql-operations'

export const executorFieldsFragment = gql`
    fragment ExecutorFields on Executor {
        __typename
        id
        hostname
        queueName
        active
        os
        architecture
        dockerVersion
        executorVersion
        gitVersion
        igniteVersion
        srcCliVersion
        firstSeenAt
        lastSeenAt
    }
`

interface ExecutorConnection {
    nodes: ExecutorFields[]
    totalCount: number | null
    pageInfo: { endCursor: string | null; hasNextPage: boolean }
}

const EXECUTORS = gql`
    query Executors($query: String, $active: Boolean, $first: Int, $after: String) {
        executors(query: $query, active: $active, first: $first, after: $after) {
            nodes {
                ...ExecutorFields
            }
            totalCount
            pageInfo {
                endCursor
                hasNextPage
            }
        }
    }

    ${executorFieldsFragment}
`

export const queryExecutors = (
    { query, active, first, after }: GQL.IExecutorsOnQueryArguments,
    client: ApolloClient<object>
): Observable<ExecutorConnection> => {
    const vars: ExecutorsVariables = {
        query: query ?? null,
        active: active ?? null,
        first: first ?? null,
        after: after ?? null,
    }

    return from(
        client.query<ExecutorsResult>({
            query: getDocumentNode(EXECUTORS),
            variables: vars,
        })
    ).pipe(
        map(({ data }) => data),
        map(({ executors }) => executors)
    )
}
