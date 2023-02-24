import { ApolloError } from '@apollo/client'

import { useQuery } from '@sourcegraph/http-client'

import {
    RepoCodeIntelStatusResult,
    RepoCodeIntelStatusVariables,
    InferredAvailableIndexersFields,
    PreciseIndexFields,
} from '../../../../graphql-operations'
import { repoCodeIntelStatusQuery } from '../backend'

export interface UseRepoCodeIntelStatusPayload {
    lastIndexScan?: string
    lastUploadRetentionScan?: string
    availableIndexers: InferredAvailableIndexersFields[]
    recentActivity: PreciseIndexFields[]
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
        fetchPolicy: 'cache-and-network', // TODO: Think about invalidation, especially after fixing
    })

    const repo = rawData?.repository

    if (!repo) {
        return { loading, error }
    }

    return {
        data: {
            availableIndexers: repo.codeIntelSummary.availableIndexers,
            lastIndexScan: repo.codeIntelSummary.lastIndexScan || undefined,
            lastUploadRetentionScan: repo.codeIntelSummary.lastUploadRetentionScan || undefined,
            recentActivity: repo.codeIntelSummary.recentActivity,
        },
        error,
        loading,
    }
}
