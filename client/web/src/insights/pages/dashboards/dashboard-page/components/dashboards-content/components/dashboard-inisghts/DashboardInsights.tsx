import React, { useMemo } from 'react'

import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import { haveInitialExtensionsLoaded } from '@sourcegraph/shared/src/api/features'
import { ExtensionsControllerProps } from '@sourcegraph/shared/src/extensions/controller'
import { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { useObservable } from '@sourcegraph/shared/src/util/useObservable'

import { Settings } from '../../../../../../../../schema/settings.schema'
import { CodeInsightsIcon } from '../../../../../../../components'
import { SmartInsightsViewGrid } from '../../../../../../../components/insights-view-grid'
import { InsightDashboard } from '../../../../../../../core/types'
import { useDistinctValue } from '../../../../../../../hooks/use-distinct-value'
import { useInsights } from '../../../../../../../hooks/use-insight/use-insight'
import { EmptyInsightDashboard } from '../empty-insight-dashboard/EmptyInsightDashboard'

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

    // Ensures that we don't show a misleading empty state when extensions haven't loaded yet.
    const areExtensionsReady = useObservable(
        useMemo(() => haveInitialExtensionsLoaded(props.extensionsController.extHostAPI), [props.extensionsController])
    )

    if (!areExtensionsReady) {
        return (
            <div className="d-flex justify-content-center align-items-center pt-5">
                <LoadingSpinner />
                <span className="mx-2">Loading code insights</span>
                <CodeInsightsIcon className="icon-inline" />
            </div>
        )
    }

    return (
        <div>
            {insightIds.length > 0 ? (
                <SmartInsightsViewGrid
                    insights={insights}
                    telemetryService={telemetryService}
                    settingsCascade={settingsCascade}
                    platformContext={platformContext}
                    extensionsController={extensionsController}
                />
            ) : (
                <EmptyInsightDashboard
                    dashboard={dashboard}
                    settingsCascade={settingsCascade}
                    onAddInsight={onAddInsightRequest}
                />
            )}
        </div>
    )
}
