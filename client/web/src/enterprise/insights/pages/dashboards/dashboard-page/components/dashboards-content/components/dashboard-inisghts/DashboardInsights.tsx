import React, { useContext, useMemo } from 'react'

import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { LoadingSpinner, useObservable } from '@sourcegraph/wildcard'

import { SmartInsightsViewGrid, InsightContext } from '../../../../../../../components'
import { CodeInsightsBackendContext, InsightDashboard } from '../../../../../../../core'
import { EmptyInsightDashboard } from '../empty-insight-dashboard/EmptyInsightDashboard'

interface DashboardInsightsProps extends TelemetryProps {
    currentDashboard: InsightDashboard
    dashboards: InsightDashboard[]
    className?: string
    onAddInsightRequest: () => void
}

export const DashboardInsights: React.FunctionComponent<React.PropsWithChildren<DashboardInsightsProps>> = props => {
    const { telemetryService, currentDashboard, dashboards, className, onAddInsightRequest } = props

    const { getInsights } = useContext(CodeInsightsBackendContext)

    const insights = useObservable(
        useMemo(() => getInsights({ dashboardId: currentDashboard.id }), [getInsights, currentDashboard.id])
    )

    const insightContextValue = useMemo(() => ({ currentDashboard, dashboards }), [currentDashboard, dashboards])

    if (insights === undefined) {
        return <LoadingSpinner inline={false} />
    }

    return (
        <InsightContext.Provider value={insightContextValue}>
            {insights.length > 0 ? (
                <SmartInsightsViewGrid insights={insights} telemetryService={telemetryService} className={className} />
            ) : (
                <EmptyInsightDashboard dashboard={currentDashboard} onAddInsight={onAddInsightRequest} />
            )}
        </InsightContext.Provider>
    )
}
