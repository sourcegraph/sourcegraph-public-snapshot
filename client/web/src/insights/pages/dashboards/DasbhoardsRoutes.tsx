import React from 'react'
import { Route, RouteComponentProps, Switch, useRouteMatch } from 'react-router';

import { ExtensionsControllerProps } from '@sourcegraph/shared/src/extensions/controller';
import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings';

import { Settings } from '../../../schema/settings.schema';
import { InsightsViewGridProps } from '../../components';

import { InsightsDashboardCreationPage } from './creation/InsightsDashboardCreationPage';
import { DashboardsPage } from './DashboardsPage';

export interface DashboardsRoutesProps extends
    Omit<InsightsViewGridProps, 'views' | 'settingsCascade'>,
    SettingsCascadeProps<Settings>,
    ExtensionsControllerProps {

}

/**
 * Displays Code Insights dashboard area.
 */
export const DashboardsRoutes: React.FunctionComponent<DashboardsRoutesProps> = props => {
    const match = useRouteMatch()

    return (
        <Switch>

            <Route
                path={`${match.url}/dashboard/:dashboardId?`}
                render={(routeProps: RouteComponentProps<{ dashboardId: string }>) => (
                    <DashboardsPage dashboardID={routeProps.match.params.dashboardId} {...props} />
                )}
            />

            <Route
                path={`${match.url}/add-dashboard`}
                render={() => <InsightsDashboardCreationPage/>}
            />
        </Switch>
    )
}
