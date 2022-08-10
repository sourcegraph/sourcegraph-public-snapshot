import { gql } from '@sourcegraph/http-client'

export const EXECUTOR_FIELD_FRAGMENT = gql`
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

export const EXECUTORS = gql`
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

    ${EXECUTOR_FIELD_FRAGMENT}
`
