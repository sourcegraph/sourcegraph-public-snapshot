import MapSearchIcon from 'mdi-react/MapSearchIcon'
import React, { useMemo } from 'react'
import { RouteComponentProps, Switch, Route, useRouteMatch } from 'react-router'
import { Redirect } from 'react-router-dom'

import { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'

import { AuthenticatedUser } from '../../auth'
import { withAuthenticatedUser } from '../../auth/withAuthenticatedUser'
import { HeroPage } from '../../components/HeroPage'
import { lazyComponent } from '../../util/lazyComponent'

import { CodeInsightsBackendContext } from './core/backend/code-insights-backend-context'
import { CodeInsightsSettingsCascadeBackend } from './core/backend/code-insights-setting-cascade-backend'
import { BetaConfirmationModal } from './modals/BetaConfirmationModal'
import { DashboardsRoutes } from './pages/dashboards/DasbhoardsRoutes'
import { CreationRoutes } from './pages/insights/creation/CreationRoutes'

const EditInsightLazyPage = lazyComponent(
    () => import('./pages/insights/edit-insight/EditInsightPage'),
    'EditInsightPage'
)

const NotFoundPage: React.FunctionComponent = () => <HeroPage icon={MapSearchIcon} title="404: Not Found" />

/**
 * This interface has to receive union type props derived from all child components
 * Because we need to pass all required prop from main Sourcegraph.tsx component to
 * sub-components withing app tree.
 */
export interface InsightsRouterProps extends SettingsCascadeProps, PlatformContextProps, TelemetryProps {
    /**
     * Authenticated user info, Used to decide where code insight will appears
     * in personal dashboard (private) or in organisation dashboard (public)
     */
    authenticatedUser: AuthenticatedUser
}

/**
 * Main Insight routing component. Main entry point to code insights UI.
 */
export const InsightsRouter = withAuthenticatedUser<InsightsRouterProps>(props => {
    const { platformContext, settingsCascade, telemetryService, authenticatedUser } = props

    const match = useRouteMatch()

    const api = useMemo(() => new CodeInsightsSettingsCascadeBackend(settingsCascade, platformContext), [
        platformContext,
        settingsCascade,
    ])

    return (
        <CodeInsightsBackendContext.Provider value={api}>
            <Route path="*" component={BetaConfirmationModal} />
            <Switch>
                <Redirect from={match.url} exact={true} to={`${match.url}/dashboards/all`} />

                <Route path={`${match.url}/create`}>
                    <CreationRoutes authenticatedUser={authenticatedUser} telemetryService={telemetryService} />
                </Route>

                <Route
                    path={`${match.url}/edit/:insightID`}
                    render={(props: RouteComponentProps<{ insightID: string }>) => (
                        <EditInsightLazyPage
                            authenticatedUser={authenticatedUser}
                            insightID={props.match.params.insightID}
                        />
                    )}
                />

                <DashboardsRoutes authenticatedUser={authenticatedUser} telemetryService={telemetryService} />

                <Route component={NotFoundPage} key="hardcoded-key" />
            </Switch>
        </CodeInsightsBackendContext.Provider>
    )
})
