import classNames from 'classnames'
import MapSearchIcon from 'mdi-react/MapSearchIcon'
import React, { useContext, useMemo, useRef, useState } from 'react'
import { useHistory } from 'react-router-dom'

import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { useObservable } from '@sourcegraph/shared/src/util/useObservable'
import { authenticatedUser } from '@sourcegraph/web/src/auth'
import { LoadingSpinner } from '@sourcegraph/wildcard'

import { HeroPage } from '../../../../../../../components/HeroPage'
import { CodeInsightsBackendContext } from '../../../../../core/backend/code-insights-backend-context'
import { isVirtualDashboard } from '../../../../../core/types'
import { isCustomInsightDashboard } from '../../../../../core/types/dashboard/real-dashboard'
import { AddInsightModal } from '../add-insight-modal/AddInsightModal'
import { DashboardMenu, DashboardMenuAction } from '../dashboard-menu/DashboardMenu'
import { DashboardSelect } from '../dashboard-select/DashboardSelect'
import { DeleteDashboardModal } from '../delete-dashboard-modal/DeleteDashboardModal'

import { DashboardInsights } from './components/dashboard-inisghts/DashboardInsights'
import styles from './DashboardsContent.module.scss'
import { useCopyURLHandler } from './hooks/use-copy-url-handler'
import { useDashboardSelectHandler } from './hooks/use-dashboard-select-handler'
import { findDashboardByUrlId } from './utils/find-dashboard-by-url-id'
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
    const { getDashboards, getDashboardSubjects } = useContext(CodeInsightsBackendContext)

    const subjects = useObservable(useMemo(() => getDashboardSubjects(), [getDashboardSubjects]))
    const dashboards = useObservable(useMemo(() => getDashboards(), [getDashboards]))

    // State to open/close add/remove insights modal UI
    const [isAddInsightOpen, setAddInsightsState] = useState<boolean>(false)
    const [isDeleteDashboardActive, setDeleteDashboardActive] = useState<boolean>(false)

    const handleDashboardSelect = useDashboardSelectHandler()
    const [copyURL, isCopied] = useCopyURLHandler()
    const menuReference = useRef<HTMLButtonElement | null>(null)

    const user = useObservable(authenticatedUser)

    if (dashboards === undefined) {
        return (
            <div data-testid="loading-spinner">
                <LoadingSpinner inline={false} />
            </div>
        )
    }

    const currentDashboard = findDashboardByUrlId(dashboards, dashboardID)

    const handleSelect = (action: DashboardMenuAction): void => {
        switch (action) {
            case DashboardMenuAction.Configure: {
                if (
                    currentDashboard &&
                    !isVirtualDashboard(currentDashboard) &&
                    isCustomInsightDashboard(currentDashboard)
                ) {
                    const dashboardURL = currentDashboard.settingsKey ?? currentDashboard.id

                    history.push(`/insights/dashboards/${dashboardURL}/edit`)
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
            <section className="d-flex flex-wrap align-items-center">
                <span className={styles.dashboardSelectLabel}>Dashboard</span>

                <DashboardSelect
                    value={currentDashboard?.id}
                    dashboards={dashboards}
                    onSelect={handleDashboardSelect}
                    className={classNames(styles.dashboardSelect, 'mr-2')}
                    user={user}
                />

                <DashboardMenu
                    subjects={subjects}
                    innerRef={menuReference}
                    tooltipText={isCopied ? 'Copied!' : undefined}
                    dashboard={currentDashboard}
                    onSelect={handleSelect}
                />
            </section>

            <hr className="mt-2 mb-3" />

            {currentDashboard ? (
                <DashboardInsights
                    subjects={subjects}
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
