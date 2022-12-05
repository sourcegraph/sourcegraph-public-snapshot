import { FC } from 'react'

import { useRouteMatch } from 'react-router'
import { Redirect } from 'react-router-dom'

import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { LoadingSpinner } from '@sourcegraph/wildcard'

import { PageTitle } from '../../../../../components/PageTitle'
import { isPersonalDashboard, useInsightDashboards } from '../../../core'

import { DashboardsContent } from './components/dashboards-content/DashboardsContent'

export interface DashboardsViewProps extends TelemetryProps {
    /**
     * Possible dashboard id. All insights on the page will be got from
     * dashboard's info from the user or org settings by the dashboard id.
     * In case if id is undefined we get insights from the final
     * version of merged settings (all insights)
     */
    dashboardId?: string
}

export const DashboardsView: FC<DashboardsViewProps> = props => {
    const { dashboardId, telemetryService } = props

    const { url } = useRouteMatch()
    const { dashboards } = useInsightDashboards()

    if (!dashboards) {
        return <LoadingSpinner aria-live="off" inline={false} />
    }

    // User doesn't have any dashboard, render empty (zero dashboard state) screen
    if (dashboards.length === 0) {
        // TODO: Improve dashboards empty state UI
        return <span>You don't have any code insight dashboard.</span>
    }

    // URL doesn't have any concrete dashboard id, render first dashboard and add
    // first dashboard ID to match URL (/dashboards/ -> /dashboards/<first dashboard id>)
    if (!dashboardId) {
        const currentDashboard = dashboards.find(isPersonalDashboard) ?? dashboards[0]

        return <Redirect push={false} to={`${url}/${currentDashboard.id}`} />
    }

    // We have dashboards and dashboard id in URL, try to find a current dashboard
    // and render dashboard content, if we can't find current dashboard DashboardsContent
    // will render dashboard not found state.
    const currentDashboard = dashboards.find(dashboard => dashboard.id === dashboardId)

    return (
        <>
            <PageTitle title={`${currentDashboard?.title || ''} - Code Insights`} />
            <DashboardsContent
                currentDashboard={currentDashboard}
                dashboards={dashboards}
                telemetryService={telemetryService}
            />
        </>
    )
}
