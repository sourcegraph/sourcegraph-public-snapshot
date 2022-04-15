import React, { useContext, useMemo } from 'react'

import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { LoadingSpinner, useObservable } from '@sourcegraph/wildcard'

import { SmartInsightsViewGrid } from '../../../../../../../components/insights-view-grid/SmartInsightsViewGrid'
import { CodeInsightsBackendContext, InsightDashboard } from '../../../../../../../core'
import { EmptyInsightDashboard } from '../empty-insight-dashboard/EmptyInsightDashboard'

import { DashboardInsightsContext } from './DashboardInsightsContext'

interface DashboardInsightsProps extends TelemetryProps {
    dashboard: InsightDashboard
    onAddInsightRequest: () => void
}

export const DashboardInsights: React.FunctionComponent<DashboardInsightsProps> = props => {
    const { telemetryService, dashboard, onAddInsightRequest } = props

    const { getInsights } = useContext(CodeInsightsBackendContext)

    const insights = useObservable(
        useMemo(() => getInsights({ dashboardId: dashboard.id }), [getInsights, dashboard.id])
    )

    if (insights === undefined) {
        return <LoadingSpinner inline={false} />
    }

    return (
        <DashboardInsightsContext.Provider value={{ dashboard }}>
            {insights.length > 0 ? (
                <SmartInsightsViewGrid insights={insights} telemetryService={telemetryService} />
            ) : (
                <EmptyInsightDashboard dashboard={dashboard} onAddInsight={onAddInsightRequest} />
            )}
        </DashboardInsightsContext.Provider>
    )
}
