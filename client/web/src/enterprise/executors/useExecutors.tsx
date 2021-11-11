import { ApolloError } from '@apollo/client'

import { gql, useQuery } from '@sourcegraph/shared/src/graphql/graphql'

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
    executors: ExecutorFields[]
    loading: boolean
    error: ApolloError | undefined
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

export const useExecutors = (): ExecutorConnection => {
    const { data, loading, error } = useQuery<ExecutorsResult, ExecutorsVariables>(EXECUTORS, {
        variables: {
            first: null, // TODO
            after: null, // TODO
        },
    })

    return {
        executors: (data?.executors.nodes || []).map(({ id, hostname, lastSeenAt }) => ({
            __typename: 'Executor',
            id,
            hostname,
            lastSeenAt,
        })),
        loading,
        error,
    }
}
