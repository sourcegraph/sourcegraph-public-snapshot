import React, { useMemo } from 'react'
import { Route, RouteComponentProps, Switch, useRouteMatch } from 'react-router'

import { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'

import { AuthenticatedUser } from '../../../../auth'
import { Settings } from '../../../../schema/settings.schema'
import { CodeInsightsBackendContext } from '../../core/backend/code-insights-backend-context';
import { CodeInsightsSettingsCascadeBackend } from '../../core/backend/code-insights-setting-cascade-backend';

import { InsightsDashboardCreationPage } from './creation/InsightsDashboardCreationPage'
import { DashboardsPage } from './dashboard-page/DashboardsPage'
import { EditDashboardPage } from './edit-dashboard/EditDashobardPage'

export interface DashboardsRoutesProps
    extends TelemetryProps,
        SettingsCascadeProps<Settings>,
        PlatformContextProps<'updateSettings'> {
    authenticatedUser: AuthenticatedUser
}

export const DashboardsRoutes: React.FunctionComponent<DashboardsRoutesProps> = props => {
    const { authenticatedUser, settingsCascade, platformContext, telemetryService } = props
    const match = useRouteMatch()

    const api = useMemo(() => {
        console.log('recreate setting based api context')

        return new CodeInsightsSettingsCascadeBackend(settingsCascade, platformContext)
    }, [platformContext, settingsCascade])

    return (
        <CodeInsightsBackendContext.Provider value={api}>
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
        </CodeInsightsBackendContext.Provider>
    )
}
