import React, { useContext, useMemo } from 'react'

import { Settings } from '@sourcegraph/shared/src/schema/settings.schema'
import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { LoadingSpinner, useObservable } from '@sourcegraph/wildcard'

import { SmartInsightsViewGrid, InsightContext } from '../../../../../../../components'
import { CodeInsightsBackendContext, InsightDashboard } from '../../../../../../../core'
import { EmptyInsightDashboard } from '../empty-insight-dashboard/EmptyInsightDashboard'

interface DashboardInsightsProps extends TelemetryProps, SettingsCascadeProps<Settings> {
    dashboard: InsightDashboard
    className?: string
    onAddInsightRequest: () => void
}

export const DashboardInsights: React.FunctionComponent<DashboardInsightsProps> = props => {
    const { telemetryService, dashboard, className, onAddInsightRequest } = props

    const { getInsights } = useContext(CodeInsightsBackendContext)

    const insights = useObservable(
        useMemo(() => getInsights({ dashboardId: dashboard.id }), [getInsights, dashboard.id])
    )

    if (insights === undefined) {
        return <LoadingSpinner inline={false} />
    }

    return (
        <InsightContext.Provider value={{ dashboard }}>
            {insights.length > 0 ? (
                <SmartInsightsViewGrid
                    settingsCascade={props.settingsCascade}
                    insights={insights}
                    telemetryService={telemetryService}
                    className={className}
                />
            ) : (
                <EmptyInsightDashboard dashboard={dashboard} onAddInsight={onAddInsightRequest} />
            )}
        </InsightContext.Provider>
    )
}
