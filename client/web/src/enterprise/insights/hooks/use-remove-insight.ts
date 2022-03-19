import { useCallback, useContext, useState } from 'react'

import { ErrorLike } from '@sourcegraph/common'

import { eventLogger } from '../../../tracking/eventLogger'
import { CodeInsightsBackendContext } from '../core/backend/code-insights-backend-context'
import { Insight, InsightDashboard } from '../core/types'
import { getTrackingTypeByInsightType } from '../pings'

interface RemoveInsightInput {
    insight: Pick<Insight, 'id' | 'title' | 'type'>
    dashboard: Pick<InsightDashboard, 'id' | 'title'>
}

export interface useRemoveInsightFromDashboardAPI {
    remove: (insight: RemoveInsightInput) => Promise<void>
    loading: boolean
    error: ErrorLike | undefined
}

export function useRemoveInsightFromDashboard(): useRemoveInsightFromDashboardAPI {
    const { removeInsightFromDashboard } = useContext(CodeInsightsBackendContext)

    const [loading, setLoading] = useState<boolean>(false)
    const [error, setError] = useState<ErrorLike | undefined>()

    const handleRemove = useCallback(
        async (input: RemoveInsightInput) => {
            const { insight, dashboard } = input
            const shouldRemove = window.confirm(
                `Are you sure you want to remove the insight "${insight.title}" from the dashboard "${dashboard.title}"?`
            )

            // Prevent double call if we already have ongoing request
            if (loading || !shouldRemove) {
                return
            }

            setLoading(true)
            setError(undefined)

            try {
                await removeInsightFromDashboard({
                    insightId: insight.id,
                    dashboardId: dashboard.id,
                }).toPromise()

                const insightType = getTrackingTypeByInsightType(insight.type)

                eventLogger.log('InsightRemovalFromDashboard', { insightType }, { insightType })
            } catch (error) {
                // TODO [VK] Improve error UI for removing
                console.error(error)
                setError(error)
            }
        },
        [loading, removeInsightFromDashboard]
    )

    return { remove: handleRemove, loading, error }
}
