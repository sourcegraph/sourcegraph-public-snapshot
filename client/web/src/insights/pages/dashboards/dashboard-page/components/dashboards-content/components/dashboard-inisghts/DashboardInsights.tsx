import React from 'react'

import { ExtensionsControllerProps } from '@sourcegraph/shared/src/extensions/controller'
import { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'

import { Settings } from '../../../../../../../../schema/settings.schema'
import { SmartInsightsViewGrid } from '../../../../../../../components/insights-view-grid'
import { InsightDashboard } from '../../../../../../../core/types'
import { useDistinctValue } from '../../../../../../../hooks/use-distinct-value'
import { useInsights } from '../../../../../../../hooks/use-insight/use-insight'
import { EmptyInsightDashboard } from '../empty-insight-dashboard/EmptyInsightDashboard'

import { DashboardInsightsContext } from './DashboardInsightsContext'

const DEFAULT_INSIGHT_IDS: string[] = []

interface DashboardInsightsProps
    extends ExtensionsControllerProps,
        TelemetryProps,
        SettingsCascadeProps<Settings>,
        PlatformContextProps<'updateSettings'> {
    dashboard: InsightDashboard
    onAddInsightRequest: () => void
}

export const DashboardInsights: React.FunctionComponent<DashboardInsightsProps> = props => {
    const {
        telemetryService,
        extensionsController,
        dashboard,
        settingsCascade,
        platformContext,
        onAddInsightRequest,
    } = props

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
                        extensionsController={extensionsController}
                        where="insightsPage"
                        context={{}}
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
