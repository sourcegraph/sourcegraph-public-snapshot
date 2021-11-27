import { useApolloClient } from '@apollo/client'
import MapSearchIcon from 'mdi-react/MapSearchIcon'
import React, { useMemo } from 'react'
import { Switch, Route, useRouteMatch } from 'react-router'

import { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'

import { AuthenticatedUser } from '../../auth'
import { withAuthenticatedUser } from '../../auth/withAuthenticatedUser'
import { HeroPage } from '../../components/HeroPage'
import { Settings } from '../../schema/settings.schema'

import { CatalogBackendContext } from './core/backend/context'
import { CatalogGqlBackend } from './core/backend/gql-api/gql-backend'
import { OverviewPage } from './pages/overview/global/OverviewPage'

const NotFoundPage: React.FunctionComponent = () => <HeroPage icon={MapSearchIcon} title="404: Not Found" />

/**
 * This interface has to receive union type props derived from all child components
 * Because we need to pass all required prop from main Sourcegraph.tsx component to
 * sub-components withing app tree.
 */
export interface CatalogRouterProps extends SettingsCascadeProps<Settings>, PlatformContextProps, TelemetryProps {
    /**
     * Authenticated user info, Used to decide where code insight will appears
     * in personal dashboard (private) or in organisation dashboard (public)
     */
    authenticatedUser: AuthenticatedUser
}

/**
 * The main Catalog routing component (the main entrypoint to the Catalog UI).
 */
export const CatalogRouter = withAuthenticatedUser<CatalogRouterProps>(props => {
    const { platformContext, settingsCascade, telemetryService, authenticatedUser } = props

    const match = useRouteMatch()
    const apolloClient = useApolloClient()

    const api = useMemo(() => new CatalogGqlBackend(apolloClient), [apolloClient])

    return (
        <CatalogBackendContext.Provider value={api}>
            <Switch>
                <Route path={match.url}>
                    <OverviewPage authenticatedUser={authenticatedUser} telemetryService={telemetryService} />
                </Route>
                <Route component={NotFoundPage} key="hardcoded-key" />
            </Switch>
        </CatalogBackendContext.Provider>
    )
})
