import classNames from 'classnames'
import MapSearchIcon from 'mdi-react/MapSearchIcon'
import React, { useContext, useEffect, useMemo, useRef, useState } from 'react'
import { useHistory } from 'react-router-dom'

import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { authenticatedUser } from '@sourcegraph/web/src/auth'
import { Button, LoadingSpinner, useObservable } from '@sourcegraph/wildcard'

import { HeroPage } from '../../../../../../../components/HeroPage'
import { CodeInsightsBackendContext } from '../../../../../core/backend/code-insights-backend-context'
import { InsightDashboard, isVirtualDashboard } from '../../../../../core/types'
import { isCustomInsightDashboard } from '../../../../../core/types/dashboard/real-dashboard'
import { getTooltipMessage, getDashboardPermissions } from '../../utils/get-dashboard-permissions'
import { AddInsightModal } from '../add-insight-modal/AddInsightModal'
import { DashboardMenu, DashboardMenuAction } from '../dashboard-menu/DashboardMenu'
import { DashboardSelect } from '../dashboard-select/DashboardSelect'
import { DeleteDashboardModal } from '../delete-dashboard-modal/DeleteDashboardModal'

import { DashboardHeader } from './components/dashboard-header/DashboardHeader'
import { DashboardInsights } from './components/dashboard-inisghts/DashboardInsights'
import styles from './DashboardsContent.module.scss'
import { useCopyURLHandler } from './hooks/use-copy-url-handler'
import { isDashboardConfigurable } from './utils/is-dashboard-configurable'

export interface DashboardsContentProps extends TelemetryProps {
    /**
     * Possible dashboard id. All insights on the page will be get from
     * dashboard's info from the user or org settings by the dashboard id.
     * In case if id is undefined we get insights from the final
     * version of merged settings (all insights)
     */
    dashboardID: string
}

export const DashboardsContent: React.FunctionComponent<DashboardsContentProps> = props => {
    const { dashboardID, telemetryService } = props

    const history = useHistory()
    const { getDashboards } = useContext(CodeInsightsBackendContext)

    const dashboards = useObservable(useMemo(() => getDashboards(), [getDashboards]))

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

    if (dashboards === undefined) {
        return (
            <div data-testid="loading-spinner">
                <LoadingSpinner inline={false} />
            </div>
        )
    }

    const currentDashboard = dashboards.find(dashboard => dashboard.id === dashboardID)
    const permissions = getDashboardPermissions(currentDashboard)

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

    return (
        <div>
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
                    disabled={!permissions.isConfigurable}
                    data-tooltip={getTooltipMessage(currentDashboard, permissions)}
                    data-placement="bottom"
                    onClick={() => handleSelect(DashboardMenuAction.AddRemoveInsights)}
                >
                    Add or remove insights
                </Button>
            </DashboardHeader>

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
        </div>
    )
}
