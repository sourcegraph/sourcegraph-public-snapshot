import { gql } from '@sourcegraph/http-client'

import { lsifIndexFieldsFragment } from '../../indexes/hooks/types'
import { lsifUploadFieldsFragment } from '../../uploads/hooks/types'

export const codeIntelStatusQuery = gql`
    query CodeIntelStatus($repository: String!, $commit: String!, $path: String!) {
        repository(name: $repository) {
            codeIntelSummary {
                lastIndexScan
                lastUploadRetentionScan
                recentUploads {
                    ...LSIFUploadsWithRepositoryNamespaceFields
                }
                recentIndexes {
                    ...LSIFIndexesWithRepositoryNamespaceFields
                }
            }
            commit(rev: $commit) {
                path(path: $path) {
                    ... on GitBlob {
                        codeIntelSupport {
                            ...CodeIntelSupportFields
                        }
                        lsif {
                            lsifUploads {
                                ...LsifUploadFields
                            }
                        }
                    }
                    ... on GitTree {
                        codeIntelInfo {
                            ...GitTreeCodeIntelInfoFields
                        }
                        lsif {
                            lsifUploads {
                                ...LsifUploadFields
                            }
                        }
                    }
                }
            }
        }
    }

    fragment CodeIntelSupportFields on CodeIntelSupport {
        preciseSupport {
            ...PreciseSupportFields
        }
        searchBasedSupport {
            ...SearchBasedCodeIntelSupportFields
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

    fragment CodeIntelIndexerFields on CodeIntelIndexer {
        name
        url
    }

    fragment LSIFUploadsWithRepositoryNamespaceFields on LSIFUploadsWithRepositoryNamespace {
        root
        indexer {
            ...CodeIntelIndexerFields
        }
        uploads {
            ...LsifUploadFields
        }
    }

    fragment LSIFIndexesWithRepositoryNamespaceFields on LSIFIndexesWithRepositoryNamespace {
        root
        indexer {
            ...CodeIntelIndexerFields
        }
        indexes {
            ...LsifIndexFields
        }
    }

    ${lsifUploadFieldsFragment}
    ${lsifIndexFieldsFragment}
`

export const requestedLanguageSupportQuery = gql`
    query RequestedLanguageSupport {
        requestedLanguageSupport
    }
`

export const requestLanguageSupportQuery = gql`
    mutation RequestLanguageSupport($language: String!) {
        requestLanguageSupport(language: $language) {
            alwaysNil
        }
    }
`
