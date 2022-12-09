import { FC, useEffect, useState } from 'react'

import { gql, useLazyQuery } from '@apollo/client'
import MapSearchIcon from 'mdi-react/MapSearchIcon'
import { Route, RouteComponentProps, Switch, useRouteMatch } from 'react-router'
import { Redirect } from 'react-router-dom'

import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { lazyComponent } from '@sourcegraph/shared/src/util/lazyComponent'
import { useTemporarySetting } from '@sourcegraph/shared/src/settings/temporary/useTemporarySetting'

import { GetFirstAvailableDashboardResult } from '../../graphql-operations'
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
                    render={props => <EditInsightLazyPage insightID={props.match.params.insightId} />}
                />

                <Route
                    path={`${match.url}/dashboards/:dashboardId/edit`}
                    render={props => <EditDashboardPage dashboardId={props.match.params.dashboardId} />}
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
                <Route render={() => <HeroPage icon={MapSearchIcon} title="404: Not Found" />} key="hardcoded-key" />
            </Switch>
        </CodeInsightsBackendContext.Provider>
    )
})

const CodeInsightsRedirect: FC = () => {
    const match = useRouteMatch()
    const state = useDashboardExistence()

    if (state.status === 'loading') {
        return null
    }

    // No dashboards status means that there are no insights either, so redirect
    // to the getting started page in this case
    if (state.status === 'noDashboards') {
        return <Redirect from={match.url} exact={true} to={`${match.url}/about`} />
    }

    // There are some dashboards, but we didn't find any particular dashboard in the user
    // temporal settings so redirect to the dashboard tab and select first private dashboard
    if (state.status === 'availableDashboard') {
        return <Redirect from={match.url} exact={true} to={`${match.url}/dashboards`} />
    }

    // We found a recently viewed dashboard id in the temporal settings, so redirect to this
    // dashboard.
    return <Redirect from={match.url} exact={true} to={`${match.url}/dashboards/${state.dashboardId}`} />
}

type DashboardExistence =
    | { status: 'availableDashboard' }
    | { status: 'lastVisitedDashboard'; dashboardId: string }
    | { status: 'noDashboards' }
    | { status: 'loading' }

function useDashboardExistence(): DashboardExistence {
    const [state, setState] = useState<DashboardExistence>({ status: 'loading' })
    const [lastVisitedDashboardId, , temporalSettingStatus] = useTemporarySetting(
        'insights.lastVisitedDashboardId',
        null
    )

    const [fetchFirstAvailableDashboard] = useLazyQuery<GetFirstAvailableDashboardResult>(gql`
        query GetFirstAvailableDashboard {
            insightsDashboards(first: 1) {
                nodes {
                    id
                }
            }
        }
    `)

    useEffect(() => {
        // We're still loading temporal settings
        if (temporalSettingStatus === 'initial') {
            return
        }

        // Skip dashboard async check if three is last visited dashboard id in the history
        // return users to the recently viewed dashboard
        if (lastVisitedDashboardId) {
            return setState({ status: 'lastVisitedDashboard', dashboardId: lastVisitedDashboardId })
        }

        // Check asynchronously about dashboard existence on the backend
        fetchFirstAvailableDashboard()
            .then(result => {
                const { data = { insightsDashboards: { nodes: [] } } } = result
                const isThereAnyDashboards = data.insightsDashboards.nodes.length > 0

                if (isThereAnyDashboards) {
                    setState({ status: 'availableDashboard' })
                } else {
                    setState({ status: 'noDashboards' })
                }
            })
            .catch(() => {
                setState({ status: 'noDashboards' })
            })
    }, [fetchFirstAvailableDashboard, lastVisitedDashboardId, temporalSettingStatus])

    return state
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
