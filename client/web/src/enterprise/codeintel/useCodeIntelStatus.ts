import { ApolloError } from '@apollo/client'

import { useQuery } from '@sourcegraph/http-client'

import {
    CodeIntelStatusResult,
    CodeIntelStatusVariables,
    InferedPreciseSupportLevel,
    LSIFIndexesWithRepositoryNamespaceFields,
    LsifUploadFields,
    LSIFUploadsWithRepositoryNamespaceFields,
    PreciseSupportFields,
    SearchBasedCodeIntelSupportFields,
} from '../../graphql-operations'

import { codeIntelStatusQuery } from './queries'

export interface UseCodeIntelStatusParameters {
    variables: CodeIntelStatusVariables
}

export interface UseCodeIntelStatusResult {
    data?: UseCodeIntelStatusPayload
    error?: ApolloError
    loading: boolean
}

export interface UseCodeIntelStatusPayload {
    lastIndexScan?: string
    lastUploadRetentionScan?: string
    activeUploads: LsifUploadFields[]
    recentUploads: LSIFUploadsWithRepositoryNamespaceFields[]
    recentIndexes: LSIFIndexesWithRepositoryNamespaceFields[]
    preciseSupport: (PreciseSupportFields & { confidence?: InferedPreciseSupportLevel })[]
    searchBasedSupport: SearchBasedCodeIntelSupportFields[]
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

    const repo = rawData?.repository
    const path = repo?.commit?.path
    const lsif = path?.lsif
    if (!repo || !path) {
        return { loading, error }
    }

    const summary = repo.codeIntelSummary
    const common: Omit<UseCodeIntelStatusPayload, 'preciseSupport' | 'searchBasedSupport'> = {
        lastIndexScan: summary.lastIndexScan || undefined,
        lastUploadRetentionScan: summary.lastUploadRetentionScan || undefined,
        activeUploads: lsif?.lsifUploads || [],
        recentUploads: summary.recentUploads,
        recentIndexes: summary.recentIndexes,
    }

    switch (path?.__typename) {
        case 'GitBlob': {
            const support = path.codeIntelSupport
            return {
                data: {
                    ...common,
                    searchBasedSupport: [support.searchBasedSupport],
                    preciseSupport: [support.preciseSupport],
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
                          ...common,
                          searchBasedSupport: (info.searchBasedSupport || []).map(wrapper => wrapper.support),
                          preciseSupport: (info.preciseSupport || []).map(wrapper => ({
                              ...wrapper.support,
                              confidence: wrapper.confidence,
                          })),
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
