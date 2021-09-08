import React from 'react'

import { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'

import { Settings } from '../../../../../../../../schema/settings.schema'
import { SmartInsightsViewGrid } from '../../../../../../../components/insights-view-grid/SmartInsightsViewGrid'
import { InsightDashboard } from '../../../../../../../core/types'
import { useDistinctValue } from '../../../../../../../hooks/use-distinct-value'
import { useInsights } from '../../../../../../../hooks/use-insight/use-insight'
import { EmptyInsightDashboard } from '../empty-insight-dashboard/EmptyInsightDashboard'

import { DashboardInsightsContext } from './DashboardInsightsContext'

const DEFAULT_INSIGHT_IDS: string[] = []

interface DashboardInsightsProps
    extends TelemetryProps,
        SettingsCascadeProps<Settings>,
        PlatformContextProps<'updateSettings'> {
    dashboard: InsightDashboard
    onAddInsightRequest: () => void
}

export const DashboardInsights: React.FunctionComponent<DashboardInsightsProps> = props => {
    const { telemetryService, dashboard, settingsCascade, platformContext, onAddInsightRequest } = props

    const dashboardInsightIds = dashboard.insightIds ?? DEFAULT_INSIGHT_IDS
    const insightIds = useDistinctValue(dashboardInsightIds)
    const insights = useInsights({ insightIds, settingsCascade })

    return (
        <DashboardInsightsContext.Provider value={{ dashboard }}>
            <div>
                {insights.length > 0 ? (
                    <SmartInsightsViewGrid
                        insights={insights}
                        telemetryService={telemetryService}
                        settingsCascade={settingsCascade}
                        platformContext={platformContext}
                    />
                ) : (
                    <EmptyInsightDashboard
                        dashboard={dashboard}
                        settingsCascade={settingsCascade}
                        onAddInsight={onAddInsightRequest}
                    />
                )}
            </div>
        </DashboardInsightsContext.Provider>
    )
}
