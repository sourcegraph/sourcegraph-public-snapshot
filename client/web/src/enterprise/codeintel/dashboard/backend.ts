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

export const gitTreeCodeIntelInfoFragment = gql`
    fragment GitTreeCodeIntelInfoFields on GitTreeCodeIntelInfo {
        preciseSupport {
            coverage {
                support {
                    ...PreciseSupportFields
                }
                confidence
            }
            limitError
        }
        searchBasedSupport {
            support {
                ...SearchBasedCodeIntelSupportFields
            }
        }
    }
`

export const preciseSupportFragment = gql`
    fragment PreciseSupportFields on PreciseCodeIntelSupport {
        supportLevel
        indexers {
            ...CodeIntelIndexerFields
        }
    }
`

export const searchBasedCodeIntelSupportFragment = gql`
    fragment SearchBasedCodeIntelSupportFields on SearchBasedCodeIntelSupport {
        language
        supportLevel
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
