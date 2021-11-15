import React from 'react'
import { Route, RouteComponentProps, Switch, useRouteMatch } from 'react-router'

import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'

import { AuthenticatedUser } from '../../../../auth'

import { InsightsDashboardCreationPage } from './creation/InsightsDashboardCreationPage'
import { DashboardsPage } from './dashboard-page/DashboardsPage'
import { EditDashboardPage } from './edit-dashboard/EditDashobardPage'

export interface DashboardsRoutesProps extends TelemetryProps {
    authenticatedUser: AuthenticatedUser
}

export const DashboardsRoutes: React.FunctionComponent<DashboardsRoutesProps> = props => {
    const { authenticatedUser, telemetryService } = props
    const match = useRouteMatch()

    return (
        <>
            <Switch>
                <Route
                    path={`${match.url}/dashboards/:dashboardId/edit`}
                    render={(routeProps: RouteComponentProps<{ dashboardId: string }>) => (
                        <EditDashboardPage
                            authenticatedUser={authenticatedUser}
                            dashboardId={routeProps.match.params.dashboardId}
                        />
                    )}
                />

                <Route
                    path={`${match.url}/dashboards/:dashboardId?`}
                    render={(routeProps: RouteComponentProps<{ dashboardId: string }>) => (
                        <DashboardsPage
                            telemetryService={telemetryService}
                            dashboardID={routeProps.match.params.dashboardId}
                        />
                    )}
                />

                <Route
                    path={`${match.url}/add-dashboard`}
                    render={() => (
                        <InsightsDashboardCreationPage
                            telemetryService={telemetryService}
                            authenticatedUser={authenticatedUser}
                        />
                    )}
                />
            </Switch>
        </>
    )
}
