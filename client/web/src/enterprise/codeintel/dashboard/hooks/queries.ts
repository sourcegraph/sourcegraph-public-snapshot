import { gql } from '@sourcegraph/http-client'
import { codeIntelIndexerFieldsFragment, preciseIndexFieldsFragment } from '../../indexes/hooks/types'

export const globalCodeIntelStatusQuery = gql`
    query GlobalCodeIntelStatus {
        codeIntelSummary {
            numRepositoriesWithCodeIntelligence
            repositoriesWithErrors {
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
            repositoriesWithConfiguration {
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
            commit(rev: "HEAD") {
                path(path: "/") {
                    ... on GitTree {
                        codeIntelInfo {
                            ...GitTreeCodeIntelInfoFields
                        }
                    }
                }
            }
        }
    }

    fragment GitTreeCodeIntelInfoFields on GitTreeCodeIntelInfo {
        preciseSupport {
            support {
                ...PreciseSupportFields
            }
            confidence
        }
        searchBasedSupport {
            support {
                ...SearchBasedCodeIntelSupportFields
            }
        }
    }

    fragment PreciseSupportFields on PreciseCodeIntelSupport {
        supportLevel
        indexers {
            ...CodeIntelIndexerFields
        }
    }

    fragment SearchBasedCodeIntelSupportFields on SearchBasedCodeIntelSupport {
        language
        supportLevel
    }

    ${preciseIndexFieldsFragment}
    ${inferredAvailableIndexersFieldsFragment}
`
