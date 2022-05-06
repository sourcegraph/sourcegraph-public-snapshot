import React, { useContext, useMemo } from 'react'

import { useRouteMatch } from 'react-router'
import { Redirect } from 'react-router-dom'

import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { LoadingSpinner, useObservable } from '@sourcegraph/wildcard'

import { PageTitle } from '../../../../../components/PageTitle'
import { CodeInsightsBackendContext, ALL_INSIGHTS_DASHBOARD } from '../../../core'

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

export const DashboardsContentPage: React.FunctionComponent<
    React.PropsWithChildren<DashboardsContentPageProps>
> = props => {
    const { dashboardID, telemetryService } = props
    const { url } = useRouteMatch()

    const { getDashboards } = useContext(CodeInsightsBackendContext)
    const dashboards = useObservable(useMemo(() => getDashboards(), [getDashboards]))

    if (!dashboardID) {
        // In case if url doesn't have a dashboard id we should fall back on
        // built-in "All insights" dashboard
        return <Redirect to={`${url}/${ALL_INSIGHTS_DASHBOARD.id}`} />
    }

    if (dashboards === undefined) {
        return (
            <div data-testid="loading-spinner">
                <LoadingSpinner inline={false} />
            </div>
        )
    }

    const currentDashboard = dashboards.find(dashboard => dashboard.id === dashboardID)

    return (
        <>
            <PageTitle title={`${currentDashboard?.title || ''} - Code Insights`} />
            <DashboardsContent telemetryService={telemetryService} dashboardID={dashboardID} dashboards={dashboards} />
        </>
    )
}
