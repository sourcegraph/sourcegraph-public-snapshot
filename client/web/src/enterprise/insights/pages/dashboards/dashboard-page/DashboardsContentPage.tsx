import { FC, useMemo } from 'react'

import { useRouteMatch } from 'react-router'
import { Redirect } from 'react-router-dom'

import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { LoadingSpinner } from '@sourcegraph/wildcard'

import { PageTitle } from '../../../../../components/PageTitle'
import { ALL_INSIGHTS_DASHBOARD } from '../../../constants'
import { useInsightDashboards } from '../../../core'

import { DashboardsContent } from './components/dashboards-content/DashboardsContent'

export interface DashboardsContentPageProps extends TelemetryProps {
    /**
     * Possible dashboard id. All insights on the page will be got from
     * dashboard's info from the user or org settings by the dashboard id.
     * In case if id is undefined we get insights from the final
     * version of merged settings (all insights)
     */
    dashboardID?: string
}

export const DashboardsContentPage: FC<DashboardsContentPageProps> = props => {
    const { dashboardID, telemetryService } = props
    const { url } = useRouteMatch()

    const { dashboards, loading } = useInsightDashboards()
    const currentDashboard = useMemo(
        () => dashboards?.find(dashboard => dashboard.id === dashboardID),
        [dashboardID, dashboards]
    )

    if (!dashboardID) {
        // In case if url doesn't have a dashboard id we should fall back on
        // built-in "All insights" dashboard
        return <Redirect to={`${url}/${ALL_INSIGHTS_DASHBOARD.id}`} />
    }

    if (loading || !dashboards) {
        return <LoadingSpinner aria-live="off" inline={false} />
    }

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
