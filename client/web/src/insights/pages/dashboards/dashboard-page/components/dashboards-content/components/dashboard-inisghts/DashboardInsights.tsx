import PuzzleIcon from 'mdi-react/PuzzleIcon'
import React, { useContext, useMemo } from 'react'

import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import { haveInitialExtensionsLoaded } from '@sourcegraph/shared/src/api/features'
import { ExtensionsControllerProps } from '@sourcegraph/shared/src/extensions/controller'
import { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { useObservable } from '@sourcegraph/shared/src/util/useObservable'

import { Settings } from '../../../../../../../../schema/settings.schema'
import { CodeInsightsIcon, InsightsViewGrid } from '../../../../../../../components'
import { InsightsApiContext } from '../../../../../../../core/backend/api-provider'
import { InsightDashboard } from '../../../../../../../core/types'
import { useDistinctValue } from '../../../../../../../hooks/use-distinct-value'
import { EmptyInsightDashboard } from '../empty-insight-dashboard/EmptyInsightDashboard'

import { useBackendInsightIds } from './hooks/use-backend-insight-ids'
import { useDeleteInsight } from './hooks/use-delete-insight'

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
    const { getInsightCombinedViews } = useContext(InsightsApiContext)

    const allInsightIds = useDistinctValue(dashboard.insightIds) ?? DEFAULT_INSIGHT_IDS
    const finalSettings = useDistinctValue(settingsCascade.final)
    const backendInsightIds = useBackendInsightIds({ insightIds: allInsightIds, finalSettings })

    const views = useObservable(
        useMemo(() => getInsightCombinedViews(extensionsController?.extHostAPI, allInsightIds, backendInsightIds), [
            allInsightIds,
            backendInsightIds,
            extensionsController,
            getInsightCombinedViews,
        ])
    )

    const { handleDelete } = useDeleteInsight({ settingsCascade, platformContext })

    // Ensures that we don't show a misleading empty state when extensions haven't loaded yet.
    const areExtensionsReady = useObservable(
        useMemo(() => haveInitialExtensionsLoaded(props.extensionsController.extHostAPI), [props.extensionsController])
    )

    if (!areExtensionsReady) {
        return (
            <div className="d-flex justify-content-center align-items-center pt-5">
                <LoadingSpinner />
                <span className="mx-2">Loading code insights</span>
                <PuzzleIcon className="icon-inline" />
            </div>
        )
    }

    if (views === undefined) {
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
            {allInsightIds.length > 0 && views.length > 0 ? (
                <InsightsViewGrid
                    views={views}
                    hasContextMenu={true}
                    telemetryService={telemetryService}
                    onDelete={handleDelete}
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
