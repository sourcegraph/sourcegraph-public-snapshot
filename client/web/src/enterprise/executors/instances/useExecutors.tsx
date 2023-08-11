import type { ApolloClient } from '@apollo/client'
import { from, type Observable } from 'rxjs'
import { map } from 'rxjs/operators'

import { gql, getDocumentNode } from '@sourcegraph/http-client'

import type { ExecutorsResult, ExecutorsVariables, ExecutorConnectionFields } from '../../../graphql-operations'

export const executorFieldsFragment = gql`
    fragment ExecutorFields on Executor {
        __typename
        id
        hostname
        queueName
        queueNames
        active
        os
        compatibility
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

const EXECUTORS = gql`
    query Executors($query: String, $active: Boolean, $first: Int, $after: String) {
        executors(query: $query, active: $active, first: $first, after: $after) {
            ...ExecutorConnectionFields
        }
    }

    fragment ExecutorConnectionFields on ExecutorConnection {
        nodes {
            ...ExecutorFields
        }
        totalCount
        pageInfo {
            endCursor
            hasNextPage
        }
    }

    ${executorFieldsFragment}
`

export const queryExecutors = (
    { query, active, first, after }: Partial<ExecutorsVariables>,
    client: ApolloClient<object>
): Observable<ExecutorConnectionFields> => {
    const variables: ExecutorsVariables = {
        query: query ?? null,
        active: active ?? null,
        first: first ?? null,
        after: after ?? null,
    }

    return from(
        client.query<ExecutorsResult>({
            query: getDocumentNode(EXECUTORS),
            variables,
        })
    ).pipe(
        map(({ data }) => data),
        map(({ executors }) => executors)
    )
}
