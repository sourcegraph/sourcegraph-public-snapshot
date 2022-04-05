import { ApolloError } from '@apollo/client'

import { gql, useQuery } from '@sourcegraph/http-client'

import {
    CodeIntelStatusResult,
    CodeIntelStatusVariables,
    PreciseSupportFields,
    SearchBasedCodeIntelSupportFields,
} from '../../graphql-operations'

import { lsifIndexFieldsFragment } from './indexes/hooks/types'
import { lsifUploadFieldsFragment } from './uploads/hooks/types'

const codeIntelIndexerFieldsFragment = gql`
    fragment CodeIntelIndexerFields on CodeIntelIndexer {
        name
        url
    }
`

const codeIntelSupportFieldsFragment = gql`
    fragment CodeIntelSupportFields on CodeIntelSupport {
        preciseSupport {
            ...PreciseSupportFields
        }
        searchBasedSupport {
            ...SearchBasedCodeIntelSupportFields
        }
    }
`

const gitTreeCodeIntelInfoFieldsFragment = gql`
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
`

const preciseSupportFieldsFragment = gql`
    fragment PreciseSupportFields on PreciseCodeIntelSupport {
        supportLevel
        indexers {
            ...CodeIntelIndexerFields
        }
    }
`

const searchBasedCodeIntelSupportFieldsFragment = gql`
    fragment SearchBasedCodeIntelSupportFields on SearchBasedCodeIntelSupport {
        language
        supportLevel
    }
`

const codeIntelStatusQuery = gql`
    query CodeIntelStatus($repository: String!, $commit: String!, $path: String!) {
        repository(name: $repository) {
            codeIntelSummary {
                lastIndexScan
                lastUploadRetentionScan
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
            }
            commit(rev: $commit) {
                path(path: $path) {
                    ... on GitBlob {
                        codeIntelSupport {
                            ...CodeIntelSupportFields
                        }
                    }
                    ... on GitTree {
                        codeIntelInfo {
                            ...GitTreeCodeIntelInfoFields
                        }
                    }
                }
            }
        }
    }

    ${lsifUploadFieldsFragment}
    ${lsifIndexFieldsFragment}
    ${codeIntelIndexerFieldsFragment}
    ${codeIntelSupportFieldsFragment}
    ${gitTreeCodeIntelInfoFieldsFragment}
    ${preciseSupportFieldsFragment}
    ${searchBasedCodeIntelSupportFieldsFragment}
`

export interface UseCodeIntelStatusParameters {
    variables: CodeIntelStatusVariables
}

export interface UseCodeIntelStatusResult {
    data?: {
        preciseSupport?: PreciseSupportFields[]
        searchBasedSupport?: SearchBasedCodeIntelSupportFields[]
    }
    error?: ApolloError
    loading: boolean
}

export const useCodeIntelStatus = ({ variables }: UseCodeIntelStatusParameters): UseCodeIntelStatusResult => {
    const { data: rawData, error, loading } = useQuery<CodeIntelStatusResult, CodeIntelStatusVariables>(
        codeIntelStatusQuery,
        {
            variables,
            notifyOnNetworkStatusChange: false,
            fetchPolicy: 'no-cache',
        }
    )

    const path = rawData?.repository?.commit?.path
    switch (path?.__typename) {
        case 'GitBlob': {
            return {
                data: {
                    searchBasedSupport: [path.codeIntelSupport.searchBasedSupport],
                    preciseSupport: [path.codeIntelSupport.preciseSupport],
                },
                error,
                loading,
            }
        }

        case 'GitTree': {
            const info = path.codeIntelInfo
            return {
                data: info
                    ? {
                          searchBasedSupport: info.searchBasedSupport?.map(wrapper => wrapper.support) || [],
                          preciseSupport: info.preciseSupport?.map(wrapper => wrapper.support) || [],
                      }
                    : undefined,
                error,
                loading,
            }
        }

        default:
            return { data: undefined, error, loading }
    }
}
