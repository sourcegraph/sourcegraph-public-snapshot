import type { ApolloError } from '@apollo/client'

import { useQuery } from '@sourcegraph/http-client'

import type {
    RepoCodeIntelStatusResult,
    RepoCodeIntelStatusVariables,
    PreciseIndexFields,
    RepoCodeIntelStatusSummaryFields,
    RepoCodeIntelStatusCommitGraphFields,
    InferredAvailableIndexersWithKeysFields,
} from '../../../../graphql-operations'
import { repoCodeIntelStatusQuery } from '../backend'

export interface UseRepoCodeIntelStatusPayload {
    lastIndexScan?: string
    lastUploadRetentionScan?: string
    availableIndexers: InferredAvailableIndexersWithKeysFields[]
    recentActivity: PreciseIndexFields[]
}

interface UseRepoCodeIntelStatusResult {
    data?: {
        summary: RepoCodeIntelStatusSummaryFields
        commitGraph: RepoCodeIntelStatusCommitGraphFields
    }
    error?: ApolloError
    loading: boolean
}

const POLL_INTERVAL = 1000 * 60 * 5 // 5 minutes

export const useRepoCodeIntelStatus = (variables: RepoCodeIntelStatusVariables): UseRepoCodeIntelStatusResult => {
    const { data, error, loading } = useQuery<RepoCodeIntelStatusResult, RepoCodeIntelStatusVariables>(
        repoCodeIntelStatusQuery,
        {
            variables,
            fetchPolicy: 'cache-and-network', // Cache when loaded, refetch in background
            pollInterval: POLL_INTERVAL, // Refetch every 5 minutes (if left open)
        }
    )

    const repo = data?.repository

    if (!repo) {
        return { loading, error }
    }

    return {
        data: {
            summary: repo.codeIntelSummary,
            commitGraph: repo.codeIntelligenceCommitGraph,
        },
        error,
        loading,
    }
}
