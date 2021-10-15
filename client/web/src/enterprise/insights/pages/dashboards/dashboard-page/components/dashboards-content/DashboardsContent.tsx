import classnames from 'classnames'
import MapSearchIcon from 'mdi-react/MapSearchIcon'
import React, { useContext, useMemo, useRef, useState } from 'react'
import { useHistory } from 'react-router-dom'

import { LoadingSpinner } from '@sourcegraph/react-loading-spinner';
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { useObservable } from '@sourcegraph/shared/src/util/useObservable';

import { HeroPage } from '../../../../../../../components/HeroPage'
import { CodeInsightsBackendContext } from '../../../../../core/backend/code-insights-backend-context';
import { isVirtualDashboard } from '../../../../../core/types'
import { isSettingsBasedInsightsDashboard } from '../../../../../core/types/dashboard/real-dashboard'
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
    const { getDashboards, getInsightSubjects } = useContext(CodeInsightsBackendContext)

    const subjects = useObservable(useMemo(() => getInsightSubjects(), [getInsightSubjects]))
    const dashboards = useObservable(useMemo(() => getDashboards(), [getDashboards]))

    // State to open/close add/remove insights modal UI
    const [isAddInsightOpen, setAddInsightsState] = useState<boolean>(false)
    const [isDeleteDashboardActive, setDeleteDashboardActive] = useState<boolean>(false)

    const handleDashboardSelect = useDashboardSelectHandler()
    const [copyURL, isCopied] = useCopyURLHandler()
    const menuReference = useRef<HTMLButtonElement | null>(null)

    const handleSelect = (action: DashboardMenuAction): void => {
        switch (action) {
            case DashboardMenuAction.Configure: {
                if (!isVirtualDashboard(currentDashboard) && isSettingsBasedInsightsDashboard(currentDashboard)) {
                    history.push(`/insights/dashboards/${currentDashboard.settingsKey}/edit`)
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

    if (dashboards === undefined) {
        return <LoadingSpinner />
    }

    const currentDashboard = findDashboardByUrlId(dashboards, dashboardID)

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
                <AddInsightModal
                    dashboard={currentDashboard}
                    onClose={() => setAddInsightsState(false)}
                />
            )}

            {isDeleteDashboardActive && isDashboardConfigurable(currentDashboard) && (
                <DeleteDashboardModal
                    dashboard={currentDashboard}
                    onClose={() => setDeleteDashboardActive(false)}
                />
            )}
        </div>
    )
}
