import { ApolloClient } from '@apollo/client'
import { from, Observable } from 'rxjs'
import { map } from 'rxjs/operators'

import { gql, getDocumentNode } from '@sourcegraph/shared/src/graphql/graphql'
import * as GQL from '@sourcegraph/shared/src/graphql/schema'

import { ExecutorFields, ExecutorsResult, ExecutorsVariables } from '../../graphql-operations'

export const executorFieldsFragment = gql`
    fragment ExecutorFields on Executor {
        __typename
        id
        hostname
        lastSeenAt
    }
`

interface ExecutorConnection {
    nodes: ExecutorFields[]
    totalCount: number | null
    pageInfo: { endCursor: string | null; hasNextPage: boolean }
}

const EXECUTORS = gql`
    query Executors($first: Int, $after: String) {
        executors(first: $first, after: $after) {
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

export const useExecutors = (
    { first, after }: GQL.IExecutorsOnQueryArguments,
    client: ApolloClient<object>
): Observable<ExecutorConnection> => {
    const vars = {
        first: first ?? null,
        after: after ?? null,
    }

    return from(
        client.query<ExecutorsResult, ExecutorsVariables>({
            query: getDocumentNode(EXECUTORS),
            variables: { ...vars },
        })
    ).pipe(
        map(({ data }) => data),
        map(({ executors }) => executors)
    )
}
