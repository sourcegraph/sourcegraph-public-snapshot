import { FC, useContext, useMemo } from 'react'

import MapSearchIcon from 'mdi-react/MapSearchIcon'
import { Route, RouteComponentProps, Switch, useRouteMatch } from 'react-router'
import { Redirect } from 'react-router-dom'

import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { lazyComponent } from '@sourcegraph/shared/src/util/lazyComponent'
import { useObservable } from '@sourcegraph/wildcard'

import { AuthenticatedUser } from '../../auth'
import { withAuthenticatedUser } from '../../auth/withAuthenticatedUser'
import { HeroPage } from '../../components/HeroPage'

import { CodeInsightsBackendContext } from './core'
import { useApi } from './hooks'
import { useLicense } from './hooks/use-license'
import { GaConfirmationModal } from './modals/GaConfirmationModal'
import { CodeInsightsRootPage, CodeInsightsRootPageTab } from './pages/CodeInsightsRootPage'
import { InsightsDashboardCreationPage } from './pages/dashboards/creation/InsightsDashboardCreationPage'
import { EditDashboardPage } from './pages/dashboards/edit-dashboard/EditDashobardPage'
import { CreationRoutes } from './pages/insights/creation/CreationRoutes'
import { CodeInsightIndependentPage } from './pages/insights/insight/CodeInsightIndependentPage'

const EditInsightLazyPage = lazyComponent(
    () => import('./pages/insights/edit-insight/EditInsightPage'),
    'EditInsightPage'
)

const NotFoundPage: FC = () => <HeroPage icon={MapSearchIcon} title="404: Not Found" />

export interface CodeInsightsAppRouter extends TelemetryProps {
    authenticatedUser: AuthenticatedUser
}

export const CodeInsightsAppRouter = withAuthenticatedUser<CodeInsightsAppRouter>(props => {
    const { telemetryService } = props
    const match = useRouteMatch()

    const fetched = useLicense()
    const api = useApi()

    if (!fetched) {
        return null
    }

    return (
        <CodeInsightsBackendContext.Provider value={api}>
            <Route path="*" component={GaConfirmationModal} />

            <Switch>
                <Route path={`${match.url}/create`}>
                    <CreationRoutes telemetryService={telemetryService} />
                </Route>

                <Route
                    path={`${match.url}/insight/:id`}
                    render={props => (
                        <CodeInsightIndependentPage
                            insightId={props.match.params.id}
                            telemetryService={telemetryService}
                        />
                    )}
                />

                <Route
                    path={`${match.url}/edit/:insightId`}
                    render={props => (
                        <EditInsightLazyPage insightID={props.match.params.insightId} />
                    )}
                />

                <Route
                    path={`${match.url}/dashboards/:dashboardId/edit`}
                    render={props => (
                        <EditDashboardPage dashboardId={props.match.params.dashboardId} />
                    )}
                />

                <Route
                    path={`${match.url}/add-dashboard`}
                    render={() => <InsightsDashboardCreationPage telemetryService={telemetryService} />}
                />

                <Route
                    path={[`${match.url}/dashboards/:dashboardId?`, `${match.url}/all`, `${match.url}/about`]}
                    render={(props: RouteComponentProps<{ dashboardId?: string }>) => (
                        <CodeInsightsRootPage
                            dashboardId={props.match.params.dashboardId}
                            activeTab={getActiveTabByURL(match.url, props)}
                            telemetryService={telemetryService}
                        />
                    )}
                />

                <Route path={match.url} exact={true} component={CodeInsightsRedirect} />

                <Route component={NotFoundPage} key="hardcoded-key" />
            </Switch>
        </CodeInsightsBackendContext.Provider>
    )
})

const CodeInsightsRedirect: FC = () => {
    const { hasInsights } = useContext(CodeInsightsBackendContext)

    const match = useRouteMatch()
    const isThereAvailableInsights = useObservable(useMemo(() => hasInsights(1), [hasInsights]))

    if (isThereAvailableInsights === undefined) {
        return null
    }

    return isThereAvailableInsights ? (
        <Redirect from={match.url} exact={true} to={`${match.url}/all`} />
    ) : (
        <Redirect from={match.url} exact={true} to={`${match.url}/about`} />
    )
}

function getActiveTabByURL(
    matchURL: string,
    props: RouteComponentProps<{ dashboardId?: string }>
): CodeInsightsRootPageTab {
    const { match } = props

    switch (match.path) {
        case `${matchURL}/dashboards/:dashboardId?`:
            return CodeInsightsRootPageTab.Dashboards

        case `${matchURL}/all`:
            return CodeInsightsRootPageTab.AllInsights

        case `${matchURL}/about`:
            return CodeInsightsRootPageTab.GettingStarted

        default:
            return CodeInsightsRootPageTab.Dashboards
    }
}
