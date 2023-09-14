import { gql } from '@sourcegraph/http-client'

export const SEARCH_REPO_META_KEYS_GQL = gql`
    query SearchRepoMetaKeys($query: String) {
        repoMeta {
            keys(query: $query, first: 10) {
                nodes
            }
        }
    }
`

export const SEARCH_REPO_META_VALUES_GQL = gql`
    query SearchRepoMetaValues($key: String!, $query: String) {
        repoMeta {
            key(key: $key) {
                values(query: $query, first: 10) {
                    nodes
                }
            }
        }
    }
`

export const ADD_REPO_METADATA_GQL = gql`
    mutation AddRepoMetadata($repo: ID!, $key: String!, $value: String) {
        addRepoMetadata(repo: $repo, key: $key, value: $value) {
            alwaysNil
        }
    }
`

export const DELETE_REPO_METADATA_GQL = gql`
    mutation DeleteRepoMetadata($repo: ID!, $key: String!) {
        deleteRepoMetadata(repo: $repo, key: $key) {
            alwaysNil
        }
    }
`

export const GET_REPO_METADATA_GQL = gql`
    query GetRepoMetadata($repo: String!) {
        repository(name: $repo) {
            metadata {
                key
                value
            }
        }
    }
`
