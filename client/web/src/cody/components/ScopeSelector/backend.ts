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

export const RecentReposSelectorQuery = gql`
    query RecentReposSelector(
        $name0: String!
        $name1: String!
        $name2: String!
        $name3: String!
        $name4: String!
        $name5: String!
        $name6: String!
        $name7: String!
        $name8: String!
        $name9: String!
        $includeJobs: Boolean!
    ) {
        repo0: repository(name: $name0) {
            ...ContextSelectorRepoFields
            embeddingJobs(first: 1) @include(if: $includeJobs) {
                nodes {
                    ...ContextSelectorEmbeddingJobFields
                }
            }
        }
        repo1: repository(name: $name1) {
            ...ContextSelectorRepoFields
            embeddingJobs(first: 1) @include(if: $includeJobs) {
                nodes {
                    ...ContextSelectorEmbeddingJobFields
                }
            }
        }
        repo2: repository(name: $name2) {
            ...ContextSelectorRepoFields
            embeddingJobs(first: 1) @include(if: $includeJobs) {
                nodes {
                    ...ContextSelectorEmbeddingJobFields
                }
            }
        }
        repo3: repository(name: $name3) {
            ...ContextSelectorRepoFields
            embeddingJobs(first: 1) @include(if: $includeJobs) {
                nodes {
                    ...ContextSelectorEmbeddingJobFields
                }
            }
        }
        repo4: repository(name: $name4) {
            ...ContextSelectorRepoFields
            embeddingJobs(first: 1) @include(if: $includeJobs) {
                nodes {
                    ...ContextSelectorEmbeddingJobFields
                }
            }
        }
        repo5: repository(name: $name5) {
            ...ContextSelectorRepoFields
            embeddingJobs(first: 1) @include(if: $includeJobs) {
                nodes {
                    ...ContextSelectorEmbeddingJobFields
                }
            }
        }
        repo6: repository(name: $name6) {
            ...ContextSelectorRepoFields
            embeddingJobs(first: 1) @include(if: $includeJobs) {
                nodes {
                    ...ContextSelectorEmbeddingJobFields
                }
            }
        }
        repo7: repository(name: $name7) {
            ...ContextSelectorRepoFields
            embeddingJobs(first: 1) @include(if: $includeJobs) {
                nodes {
                    ...ContextSelectorEmbeddingJobFields
                }
            }
        }
        repo8: repository(name: $name8) {
            ...ContextSelectorRepoFields
            embeddingJobs(first: 1) @include(if: $includeJobs) {
                nodes {
                    ...ContextSelectorEmbeddingJobFields
                }
            }
        }
        repo9: repository(name: $name9) {
            ...ContextSelectorRepoFields
            embeddingJobs(first: 1) @include(if: $includeJobs) {
                nodes {
                    ...ContextSelectorEmbeddingJobFields
                }
            }
        }
    }

    ${REPO_FIELDS}
    ${EMBEDDING_JOB_FIELDS}
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
