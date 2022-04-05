import { ApolloError } from '@apollo/client'

import { gql, useQuery } from '@sourcegraph/http-client'

import {
    TreeAndBlobCodeIntelStatusVariables,
    TreeAndBlobCodeIntelStatusResult,
    PreciseSupportLevel,
    InferedPreciseSupportLevel,
    SearchBasedSupportLevel,
    LsifUploadFields,
    LsifIndexFields,
    CodeIntelIndexerFields,
} from '../../graphql-operations'

import { lsifIndexFieldsFragment } from './indexes/hooks/types'
import { lsifUploadFieldsFragment } from './uploads/hooks/types'

const codeIntelRepositorySummaryFragment = gql`
    fragment CodeIntelRepositorySummaryFields on CodeIntelRepositorySummary {
        __typename
        recentUploads {
            root
            indexer {
                ...CodeIntelIndexerFields
            }
            uploads {
                ...LsifUploadFields
            }
        }
        recentIndexes {
            root
            indexer {
                ...CodeIntelIndexerFields
            }
            indexes {
                ...LsifIndexFields
            }
        }
        lastUploadRetentionScan
        lastIndexScan
    }
`

const gitTreeCodeIntelInfoFragment = gql`
    fragment GitTreeCodeIntelInfoFields on GitTreeCodeIntelInfo {
        searchBasedSupport {
            support {
                supportLevel
                language
            }
        }
        preciseSupport {
            support {
                supportLevel
                indexers {
                    ...CodeIntelIndexerFields
                }
            }
            confidence
        }
    }
`

const codeIntelSupportFragment = gql`
    fragment CodeIntelSupportFields on CodeIntelSupport {
        searchBasedSupport {
            supportLevel
            language
        }
        preciseSupport {
            supportLevel
            indexers {
                ...CodeIntelIndexerFields
            }
        }
    }
`

const codeIntelIndexerFragment = gql`
    fragment CodeIntelIndexerFields on CodeIntelIndexer {
        name
        url
    }
`

const BLOB_AND_TREE_CODE_INTEL_STATUS_QUERY = gql`
    query TreeAndBlobCodeIntelStatus($repository: String!, $commit: String!, $path: String!) {
        repository(name: $repository) {
            codeIntelSummary {
                ...CodeIntelRepositorySummaryFields
            }
            commit(rev: $commit) {
                tree(path: $path) {
                    codeIntelInfo {
                        ...GitTreeCodeIntelInfoFields
                    }
                    lsif {
                        lsifUploads {
                            ...LsifUploadFields
                        }
                    }
                }
                blob(path: $path) {
                    codeIntelSupport {
                        ...CodeIntelSupportFields
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

    ${codeIntelRepositorySummaryFragment}
    ${gitTreeCodeIntelInfoFragment}
    ${codeIntelSupportFragment}
    ${codeIntelIndexerFragment}
    ${lsifUploadFieldsFragment}
    ${lsifIndexFieldsFragment}
`

interface UseCodeIntelStatusParameters {
    variables: TreeAndBlobCodeIntelStatusVariables
}

interface UseCodeIntelStatusResult {
    data?: {
        searchBasedSupport: {
            supportLevel: SearchBasedSupportLevel
            language?: string
        }[]
        preciseSupport: {
            supportLevel: PreciseSupportLevel
            indexers?: CodeIntelIndexerFields[]
            confidence?: InferedPreciseSupportLevel
        }[]
        uploads: LsifUploadFields[]

        // repository.codeIntelSummary
        recentUploads: { root: string; indexer: CodeIntelIndexerFields; uploads: LsifUploadFields[] }[]
        recentIndexes: { root: string; indexer: CodeIntelIndexerFields; indexes: LsifIndexFields[] }[]
        lastUploadRetentionScan: string | null
        lastIndexScan: string | null
    }
    error?: ApolloError
    loading: boolean
}

export const useCodeIntelStatus = ({ variables }: UseCodeIntelStatusParameters): UseCodeIntelStatusResult => {
    const { data: rawData, error, loading } = useQuery<
        TreeAndBlobCodeIntelStatusResult,
        TreeAndBlobCodeIntelStatusVariables
    >(BLOB_AND_TREE_CODE_INTEL_STATUS_QUERY, {
        variables,
        notifyOnNetworkStatusChange: false,
        fetchPolicy: 'no-cache',
        errorPolicy: 'ignore', // TODO - necessary because tree OR blob will fail
    })

    if (!rawData?.repository) {
        return { data: undefined, error, loading }
    }

    const tree = rawData.repository.commit?.tree
    if (tree) {
        return {
            data: {
                searchBasedSupport:
                    tree.codeIntelInfo?.searchBasedSupport?.map(support => ({
                        supportLevel: support.support.supportLevel,
                        language: support.support.language || undefined,
                    })) || [],
                preciseSupport:
                    tree.codeIntelInfo?.preciseSupport?.map(support => ({
                        supportLevel: support.support.supportLevel,
                        indexers: support.support.indexers || [],
                        confidence: support.confidence,
                    })) || [],
                uploads: tree.lsif?.lsifUploads || [],
                ...rawData.repository.codeIntelSummary,
            },
            error,
            loading,
        }
    }

    const blob = rawData.repository.commit?.blob
    if (blob) {
        return {
            data: {
                searchBasedSupport: [
                    {
                        supportLevel: blob.codeIntelSupport.searchBasedSupport.supportLevel,
                        language: blob.codeIntelSupport.searchBasedSupport.language || undefined,
                    },
                ],
                preciseSupport: [
                    {
                        supportLevel: blob.codeIntelSupport.preciseSupport.supportLevel,
                        indexers: blob.codeIntelSupport.preciseSupport.indexers || undefined,
                    },
                ],
                uploads: blob.lsif?.lsifUploads || [],
                ...rawData.repository.codeIntelSummary,
            },
            error,
            loading,
        }
    }

    return { data: undefined, error, loading }
}
