import { gql } from '@sourcegraph/http-client'

import { codeIntelIndexerFieldsFragment, preciseIndexFieldsFragment } from '../indexes/hooks/types'

export const dashboardRepoFieldsFragment = gql`
    fragment DashboardRepoFields on CodeIntelRepository {
        name
        url
        externalRepository {
            serviceID
            serviceType
        }
    }
`

export const globalCodeIntelStatusQuery = gql`
    query GlobalCodeIntelStatus {
        codeIntelSummary {
            numRepositoriesWithCodeIntelligence
            repositoriesWithErrors {
                nodes {
                    repository {
                        id
                        name
                        url
                        externalRepository {
                            serviceID
                            serviceType
                        }
                    }
                    count
                }
            }
            repositoriesWithConfiguration {
                nodes {
                    repository {
                        id
                        name
                        url
                        externalRepository {
                            serviceID
                            serviceType
                        }
                    }
                    indexers {
                        indexer {
                            ...CodeIntelIndexerFields
                        }
                        count
                    }
                }
            }
        }
    }

    ${codeIntelIndexerFieldsFragment}
`

export const inferredAvailableIndexersFieldsFragment = gql`
    fragment InferredAvailableIndexersFields on InferredAvailableIndexers {
        indexer {
            ...CodeIntelIndexerFields
        }
        roots
    }
`

export const inferredAvailableIndexersWithKeysFieldsFragment = gql`
    fragment InferredAvailableIndexersWithKeysFields on InferredAvailableIndexers {
        indexer {
            ...CodeIntelIndexerFields
        }
        rootsWithKeys {
            root
            comparisonKey
        }
    }
`

export const repoCodeIntelStatusCommitGraphFragment = gql`
    fragment RepoCodeIntelStatusCommitGraphFields on CodeIntelligenceCommitGraph {
        stale
        updatedAt
    }
`

export const repoCodeIntelStatusSummaryFragment = gql`
    fragment RepoCodeIntelStatusSummaryFields on CodeIntelRepositorySummary {
        lastIndexScan
        lastUploadRetentionScan
        recentActivity {
            ...PreciseIndexFields
        }
        availableIndexers {
            ...InferredAvailableIndexersWithKeysFields
        }
    }

    ${preciseIndexFieldsFragment}
    ${inferredAvailableIndexersWithKeysFieldsFragment}
`

export const repoCodeIntelStatusQuery = gql`
    query RepoCodeIntelStatus($repository: String!) {
        repository(name: $repository) {
            id
            codeIntelSummary {
                ...RepoCodeIntelStatusSummaryFields
            }
            codeIntelligenceCommitGraph {
                ...RepoCodeIntelStatusCommitGraphFields
            }
        }
    }

    ${repoCodeIntelStatusSummaryFragment}
    ${repoCodeIntelStatusCommitGraphFragment}
`

export const visibleIndexesQuery = gql`
    query VisibleIndexes($repository: String!, $commit: String!, $path: String!) {
        repository(name: $repository) {
            id
            commit(rev: $commit) {
                blob(path: $path) {
                    lsif {
                        visibleIndexes {
                            id
                            uploadedAt
                            inputCommit
                            indexer {
                                name
                                url
                            }
                        }
                    }
                }
            }
        }
    }
`
