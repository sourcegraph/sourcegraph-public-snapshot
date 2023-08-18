import { gql } from '@sourcegraph/http-client'

const REPO_FIELDS = gql`
    fragment ContextSelectorRepoFields on Repository {
        id
        name
        embeddingExists
        externalRepository {
            id
            serviceType
        }
    }
`

const EMBEDDING_JOB_FIELDS = gql`
    fragment ContextSelectorEmbeddingJobFields on RepoEmbeddingJob {
        id
        state
        failureMessage
    }
`

export const ReposSelectorSearchQuery = gql`
    query ReposSelectorSearch($query: String!, $includeJobs: Boolean!) {
        repositories(query: $query, first: 10) {
            nodes {
                ...ContextSelectorRepoFields
                embeddingJobs(first: 1) @include(if: $includeJobs) {
                    nodes {
                        ...ContextSelectorEmbeddingJobFields
                    }
                }
            }
        }
    }

    ${REPO_FIELDS}
    ${EMBEDDING_JOB_FIELDS}
`
`

export const ReposStatusQuery = gql`
    query ReposStatus($repoNames: [String!]!, $first: Int!, $includeJobs: Boolean!) {
        repositories(names: $repoNames, first: $first) {
            nodes {
                ...ContextSelectorRepoFields
                embeddingJobs(first: 1) @include(if: $includeJobs) {
                    nodes {
                        ...ContextSelectorEmbeddingJobFields
                    }
                }
            }
        }
    }

    ${REPO_FIELDS}
    ${EMBEDDING_JOB_FIELDS}
`
