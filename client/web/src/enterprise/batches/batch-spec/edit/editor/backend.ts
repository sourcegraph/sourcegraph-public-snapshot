import { gql } from '@sourcegraph/http-client'

export const BATCH_SPEC_EXECUTION_AVAILABLE_SECRET_KEYS = gql`
    query BatchSpecExecutionAvailableSecretKeys($namespace: ID!) {
        node(id: $namespace) {
            __typename
            ... on User {
                executorSecrets(scope: BATCHES, first: 99999) {
                    nodes {
                        key
                    }
                }
            }
            ... on Org {
                executorSecrets(scope: BATCHES, first: 99999) {
                    nodes {
                        key
                    }
                }
            }
        }
    }
`
