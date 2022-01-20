import { useApolloClient } from '@apollo/client'
import MapSearchIcon from 'mdi-react/MapSearchIcon'
import React, { useMemo } from 'react'
import { RouteComponentProps, Switch, Route, useRouteMatch } from 'react-router'
import { Redirect } from 'react-router-dom'

import { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import { Settings } from '@sourcegraph/shared/src/schema/settings.schema'
import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { lazyComponent } from '@sourcegraph/shared/src/util/lazyComponent'

import { AuthenticatedUser } from '../../auth'
import { withAuthenticatedUser } from '../../auth/withAuthenticatedUser'
import { HeroPage } from '../../components/HeroPage'
import { CodeInsightsContextProps } from '../../insights/types'

import { CodeInsightsBackendContext } from './core/backend/code-insights-backend-context'
import { CodeInsightsGqlBackend } from './core/backend/gql-api/code-insights-gql-backend'
import { CodeInsightsSettingsCascadeBackend } from './core/backend/setting-based-api/code-insights-setting-cascade-backend'
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
export interface InsightsRouterProps
    extends CodeInsightsContextProps,
        SettingsCascadeProps<Settings>,
        PlatformContextProps,
        TelemetryProps {
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
    const { isCodeInsightsGqlApiEnabled, platformContext, settingsCascade, telemetryService, authenticatedUser } = props

    const match = useRouteMatch()
    const apolloClient = useApolloClient()

    const gqlApi = useMemo(() => new CodeInsightsGqlBackend(apolloClient), [apolloClient])
    const api = useMemo(
        () =>
            isCodeInsightsGqlApiEnabled
                ? gqlApi
                : new CodeInsightsSettingsCascadeBackend(settingsCascade, platformContext),
        [isCodeInsightsGqlApiEnabled, gqlApi, settingsCascade, platformContext]
    )

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
