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

// todo: recent uploads?
export const repoCodeIntelStatusQuery = gql`
    query RepoCodeIntelStatus($repository: String!) {
        repository(name: $repository) {
            codeIntelSummary {
                lastIndexScan
                lastUploadRetentionScan
                recentActivity {
                    ...PreciseIndexFields
                }
                availableIndexers {
                    ...InferredAvailableIndexersFields
                }
            }
        }
    }

    ${preciseIndexFieldsFragment}
    ${inferredAvailableIndexersFieldsFragment}
`
