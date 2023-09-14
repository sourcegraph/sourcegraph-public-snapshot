import type { ApolloError } from '@apollo/client'

import { useQuery } from '@sourcegraph/http-client'

import type { VisibleIndexesResult, VisibleIndexesVariables } from '../../../../graphql-operations'
import { visibleIndexesQuery } from '../backend'

export interface UseVisibleIndexesResult {
    data?: {
        id: string
        uploadedAt: string | null
        inputCommit: string
        indexer: { name: string; url: string } | null
    }[]
    error?: ApolloError
    loading: boolean
}

export const useVisibleIndexes = (variables: VisibleIndexesVariables): UseVisibleIndexesResult => {
    const { data, error, loading } = useQuery<VisibleIndexesResult, VisibleIndexesVariables>(visibleIndexesQuery, {
        variables,
        fetchPolicy: 'cache-first',
    })

    const indexes = data?.repository?.commit?.blob?.lsif?.visibleIndexes

    if (!indexes) {
        return { loading, error }
    }

    return {
        data: indexes,
        error,
        loading,
    }
}
