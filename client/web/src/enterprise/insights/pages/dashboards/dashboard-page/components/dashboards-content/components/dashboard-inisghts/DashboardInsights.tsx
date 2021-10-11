import React from 'react'

import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'

import { SmartInsightsViewGrid } from '../../../../../../../components/insights-view-grid/SmartInsightsViewGrid'
import { InsightDashboard } from '../../../../../../../core/types'
import { useDistinctValue } from '../../../../../../../hooks/use-distinct-value'

import { EmptyInsightDashboard } from '../empty-insight-dashboard/EmptyInsightDashboard'

import { DashboardInsightsContext } from './DashboardInsightsContext'
import { useObservable } from '@sourcegraph/shared/src/util/useObservable';

const DEFAULT_INSIGHT_IDS: string[] = []

interface DashboardInsightsProps extends TelemetryProps {
    dashboard: InsightDashboard
    onAddInsightRequest: () => void
}

export const DashboardInsights: React.FunctionComponent<DashboardInsightsProps> = props => {
    const { telemetryService, dashboard, onAddInsightRequest } = props

    const dashboardInsightIds = dashboard.insightIds ?? DEFAULT_INSIGHT_IDS
    const insightIds = useDistinctValue(dashboardInsightIds)

    const insights = useObservable() // useInsights({ insightIds, settingsCascade })

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
