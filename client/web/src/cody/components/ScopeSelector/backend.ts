import { gql } from '@sourcegraph/http-client'

export const ReposSelectorSearchQuery = gql`
    query ReposSelectorSearch($query: String!, $includeJobs: Boolean!, $orderBy: RepositoryOrderBy) {
        repositories(query: $query, first: 10, orderBy: $orderBy) {
            nodes {
                id
                name
                embeddingExists
                externalRepository {
                    id
                    serviceType
                }
                embeddingJobs(first: 1) @include(if: $includeJobs) {
                    nodes {
                        id
                        state
                        failureMessage
                    }
                }
            }
        }
    }
`

export const ReposStatusQuery = gql`
    query ReposStatus($repoNames: [String!]!, $first: Int!, $includeJobs: Boolean!) {
        repositories(names: $repoNames, first: $first) {
            nodes {
                id
                name
                embeddingExists
                externalRepository {
                    id
                    serviceType
                }
                embeddingJobs(first: 1) @include(if: $includeJobs) {
                    nodes {
                        id
                        state
                        failureMessage
                    }
                }
            }
        }
    }
`
