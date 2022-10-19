import { gql } from '@sourcegraph/http-client'

export const BATCH_SPEC_EXECUTION_AVAILABLE_SECRET_KEYS = gql`
    query UserBatchSpecExecutionAvailableSecretKeys($user: ID!) {
        node(id: $user) {
            __typename
            ... on User {
                executorSecrets(scope: BATCHES, first: 1000) {
                    nodes {
                        key
                    }
                }
            }
        }
    }
`
