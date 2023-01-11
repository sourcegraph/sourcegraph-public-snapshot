import React from 'react'

import { RouteComponentProps, Switch, Route, Redirect } from 'react-router'
import { SiteExternalServiceConfigResult, SiteExternalServiceConfigVariables } from 'src/graphql-operations'

import { useQuery } from '@sourcegraph/http-client'
import { AuthenticatedUser } from '@sourcegraph/shared/src/auth'
import { Scalars } from '@sourcegraph/shared/src/graphql-operations'
import { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { lazyComponent } from '@sourcegraph/shared/src/util/lazyComponent'
import { LoadingSpinner, ErrorAlert } from '@sourcegraph/wildcard'

import { codeHostExternalServices, nonCodeHostExternalServices } from '../components/externalServices/externalServices'

import { SITE_EXTERNAL_SERVICE_CONFIG } from './backend'

const ExternalServicesPage = lazyComponent(
    () => import('../components/externalServices/ExternalServicesPage'),
    'ExternalServicesPage'
)
const ExternalServiceEditPage = lazyComponent(
    () => import('../components/externalServices/ExternalServiceEditPage'),
    'ExternalServiceEditPage'
)
const ExternalServicePage = lazyComponent(
    () => import('../components/externalServices/ExternalServicePage'),
    'ExternalServicePage'
)
const AddExternalServicesPage = lazyComponent(
    () => import('../components/externalServices/AddExternalServicesPage'),
    'AddExternalServicesPage'
)

interface Props
    extends RouteComponentProps<{}>,
        ThemeProps,
        TelemetryProps,
        PlatformContextProps,
        SettingsCascadeProps {
    authenticatedUser: AuthenticatedUser
}

export const SiteAdminExternalServicesArea: React.FunctionComponent<React.PropsWithChildren<Props>> = ({
    match,
    ...outerProps
}) => {
    const { data, error, loading } = useQuery<SiteExternalServiceConfigResult, SiteExternalServiceConfigVariables>(
        SITE_EXTERNAL_SERVICE_CONFIG,
        {}
    )

    if (error && !loading) {
        return <ErrorAlert error={error} />
    }

    if (loading && !error) {
        return <LoadingSpinner />
    }

    if (!data) {
        return null
    }

    return (
        <Switch>
            <Route
                path={match.url}
                render={props => (
                    <ExternalServicesPage
                        {...outerProps}
                        {...props}
                        routingPrefix="/site-admin"
                        afterDeleteRoute="/site-admin/repositories?repositoriesUpdated"
                        externalServicesFromFile={data?.site?.externalServicesFromFile}
                        allowEditExternalServicesWithFile={data?.site?.allowEditExternalServicesWithFile}
                    />
                )}
                exact={true}
            />
            <Route path={match.url + '/add'} render={() => <Redirect to="new" />} exact={true} />
            <Route
                path={`${match.url}/new`}
                render={props => (
                    <AddExternalServicesPage
                        {...outerProps}
                        {...props}
                        routingPrefix="/site-admin"
                        codeHostExternalServices={codeHostExternalServices}
                        nonCodeHostExternalServices={nonCodeHostExternalServices}
                        externalServicesFromFile={data?.site?.externalServicesFromFile}
                        allowEditExternalServicesWithFile={data?.site?.allowEditExternalServicesWithFile}
                    />
                )}
                exact={true}
            />
            <Route
                path={`${match.url}/:id`}
                render={({ match, ...props }: RouteComponentProps<{ id: Scalars['ID'] }>) => (
                    <ExternalServicePage
                        {...outerProps}
                        {...props}
                        routingPrefix="/site-admin"
                        afterDeleteRoute="/site-admin/external-services"
                        externalServiceID={match.params.id}
                        externalServicesFromFile={data?.site?.externalServicesFromFile}
                        allowEditExternalServicesWithFile={data?.site?.allowEditExternalServicesWithFile}
                    />
                )}
                exact={true}
            />
            <Route
                path={`${match.url}/:id/edit`}
                render={({ match, ...props }: RouteComponentProps<{ id: Scalars['ID'] }>) => (
                    <ExternalServiceEditPage
                        {...outerProps}
                        {...props}
                        routingPrefix="/site-admin"
                        externalServiceID={match.params.id}
                        externalServicesFromFile={data?.site?.externalServicesFromFile}
                        allowEditExternalServicesWithFile={data?.site?.allowEditExternalServicesWithFile}
                    />
                )}
                exact={true}
            />
        </Switch>
    )
}
