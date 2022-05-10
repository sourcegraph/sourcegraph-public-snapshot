import React, { useContext, useMemo } from 'react'

import MapSearchIcon from 'mdi-react/MapSearchIcon'
import { RouteComponentProps, Switch, Route, useRouteMatch } from 'react-router'
import { Redirect } from 'react-router-dom'

import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { lazyComponent } from '@sourcegraph/shared/src/util/lazyComponent'
import { LoadingSpinner, useObservable } from '@sourcegraph/wildcard'

import { AuthenticatedUser } from '../../auth'
import { withAuthenticatedUser } from '../../auth/withAuthenticatedUser'
import { HeroPage } from '../../components/HeroPage'

import { CodeInsightsBackendContext } from './core'
import { GaConfirmationModal } from './modals/GaConfirmationModal'
import {
    CodeInsightsRootPage,
    CodeInsightsRootPageTab,
    CodeInsightsRootPageURLPaths,
} from './pages/CodeInsightsRootPage'
import { InsightsDashboardCreationPage } from './pages/dashboards/creation/InsightsDashboardCreationPage'
import { EditDashboardPage } from './pages/dashboards/edit-dashboard/EditDashobardPage'
import { CreationRoutes } from './pages/insights/creation/CreationRoutes'
import { CodeInsightIndependentPage } from './pages/insights/insight/CodeInsightIndependentPage'

const EditInsightLazyPage = lazyComponent(
    () => import('./pages/insights/edit-insight/EditInsightPage'),
    'EditInsightPage'
)

const NotFoundPage: React.FunctionComponent<React.PropsWithChildren<unknown>> = () => (
    <HeroPage icon={MapSearchIcon} title="404: Not Found" />
)

const CODE_INSIGHT_STANDALONE_PAGE = true

/**
 * This interface has to receive union type props derived from all child components
 * Because we need to pass all required prop from main Sourcegraph.tsx component to
 * subcomponents withing app tree.
 */
export interface CodeInsightsAppRouter extends TelemetryProps {
    /**
     * Authenticated user info, Used to decide where code insight will appear
     * in personal dashboard (private) or in organisation dashboard (public)
     */
    authenticatedUser: AuthenticatedUser
}

/**
 * Main Insight routing component. Main entry point to code insights UI.
 */
export const CodeInsightsAppRouter = withAuthenticatedUser<CodeInsightsAppRouter>(props => {
    const { telemetryService, authenticatedUser } = props
    const match = useRouteMatch()

    return (
        <>
            <Route path="*" component={GaConfirmationModal} />

            <Switch>
                <Route path={`${match.url}/create`}>
                    <CreationRoutes authenticatedUser={authenticatedUser} telemetryService={telemetryService} />
                </Route>

                {CODE_INSIGHT_STANDALONE_PAGE && (
                    <Route
                        path={`${match.url}/insight/:id`}
                        render={(props: RouteComponentProps<{ id: string }>) => (
                            <CodeInsightIndependentPage
                                insightId={props.match.params.id}
                                telemetryService={telemetryService}
                            />
                        )}
                    />
                )}

                <Route
                    path={`${match.url}/edit/:insightID`}
                    render={(props: RouteComponentProps<{ insightID: string }>) => (
                        <EditInsightLazyPage
                            authenticatedUser={authenticatedUser}
                            insightID={props.match.params.insightID}
                        />
                    )}
                />

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
                    path={`${match.url}/add-dashboard`}
                    render={() => <InsightsDashboardCreationPage telemetryService={telemetryService} />}
                />

                <Route
                    path={[
                        `${match.url}${CodeInsightsRootPageURLPaths.CodeInsights}`,
                        `${match.url}${CodeInsightsRootPageURLPaths.GettingStarted}`,
                    ]}
                    render={props => (
                        <CodeInsightsRootPage
                            activeView={
                                props.match.path === `${match.url}${CodeInsightsRootPageURLPaths.CodeInsights}`
                                    ? CodeInsightsRootPageTab.CodeInsights
                                    : CodeInsightsRootPageTab.GettingStarted
                            }
                            telemetryService={telemetryService}
                        />
                    )}
                />

                <Route path={match.url} exact={true} component={CodeInsightsRedirect} />

                <Route component={NotFoundPage} key="hardcoded-key" />
            </Switch>
        </>
    )
})

const CodeInsightsRedirect: React.FunctionComponent<React.PropsWithChildren<unknown>> = () => {
    const { hasInsights } = useContext(CodeInsightsBackendContext)

    const match = useRouteMatch()
    const isThereAvailableInsights = useObservable(useMemo(() => hasInsights(1), [hasInsights]))

    if (isThereAvailableInsights === undefined) {
        return <LoadingSpinner />
    }

    return isThereAvailableInsights ? (
        <Redirect from={match.url} exact={true} to={`${match.url}/dashboards/all`} />
    ) : (
        <Redirect from={match.url} exact={true} to={`${match.url}/about`} />
    )
}
