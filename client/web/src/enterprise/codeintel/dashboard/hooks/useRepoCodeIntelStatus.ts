import { ApolloError } from '@apollo/client'

import { useQuery } from '@sourcegraph/http-client'

import {
    RepoCodeIntelStatusResult,
    RepoCodeIntelStatusVariables,
    InferedPreciseSupportLevel,
    InferredAvailableIndexersFields,
    PreciseIndexFields,
    PreciseSupportFields,
    SearchBasedCodeIntelSupportFields,
} from '../../../../graphql-operations'
import { repoCodeIntelStatusQuery } from '../backend'

export interface UseRepoCodeIntelStatusPayload {
    lastIndexScan?: string
    lastUploadRetentionScan?: string
    availableIndexers: InferredAvailableIndexersFields[]
    recentActivity: PreciseIndexFields[]
    preciseSupport?: (PreciseSupportFields & { confidence?: InferedPreciseSupportLevel })[]
    searchBasedSupport?: SearchBasedCodeIntelSupportFields[]
}

interface UseRepoCodeIntelStatusParameters {
    variables: RepoCodeIntelStatusVariables
}

interface UseRepoCodeIntelStatusResult {
    data?: UseRepoCodeIntelStatusPayload
    error?: ApolloError
    loading: boolean
}

export const useRepoCodeIntelStatus = ({
    variables,
}: UseRepoCodeIntelStatusParameters): UseRepoCodeIntelStatusResult => {
    const {
        data: rawData,
        error,
        loading,
    } = useQuery<RepoCodeIntelStatusResult, RepoCodeIntelStatusVariables>(repoCodeIntelStatusQuery, {
        variables,
        notifyOnNetworkStatusChange: false,
        fetchPolicy: 'cache-first', // TODO: Think about invalidation, especially after fixing
    })

    const repo = rawData?.repository
    const path = repo?.commit?.path

    if (!repo || !path) {
        return { loading, error }
    }

    const summary = repo.codeIntelSummary
    const common: Omit<UseRepoCodeIntelStatusPayload, 'preciseSupport' | 'searchBasedSupport'> = {
        availableIndexers: summary.availableIndexers,
        lastIndexScan: summary.lastIndexScan || undefined,
        lastUploadRetentionScan: summary.lastUploadRetentionScan || undefined,
        recentActivity: summary.recentActivity,
    }

    switch (path?.__typename) {
        case 'GitTree': {
            const info = path.codeIntelInfo
            return {
                data: info
                    ? {
                          ...common,
                          searchBasedSupport: (info.searchBasedSupport || []).map(wrapper => wrapper.support),
                          preciseSupport: (info.preciseSupport?.coverage || []).map(wrapper => ({
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
