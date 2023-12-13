import { useCallback, useContext, useState } from 'react'

import { type ErrorLike, logger } from '@sourcegraph/common'

import { eventLogger } from '../../../tracking/eventLogger'
import { CodeInsightsBackendContext, type Insight } from '../core'
import { getTrackingTypeByInsightType } from '../pings'

type DeletionInsight = Pick<Insight, 'id' | 'type'>

export interface UseDeleteInsightAPI {
    delete: (insight: DeletionInsight) => Promise<void>
    loading: boolean
    error: ErrorLike | undefined
}

/**
 * Returns delete handler that deletes insight from all subject settings and from all dashboards
 * that include this insight.
 */
export function useDeleteInsight(): UseDeleteInsightAPI {
    const { deleteInsight } = useContext(CodeInsightsBackendContext)

    const [loading, setLoading] = useState<boolean>(false)
    const [error, setError] = useState<ErrorLike | undefined>()

    const handleDelete = useCallback(
        async (insight: DeletionInsight) => {
            // Prevent double call if we already have ongoing request
            if (loading) {
                return
            }

            setLoading(true)
            setError(undefined)

            try {
                await deleteInsight(insight.id).toPromise()
                const insightType = getTrackingTypeByInsightType(insight.type)

                window.context.telemetryRecorder?.recordEvent('insightRemoval', 'deleted', {
                    privateMetadata: { insightType },
                })
                eventLogger.log('InsightRemoval', { insightType }, { insightType })
            } catch (error) {
                // TODO [VK] Improve error UI for deleting
                logger.error(error)
                setError(error)
            }
        },
        [loading, deleteInsight, window.context.telemetryRecorder]
    )

    return { delete: handleDelete, loading, error }
}
