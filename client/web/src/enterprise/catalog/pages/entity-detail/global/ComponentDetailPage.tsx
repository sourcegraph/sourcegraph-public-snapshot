import React, { useEffect } from 'react'

import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import { ExtensionsControllerProps } from '@sourcegraph/shared/src/extensions/controller'
import { useQuery } from '@sourcegraph/shared/src/graphql/apollo'
import { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeProps } from '@sourcegraph/shared/src/theme'

import { PageTitle } from '../../../../../components/PageTitle'
import { ComponentByNameResult, ComponentByNameVariables } from '../../../../../graphql-operations'

import { ComponentDetailContent } from './ComponentDetailContent'
import { COMPONENT_BY_NAME } from './gql'

export interface Props
    extends TelemetryProps,
        ExtensionsControllerProps,
        ThemeProps,
        SettingsCascadeProps,
        PlatformContextProps {
    /** The name of the catalog component. */
    componentName: string
}

/**
 * The catalog component detail page.
 */
export const ComponentDetailPage: React.FunctionComponent<Props> = ({ componentName, telemetryService, ...props }) => {
    useEffect(() => {
        telemetryService.logViewEvent('ComponentDetail')
    }, [telemetryService])

    const { data, error, loading } = useQuery<ComponentByNameResult, ComponentByNameVariables>(COMPONENT_BY_NAME, {
        variables: { name: componentName },

        // Cache this data but always re-request it in the background when we revisit
        // this page to pick up newer changes.
        fetchPolicy: 'cache-and-network',

        // For subsequent requests while this page is open, make additional network
        // requests; this is necessary for `refetch` to actually use the network. (see
        // https://github.com/apollographql/apollo-client/issues/5515)
        nextFetchPolicy: 'network-only',
    })

    return (
        <>
            <PageTitle
                title={
                    error
                        ? 'Error loading component'
                        : loading && !data
                        ? 'Loading component...'
                        : !data || !data.component
                        ? 'Component not found'
                        : data.component.name
                }
            />
            {loading && !data ? (
                <LoadingSpinner className="m-3 icon-inline" />
            ) : error && !data ? (
                <div className="m-3 alert alert-danger">Error: {error.message}</div>
            ) : !data || !data.component ? (
                <div className="m-3 alert alert-danger">Component not found in catalog</div>
            ) : (
                <ComponentDetailContent {...props} component={data.component} telemetryService={telemetryService} />
            )}
        </>
    )
}
