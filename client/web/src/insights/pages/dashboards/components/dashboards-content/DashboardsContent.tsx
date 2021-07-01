import MapSearchIcon from 'mdi-react/MapSearchIcon'
import React, { useContext, useMemo } from 'react'
import { useHistory } from 'react-router-dom'

import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import { ExtensionsControllerProps } from '@sourcegraph/shared/src/extensions/controller'
import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { useObservable } from '@sourcegraph/shared/src/util/useObservable'

import { HeroPage } from '../../../../../components/HeroPage'
import { Settings } from '../../../../../schema/settings.schema'
import { InsightsViewGrid, InsightsViewGridProps } from '../../../../components'
import { InsightsApiContext } from '../../../../core/backend/api-provider'
import { InsightDashboard, isVirtualDashboard } from '../../../../core/types'
import { useDashboards } from '../../hooks/use-dashboards/use-dashboards'
import { DashboardSelect } from '../dashboard-select/DashboardSelect'

import styles from './DashboardsContent.module.scss'

export interface DashboardsContentProps
    extends Omit<InsightsViewGridProps, 'views' | 'settingsCascade'>,
        SettingsCascadeProps<Settings>,
        ExtensionsControllerProps {
    /**
     * Possible dashboard id. All insights on the page will be get from
     * dashboard's info from the user or org settings by the dashboard id.
     * In case if id is undefined we get insights from the final
     * version of merged settings (all insights)
     */
    dashboardID: string
}

export const DashboardsContent: React.FunctionComponent<DashboardsContentProps> = props => {
    const { settingsCascade, dashboardID } = props

    const history = useHistory()
    const dashboards = useDashboards(settingsCascade)

    const currentDashboard = useMemo(
        () =>
            dashboards.find(dashboard => {
                if (isVirtualDashboard(dashboard)) {
                    return (
                        dashboard.id === dashboardID.toLowerCase() ||
                        dashboard.type.toLowerCase() === dashboardID.toLowerCase()
                    )
                }

                return dashboard.id === dashboardID || dashboard.title.toLowerCase() === dashboardID?.toLowerCase()
            }),
        [dashboardID, dashboards]
    )

    const handleDashboardSelect = (dashboard: InsightDashboard): void => {
        history.push(`/insights/dashboard/${dashboard.id}`)
    }

    return (
        <div>
            <DashboardSelect
                value={currentDashboard?.id}
                dashboards={dashboards}
                onSelect={handleDashboardSelect}
                className={styles.dashboardSelect}
            />

            <hr className="mt-2 mb-3" />

            {currentDashboard ? (
                <DashboardInsights {...props} insightIds={currentDashboard.insightIds} />
            ) : (
                <HeroPage icon={MapSearchIcon} title="Hmm dashboard wasn't found." />
            )}
        </div>
    )
}

// This strange props extending here (child props interface is extended by parent props)
// is needed for the InsightsViewGrid component and related to '/views' specific component usage.
// This problem will be resolved here https://github.com/sourcegraph/sourcegraph/issues/22462
interface DashboardInsightsProps extends DashboardsContentProps {
    /**
     * Dashboard specific insight ids.
     */
    insightIds?: string[]
}

/**
 * Renders code insight view grid.
 */
const DashboardInsights: React.FunctionComponent<DashboardInsightsProps> = props => {
    const { extensionsController, insightIds } = props
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
                <InsightsViewGrid {...props} views={views} hasContextMenu={true} />
            )}
        </div>
    )
}
