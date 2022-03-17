import React, { useEffect, useRef, useState } from 'react'

import classNames from 'classnames'
import MapSearchIcon from 'mdi-react/MapSearchIcon'
import { useHistory } from 'react-router-dom'

import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { authenticatedUser } from '@sourcegraph/web/src/auth'
import { Button, useObservable } from '@sourcegraph/wildcard'

import { HeroPage } from '../../../../../../../components/HeroPage'
import { LimitedAccessLabel } from '../../../../../components/limited-access-label/LimitedAccessLabel'
import { InsightDashboard, isVirtualDashboard } from '../../../../../core/types'
import { isCustomInsightDashboard } from '../../../../../core/types/dashboard/real-dashboard'
import { ALL_INSIGHTS_DASHBOARD_ID } from '../../../../../core/types/dashboard/virtual-dashboard'
import { useUiFeatures } from '../../../../../hooks/use-ui-features'
import { AddInsightModal } from '../add-insight-modal/AddInsightModal'
import { DashboardMenu, DashboardMenuAction } from '../dashboard-menu/DashboardMenu'
import { DashboardSelect } from '../dashboard-select/DashboardSelect'
import { DeleteDashboardModal } from '../delete-dashboard-modal/DeleteDashboardModal'

import { DashboardHeader } from './components/dashboard-header/DashboardHeader'
import { DashboardInsights } from './components/dashboard-inisghts/DashboardInsights'
import { useCopyURLHandler } from './hooks/use-copy-url-handler'
import { isDashboardConfigurable } from './utils/is-dashboard-configurable'

import styles from './DashboardsContent.module.scss'

export interface DashboardsContentProps extends TelemetryProps {
    /**
     * Possible dashboard id. All insights on the page will be get from
     * dashboard's info from the user or org settings by the dashboard id.
     * In case if id is undefined we get insights from the final
     * version of merged settings (all insights)
     */
    dashboardID: string
    dashboards: InsightDashboard[]
}

export const DashboardsContent: React.FunctionComponent<DashboardsContentProps> = props => {
    const { dashboardID, telemetryService, dashboards } = props
    const currentDashboard = dashboards.find(dashboard => dashboard.id === dashboardID)

    const history = useHistory()
    const { dashboard: dashboardPermission, licensed } = useUiFeatures()

    // State to open/close add/remove insights modal UI
    const [isAddInsightOpen, setAddInsightsState] = useState<boolean>(false)
    const [isDeleteDashboardActive, setDeleteDashboardActive] = useState<boolean>(false)

    const handleDashboardSelect = (dashboard: InsightDashboard): void =>
        history.push(`/insights/dashboards/${dashboard.id}`)

    const [copyURL, isCopied] = useCopyURLHandler()
    const menuReference = useRef<HTMLButtonElement | null>(null)

    const user = useObservable(authenticatedUser)

    useEffect(() => {
        telemetryService.logViewEvent('Insights')
    }, [telemetryService, dashboardID])

    const handleSelect = (action: DashboardMenuAction): void => {
        switch (action) {
            case DashboardMenuAction.Configure: {
                if (
                    currentDashboard &&
                    !isVirtualDashboard(currentDashboard) &&
                    isCustomInsightDashboard(currentDashboard)
                ) {
                    history.push(`/insights/dashboards/${currentDashboard.id}/edit`)
                }
                return
            }
            case DashboardMenuAction.AddRemoveInsights: {
                setAddInsightsState(true)
                return
            }
            case DashboardMenuAction.Delete: {
                setDeleteDashboardActive(true)
                return
            }
            case DashboardMenuAction.CopyLink: {
                copyURL()

                // Re-trigger trigger tooltip event catching logic to activate
                // copied tooltip appearance
                requestAnimationFrame(() => {
                    menuReference.current?.blur()
                    menuReference.current?.focus()
                })

                return
            }
        }
    }

    const handleAddInsightRequest = (): void => {
        setAddInsightsState(true)
    }

    const addRemovePermissions = dashboardPermission.getAddRemoveInsightsPermission(currentDashboard)

    return (
        <main className="pb-4">
            <DashboardHeader className="d-flex flex-wrap align-items-center mb-3">
                <span className={styles.dashboardSelectLabel}>Dashboard:</span>

                <DashboardSelect
                    value={currentDashboard?.id}
                    dashboards={dashboards}
                    onSelect={handleDashboardSelect}
                    className={classNames(styles.dashboardSelect, 'mr-2')}
                    user={user}
                />

                <DashboardMenu
                    innerRef={menuReference}
                    tooltipText={isCopied ? 'Copied!' : undefined}
                    dashboard={currentDashboard}
                    onSelect={handleSelect}
                    className="mr-auto"
                />

                <Button
                    outline={true}
                    variant="secondary"
                    disabled={addRemovePermissions.disabled}
                    data-tooltip={addRemovePermissions.tooltip}
                    data-placement="bottom"
                    onClick={() => handleSelect(DashboardMenuAction.AddRemoveInsights)}
                >
                    Add or remove insights
                </Button>
            </DashboardHeader>

            {!licensed && (
                <LimitedAccessLabel
                    className={classNames(styles.limitedAccessLabel)}
                    message={
                        dashboardID === ALL_INSIGHTS_DASHBOARD_ID
                            ? 'Create up to two global insights'
                            : 'Unlock Code Insights for full access to custom dashboards'
                    }
                />
            )}

            {currentDashboard ? (
                <DashboardInsights
                    dashboard={currentDashboard}
                    telemetryService={telemetryService}
                    onAddInsightRequest={handleAddInsightRequest}
                />
            ) : (
                <HeroPage icon={MapSearchIcon} title="Hmm, the dashboard wasn't found." />
            )}

            {isAddInsightOpen && isDashboardConfigurable(currentDashboard) && (
                <AddInsightModal dashboard={currentDashboard} onClose={() => setAddInsightsState(false)} />
            )}

            {isDeleteDashboardActive && isDashboardConfigurable(currentDashboard) && (
                <DeleteDashboardModal dashboard={currentDashboard} onClose={() => setDeleteDashboardActive(false)} />
            )}
        </main>
    )
}
