import React, { useContext, useMemo } from 'react'

import { LoadingSpinner } from '@sourcegraph/react-loading-spinner';
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { useObservable } from '@sourcegraph/shared/src/util/useObservable';

import { SmartInsightsViewGrid } from '../../../../../../../components/insights-view-grid/SmartInsightsViewGrid'
import { InsightsApiContext } from '../../../../../../../core/backend/api-provider';
import { InsightDashboard } from '../../../../../../../core/types'
import { useDistinctValue } from '../../../../../../../hooks/use-distinct-value'
import { EmptyInsightDashboard } from '../empty-insight-dashboard/EmptyInsightDashboard'

import { DashboardInsightsContext } from './DashboardInsightsContext'

const DEFAULT_INSIGHT_IDS: string[] = []

interface DashboardInsightsProps extends TelemetryProps {
    dashboard: InsightDashboard
    onAddInsightRequest: () => void
}

export const DashboardInsights: React.FunctionComponent<DashboardInsightsProps> = props => {
    const { telemetryService, dashboard, onAddInsightRequest } = props

    const dashboardInsightIds = dashboard.insightIds ?? DEFAULT_INSIGHT_IDS

    const { getInsights } = useContext(InsightsApiContext)
    const insightIds = useDistinctValue(dashboardInsightIds)

    const insights = useObservable(
        useMemo(() => getInsights(insightIds), [getInsights, insightIds])
    )

    if (insights === undefined) {
        return (
            <LoadingSpinner/>
        )
    }

    return (
        <DashboardInsightsContext.Provider value={{ dashboard }}>
            <div>
                {insights.length > 0 ? (
                    <SmartInsightsViewGrid
                        insights={insights}
                        telemetryService={telemetryService}
                    />
                ) : (
                    <EmptyInsightDashboard
                        dashboard={dashboard}
                        onAddInsight={onAddInsightRequest}
                    />
                )}
            </div>
        </DashboardInsightsContext.Provider>
    )
}
