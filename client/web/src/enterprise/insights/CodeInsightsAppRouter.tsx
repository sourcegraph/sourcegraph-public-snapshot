import { FC, useEffect, useState } from 'react'

import { gql, useLazyQuery } from '@apollo/client'
import MapSearchIcon from 'mdi-react/MapSearchIcon'
import { Route, RouteComponentProps, Switch, useRouteMatch } from 'react-router'
import { Redirect } from 'react-router-dom'

import { useTemporarySetting } from '@sourcegraph/shared/src/settings/temporary/useTemporarySetting'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { lazyComponent } from '@sourcegraph/shared/src/util/lazyComponent'

import { AuthenticatedUser } from '../../auth'
import { withAuthenticatedUser } from '../../auth/withAuthenticatedUser'
import { HeroPage } from '../../components/HeroPage'
import { GetFirstAvailableDashboardResult, GetFirstAvailableDashboardVariables } from '../../graphql-operations'

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

    const fetched = useLicense()
    const api = useApi()

    if (!fetched) {
        return null
    }

    return (
        <CodeInsightsBackendContext.Provider value={api}>
            <Route path="*" component={GaConfirmationModal} />

            <Switch>
                <Route path="/insights" exact={true} component={CodeInsightsSmartRoutingRedirect} />

                <Route path="/insights/create" render={() => <CreationRoutes telemetryService={telemetryService} />} />

                <Route
                    path="/insights/dashboards/:dashboardId/edit"
                    render={props => <EditDashboardPage dashboardId={props.match.params.dashboardId} />}
                />

                <Route
                    path="/insights/add-dashboard"
                    render={() => <InsightsDashboardCreationPage telemetryService={telemetryService} />}
                />

                <Route
                    path={['/insights/dashboards/:dashboardId?', '/insights/all', '/insights/about']}
                    render={(props: RouteComponentProps<{ dashboardId?: string }>) => (
                        <CodeInsightsRootPage
                            dashboardId={props.match.params.dashboardId}
                            activeTab={getActiveTabByURL('/insights', props)}
                            telemetryService={telemetryService}
                        />
                    )}
                />

                <Route
                    // Deprecated URL, delete this in the 4.10
                    path="/insights/edit/:insightId"
                    render={props => <Redirect to={`/insights/${props.match.params.insightId}/edit`} />}
                />

                <Route
                    path="/insights/:insightId/edit"
                    render={props => <EditInsightLazyPage insightID={props.match.params.insightId} />}
                />

                <Route
                    // Deprecated URL, delete this in the 4.10
                    path="/insights/insight/:id"
                    render={props => <Redirect to={`/insights/${props.match.params.id}`} />}
                />

                <Route
                    path="/insights/:id"
                    render={props => (
                        <CodeInsightIndependentPage
                            insightId={props.match.params.id}
                            telemetryService={telemetryService}
                        />
                    )}
                />

                <Route render={() => <HeroPage icon={MapSearchIcon} title="404: Not Found" />} key="hardcoded-key" />
            </Switch>
        </CodeInsightsBackendContext.Provider>
    )
})

const CodeInsightsSmartRoutingRedirect: FC = () => {
    const match = useRouteMatch()
    const state = useDashboardExistence()

    if (state.status === 'loading') {
        return null
    }

    // No dashboards status means that there are no insights either, so redirect
    // to the getting started page in this case
    if (state.status === 'noDashboards') {
        return <Redirect from={match.url} exact={true} to="/insights/about" />
    }

    // There are some dashboards, but we didn't find any particular dashboard in the user
    // temporal settings so redirect to the dashboard tab and select first private dashboard
    if (state.status === 'availableDashboard') {
        return <Redirect from={match.url} exact={true} to="/insights/dashboards" />
    }

    // We found a recently viewed dashboard id in the temporal settings, so redirect to this
    // dashboard.
    return <Redirect from={match.url} exact={true} to={`/insights/dashboards/${state.dashboardId}`} />
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

    const [fetchFirstAvailableDashboard] = useLazyQuery<
        GetFirstAvailableDashboardResult,
        GetFirstAvailableDashboardVariables
    >(gql`
        query GetFirstAvailableDashboard($lastVisitedDashboardId: ID) {
            lastVisitedDashboard: insightsDashboards(id: $lastVisitedDashboardId) {
                nodes {
                    id
                }
            }
            firstAvailableDashboard: insightsDashboards(first: 1) {
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

        const normalizedLastVisitedDashboardId = lastVisitedDashboardId ?? null

        // Check asynchronously about dashboard existence on the backend
        fetchFirstAvailableDashboard({ variables: { lastVisitedDashboardId: normalizedLastVisitedDashboardId } })
            .then(result => {
                const {
                    data = {
                        lastVisitedDashboard: { nodes: [] },
                        firstAvailableDashboard: { nodes: [] },
                    },
                } = result

                const [lastVisitedDashboard] = data.lastVisitedDashboard.nodes
                const [firstAvailableDashboard] = data.firstAvailableDashboard.nodes

                // We resolved dashboard by lastVisitedDashboardId in the temporal settings
                if (lastVisitedDashboard) {
                    setState({ status: 'lastVisitedDashboard', dashboardId: lastVisitedDashboard.id })
                    return
                }

                // If it's just another dashboard and not undefined (this mean we have at least one dashboard
                // on the backend redirect to the dashboard page without id
                if (firstAvailableDashboard) {
                    setState({ status: 'availableDashboard' })
                } else {
                    // Otherwise, redirect to the standard no dashboard view (which is getting started tab)
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
