import { gql } from '@sourcegraph/http-client'

export const STATUS_AND_REPO_STATS = gql`
    query StatusAndRepoStats {
        repositoryStats {
            __typename
            total
            notCloned
            cloned
            cloning
            failedFetch
            corrupted
            indexed
        }
        statusMessages {
            ... on GitUpdatesDisabled {
                __typename

                message
            }

            ... on CloningProgress {
                __typename

                message
            }

            ... on IndexingProgress {
                __typename

                notIndexed
                indexed
            }

            ... on SyncError {
                __typename

                message
            }

            ... on ExternalServiceSyncError {
                __typename

                externalService {
                    id
                    displayName
                }
            }
        }
    }
`
