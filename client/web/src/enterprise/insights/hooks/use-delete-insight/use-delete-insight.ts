import { useCallback, useContext, useState } from 'react'

import { ErrorLike } from '@sourcegraph/shared/src/util/errors'

import { InsightsApiContext } from '../../core/backend/api-provider'
import { Insight } from '../../core/types'

export interface UseDeleteInsightAPI {
    delete: (insight: Pick<Insight, 'id' | 'title'>) => Promise<void>
    loading: boolean
    error: ErrorLike | undefined
}

/**
 * Returns delete handler that deletes insight from all subject settings and from all dashboards
 * that include this insight.
 */
export function useDeleteInsight(): UseDeleteInsightAPI {
    const { deleteInsight } = useContext(InsightsApiContext)

    const [loading, setLoading] = useState<boolean>(false)
    const [error, setError] = useState<ErrorLike | undefined>()

    const handleDelete = useCallback(
        async (insight: Pick<Insight, 'id' | 'title'>) => {
            const shouldDelete = window.confirm(`Are you sure you want to delete the insight "${insight.title}"?`)

            // Prevent double call if we already have ongoing request
            if (loading || !shouldDelete) {
                return
            }

            setLoading(true)
            setError(undefined)

            try {
                await deleteInsight(insight.id).toPromise()
            } catch (error) {
                // TODO [VK] Improve error UI for deleting
                console.error(error)
                setError(error)
            }

            setLoading(false)
        },
        [deleteInsight, loading]
    )

    return { delete: handleDelete, loading, error }
}
