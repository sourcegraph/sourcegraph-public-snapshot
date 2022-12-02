import React, { useEffect, useState } from 'react'

import classNames from 'classnames'
import MapSearchIcon from 'mdi-react/MapSearchIcon'
import { useHistory } from 'react-router-dom'

import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { Button, Tooltip } from '@sourcegraph/wildcard'

import { HeroPage } from '../../../../../../../components/HeroPage'
import { LimitedAccessLabel } from '../../../../../components'
import { CustomInsightDashboard } from '../../../../../core'
import { useCopyURLHandler, useUiFeatures } from '../../../../../hooks'
import { AddInsightModal } from '../add-insight-modal'
import { DashboardMenu, DashboardMenuAction } from '../dashboard-menu/DashboardMenu'
import { DashboardSelect } from '../dashboard-select/DashboardSelect'
import { DeleteDashboardModal } from '../delete-dashboard-modal/DeleteDashboardModal'

import { DashboardHeader } from './components/dashboard-header/DashboardHeader'
import { DashboardInsights } from './components/dashboard-inisghts/DashboardInsights'

import styles from './DashboardsContent.module.scss'

export interface DashboardsContentProps extends TelemetryProps {
    /**
     * Possible dashboard id. All insights on the page will be get from
     * dashboard's info from the user or org settings by the dashboard id.
     * In case if id is undefined we get insights from the final
     * version of merged settings (all insights)
     */
    currentDashboard?: CustomInsightDashboard
    dashboards: CustomInsightDashboard[]
}

export const DashboardsContent: React.FunctionComponent<React.PropsWithChildren<DashboardsContentProps>> = props => {
    const { currentDashboard, dashboards, telemetryService } = props

    const history = useHistory()
    const { dashboard: dashboardPermission, licensed } = useUiFeatures()

    const [copyURL, isCopied] = useCopyURLHandler()
    const [isAddInsightOpen, setAddInsightsState] = useState<boolean>(false)
    const [isDeleteDashboardActive, setDeleteDashboardActive] = useState<boolean>(false)

    useEffect(() => {
        telemetryService.logViewEvent('Insights')
    }, [telemetryService])

    const handleDashboardSelect = (dashboard: CustomInsightDashboard): void =>
        history.push(`/insights/dashboards/${dashboard.id}`)

    const handleSelect = (action: DashboardMenuAction): void => {
        switch (action) {
            case DashboardMenuAction.Configure: {
                if (currentDashboard) {
                    history.push(`/insights/dashboards/${currentDashboard.id}/edit`)
                }
                return
            }
            case DashboardMenuAction.Delete: {
                setDeleteDashboardActive(true)
                return
            }
            case DashboardMenuAction.CopyLink: {
                copyURL()
                return
            }
        }
    }

    const addRemovePermissions = dashboardPermission.getAddRemoveInsightsPermission(currentDashboard)

    return (
        <div className={styles.root}>
            <DashboardHeader className={styles.header}>
                <span className={styles.dashboardSelectLabel}>Dashboard:</span>

                <DashboardSelect
                    dashboard={currentDashboard}
                    dashboards={dashboards}
                    className={styles.dashboardSelect}
                    onSelect={handleDashboardSelect}
                />

                <DashboardMenu
                    dashboard={currentDashboard}
                    tooltipText={isCopied ? 'Copied!' : undefined}
                    className={styles.dashboardMenu}
                    onSelect={handleSelect}
                />

                <Tooltip content={addRemovePermissions.tooltip} placement="bottom">
                    <Button
                        outline={true}
                        variant="secondary"
                        disabled={addRemovePermissions.disabled}
                        onClick={() => setAddInsightsState(true)}
                    >
                        Add or remove insights
                    </Button>
                </Tooltip>
            </DashboardHeader>

            {!licensed && (
                <LimitedAccessLabel
                    className={classNames(styles.limitedAccessLabel)}
                    message="Unlock Code Insights for full access to custom dashboards"
                />
            )}

            {currentDashboard ? (
                <DashboardInsights
                    currentDashboard={currentDashboard}
                    telemetryService={telemetryService}
                    className={styles.insights}
                    onAddInsightRequest={() => setAddInsightsState(true)}
                />
            ) : (
                <HeroPage icon={MapSearchIcon} title="Hmm, the dashboard wasn't found." />
            )}

            {isAddInsightOpen && currentDashboard && (
                <AddInsightModal dashboard={currentDashboard} onClose={() => setAddInsightsState(false)} />
            )}

            {isDeleteDashboardActive && currentDashboard && (
                <DeleteDashboardModal dashboard={currentDashboard} onClose={() => setDeleteDashboardActive(false)} />
            )}
        </div>
    )
}
