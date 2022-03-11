import React from 'react'
import { useRouteMatch } from 'react-router'
import { Redirect } from 'react-router-dom'

import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'

import { ALL_INSIGHTS_DASHBOARD_ID } from '../../../core/types/dashboard/virtual-dashboard'

import { DashboardsContent } from './components/dashboards-content/DashboardsContent'

export interface DashboardsContentPageProps extends TelemetryProps {
    /**
     * Possible dashboard id. All insights on the page will be get from
     * dashboard's info from the user or org settings by the dashboard id.
     * In case if id is undefined we get insights from the final
     * version of merged settings (all insights)
     */
    dashboardID?: string
}

export const DashboardsContentPage: React.FunctionComponent<DashboardsContentPageProps> = props => {
    const { dashboardID, telemetryService } = props
    const { url } = useRouteMatch()

    if (!dashboardID) {
        // In case if url doesn't have a dashboard id we should fall back on
        // built-in "All insights" dashboard
        return <Redirect to={`${url}/${ALL_INSIGHTS_DASHBOARD_ID}`} />
    }

    return <DashboardsContent telemetryService={telemetryService} dashboardID={dashboardID} />
}
