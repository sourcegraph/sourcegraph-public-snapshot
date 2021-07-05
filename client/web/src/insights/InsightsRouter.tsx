import MapSearchIcon from 'mdi-react/MapSearchIcon'
import React from 'react'
import { RouteComponentProps, Switch, Route, useRouteMatch } from 'react-router'

import { ExtensionsControllerProps } from '@sourcegraph/shared/src/extensions/controller'
import { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'

import { AuthenticatedUser } from '../auth'
import { withAuthenticatedUser } from '../auth/withAuthenticatedUser'
import { HeroPage } from '../components/HeroPage'
import { lazyComponent } from '../util/lazyComponent'

import { CreationRoutes } from './pages/creation/CreationRoutes'
import { getExperimentalFeatures } from './utils/get-experimental-features'

const InsightsLazyPage = lazyComponent(() => import('./pages/insights/InsightsPage'), 'InsightsPage')
const EditInsightLazyPage = lazyComponent(() => import('./pages/edit/EditInsightPage'), 'EditInsightPage')
const DashboardsLazyPage = lazyComponent(() => import('./pages/dashboards/DashboardsPage'), 'DashboardsPage')

const NotFoundPage: React.FunctionComponent = () => <HeroPage icon={MapSearchIcon} title="404: Not Found" />

/**
 * This interface has to receive union type props derived from all child components
 * Because we need to pass all required prop from main Sourcegraph.tsx component to
 * sub-components withing app tree.
 */
export interface InsightsRouterProps
    extends SettingsCascadeProps,
        PlatformContextProps,
        TelemetryProps,
        ExtensionsControllerProps {
    /**
     * Authenticated user info, Used to decide where code insight will appears
     * in personal dashboard (private) or in organisation dashboard (public)
     * */
    authenticatedUser: AuthenticatedUser
}

/**
 * Main Insight routing component. Main entry point to code insights UI.
 */
export const InsightsRouter = withAuthenticatedUser<InsightsRouterProps>(props => {
    const { platformContext, settingsCascade, telemetryService, extensionsController, authenticatedUser } = props

    const match = useRouteMatch()
    const { codeInsightsDashboards } = getExperimentalFeatures(settingsCascade)

    return (
        <Switch>
            <Route path={match.url} exact={true}>
                <InsightsLazyPage
                    telemetryService={telemetryService}
                    platformContext={platformContext}
                    settingsCascade={settingsCascade}
                    extensionsController={extensionsController}
                />
            </Route>

            <Route path={`${match.url}/create`}>
                <CreationRoutes
                    platformContext={platformContext}
                    authenticatedUser={authenticatedUser}
                    settingsCascade={settingsCascade}
                    telemetryService={telemetryService}
                />
            </Route>

            <Route
                path={`${match.url}/edit/:insightID`}
                render={(props: RouteComponentProps<{ insightID: string }>) => (
                    <EditInsightLazyPage
                        platformContext={platformContext}
                        authenticatedUser={authenticatedUser}
                        settingsCascade={settingsCascade}
                        insightID={props.match.params.insightID}
                    />
                )}
            />

            {codeInsightsDashboards && (
                <Route
                    path={`${match.url}/dashboard/:dashboardId?`}
                    render={(props: RouteComponentProps<{ dashboardId: string }>) => (
                        <DashboardsLazyPage
                            telemetryService={telemetryService}
                            extensionsController={extensionsController}
                            settingsCascade={settingsCascade}
                            dashboardID={props.match.params.dashboardId}
                        />
                    )}
                />
            )}

            <Route component={NotFoundPage} key="hardcoded-key" />
        </Switch>
    )
})
