import React, { useContext, useMemo } from 'react'

import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { useObservable } from '@sourcegraph/shared/src/util/useObservable'
import { LoadingSpinner } from '@sourcegraph/wildcard'

import { SmartInsightsViewGrid } from '../../../../../../../components/insights-view-grid/SmartInsightsViewGrid'
import { CodeInsightsBackendContext } from '../../../../../../../core/backend/code-insights-backend-context'
import { InsightDashboard } from '../../../../../../../core/types'
import { SupportedInsightSubject } from '../../../../../../../core/types/subjects'
import { EmptyInsightDashboard } from '../empty-insight-dashboard/EmptyInsightDashboard'

import { DashboardInsightsContext } from './DashboardInsightsContext'

interface DashboardInsightsProps extends TelemetryProps {
    dashboard: InsightDashboard
    subjects?: SupportedInsightSubject[]
    onAddInsightRequest: () => void
}

export const DashboardInsights: React.FunctionComponent<DashboardInsightsProps> = props => {
    const { telemetryService, dashboard, subjects, onAddInsightRequest } = props

    const { getInsights } = useContext(CodeInsightsBackendContext)

    const insights = useObservable(
        useMemo(() => getInsights({ dashboardId: dashboard.id }), [getInsights, dashboard.id])
    )

    if (insights === undefined) {
        return <LoadingSpinner inline={false} />
    }

    return (
        <DashboardInsightsContext.Provider value={{ dashboard }}>
            <div>
                {insights.length > 0 ? (
                    <SmartInsightsViewGrid insights={insights} telemetryService={telemetryService} />
                ) : (
                    <EmptyInsightDashboard
                        subjects={subjects}
                        dashboard={dashboard}
                        onAddInsight={onAddInsightRequest}
                    />
                )}
            </div>
        </DashboardInsightsContext.Provider>
    )
}
