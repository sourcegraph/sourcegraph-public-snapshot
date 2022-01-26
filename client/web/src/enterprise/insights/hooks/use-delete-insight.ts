import { useCallback, useContext, useState } from 'react'

import { ErrorLike } from '@sourcegraph/common'
import { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'

import { eventLogger } from '../../../tracking/eventLogger'
import { CodeInsightsBackendContext } from '../core/backend/code-insights-backend-context'
import { Insight } from '../core/types'

export interface UseDeleteInsightProps extends SettingsCascadeProps, PlatformContextProps<'updateSettings'> {}

export interface UseDeleteInsightAPI {
    delete: (insight: Pick<Insight, 'id' | 'title' | 'type'>) => Promise<void>
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
        async (insight: Pick<Insight, 'id' | 'title' | 'type'>) => {
            const shouldDelete = window.confirm(`Are you sure you want to delete the insight "${insight.title}"?`)

            // Prevent double call if we already have ongoing request
            if (loading || !shouldDelete) {
                return
            }

            setLoading(true)
            setError(undefined)

            try {
                await deleteInsight(insight.id).toPromise()
                eventLogger.log('InsightRemoval', { insightType: insight.type }, { insightType: insight.type })
            } catch (error) {
                // TODO [VK] Improve error UI for deleting
                console.error(error)
                setError(error)
            }
        },
        [loading, deleteInsight]
    )

    return { delete: handleDelete, loading, error }
}
