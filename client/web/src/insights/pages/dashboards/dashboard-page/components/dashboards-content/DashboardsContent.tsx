import classnames from 'classnames'
import MapSearchIcon from 'mdi-react/MapSearchIcon'
import React, { useContext, useMemo, useState } from 'react'
import { useHistory } from 'react-router-dom'

import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import { ExtensionsControllerProps } from '@sourcegraph/shared/src/extensions/controller'
import { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { useObservable } from '@sourcegraph/shared/src/util/useObservable'

import { HeroPage } from '../../../../../../components/HeroPage'
import { Settings } from '../../../../../../schema/settings.schema'
import { InsightsViewGrid } from '../../../../../components'
import { InsightsApiContext } from '../../../../../core/backend/api-provider'
import { InsightDashboard, isRealDashboard, isVirtualDashboard } from '../../../../../core/types'
import { isSettingsBasedInsightsDashboard } from '../../../../../core/types/dashboard/real-dashboard'
import { useDashboards } from '../../../../../hooks/use-dashboards/use-dashboards'
import { AddInsightModal } from '../add-insight-modal/AddInsightModal'
import { DashboardMenu, DashboardMenuAction } from '../dashboard-menu/DashboardMenu'
import { DashboardSelect } from '../dashboard-select/DashboardSelect'

import styles from './DashboardsContent.module.scss'

export interface DashboardsContentProps
    extends SettingsCascadeProps<Settings>,
        ExtensionsControllerProps,
        TelemetryProps,
        PlatformContextProps<'updateSettings'> {
    /**
     * Possible dashboard id. All insights on the page will be get from
     * dashboard's info from the user or org settings by the dashboard id.
     * In case if id is undefined we get insights from the final
     * version of merged settings (all insights)
     */
    dashboardID: string
}

export const DashboardsContent: React.FunctionComponent<DashboardsContentProps> = props => {
    const { extensionsController, settingsCascade, dashboardID, telemetryService, platformContext } = props

    const history = useHistory()
    const dashboards = useDashboards(settingsCascade)

    // State to open/close add/remove insights modal UI
    const [isAddInsightOpen, setAddInsightsState] = useState<boolean>(false)

    const currentDashboard = dashboards.find(dashboard => {
        if (isVirtualDashboard(dashboard)) {
            return (
                dashboard.id === dashboardID.toLowerCase() || dashboard.type.toLowerCase() === dashboardID.toLowerCase()
            )
        }

        return (
            dashboard.id === dashboardID ||
            dashboard.title.toLowerCase() === dashboardID?.toLowerCase() ||
            (isSettingsBasedInsightsDashboard(dashboard) &&
                dashboard.settingsKey.toLowerCase() === dashboardID?.toLowerCase())
        )
    })

    const handleDashboardSelect = (dashboard: InsightDashboard): void => {
        if (isVirtualDashboard(dashboard)) {
            history.push(`/insights/dashboards/${dashboard.type}`)

            return
        }

        if (isSettingsBasedInsightsDashboard(dashboard)) {
            history.push(`/insights/dashboards/${dashboard.settingsKey}`)

            return
        }

        history.push(`/insights/dashboards/${dashboard.id}`)
    }

    const handleSelect = (action: DashboardMenuAction): void => {
        switch (action) {
            case DashboardMenuAction.Configure: {
                if (
                    currentDashboard &&
                    !isVirtualDashboard(currentDashboard) &&
                    isSettingsBasedInsightsDashboard(currentDashboard)
                ) {
                    history.push(`/insights/dashboards/${currentDashboard.settingsKey}/edit`)
                }

                return
            }

            case DashboardMenuAction.AddRemoveInsights: {
                setAddInsightsState(true)
            }

            // Implement other actions
        }
    }

    return (
        <div>
            <section className="d-flex flex-wrap align-items-center">
                <span className={styles.dashboardSelectLabel}>Dashboard</span>

                <DashboardSelect
                    value={currentDashboard?.id}
                    dashboards={dashboards}
                    onSelect={handleDashboardSelect}
                    className={classnames(styles.dashboardSelect, 'mr-2')}
                />

                <DashboardMenu dashboard={currentDashboard} onSelect={handleSelect} />
            </section>

            <hr className="mt-2 mb-3" />

            {currentDashboard ? (
                <DashboardInsights
                    insightIds={currentDashboard.insightIds}
                    extensionsController={extensionsController}
                    telemetryService={telemetryService}
                />
            ) : (
                <HeroPage icon={MapSearchIcon} title="Hmm, the dashboard wasn't found." />
            )}

            {isAddInsightOpen &&
                currentDashboard &&
                isRealDashboard(currentDashboard) &&
                isSettingsBasedInsightsDashboard(currentDashboard) && (
                    <AddInsightModal
                        platformContext={platformContext}
                        settingsCascade={settingsCascade}
                        dashboard={currentDashboard}
                        onClose={() => setAddInsightsState(false)}
                    />
                )}
        </div>
    )
}

interface DashboardInsightsProps extends ExtensionsControllerProps, TelemetryProps {
    /**
     * Dashboard specific insight ids.
     */
    insightIds?: string[]
}

/**
 * Renders code insight view grid.
 */
const DashboardInsights: React.FunctionComponent<DashboardInsightsProps> = props => {
    const { telemetryService, extensionsController, insightIds } = props
    const { getInsightCombinedViews } = useContext(InsightsApiContext)

    const views = useObservable(
        useMemo(() => getInsightCombinedViews(extensionsController?.extHostAPI, insightIds), [
            insightIds,
            extensionsController,
            getInsightCombinedViews,
        ])
    )

    return (
        <div>
            {views === undefined ? (
                <div className="d-flex w-100">
                    <LoadingSpinner className="my-4" />
                </div>
            ) : (
                <InsightsViewGrid views={views} hasContextMenu={true} telemetryService={telemetryService} />
            )}
        </div>
    )
}
