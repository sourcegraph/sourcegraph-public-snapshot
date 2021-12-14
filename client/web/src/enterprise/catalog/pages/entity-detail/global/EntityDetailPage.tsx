import React, { useEffect } from 'react'

import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import { ExtensionsControllerProps } from '@sourcegraph/shared/src/extensions/controller'
import { useQuery } from '@sourcegraph/shared/src/graphql/apollo'
import { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeProps } from '@sourcegraph/shared/src/theme'

import { PageTitle } from '../../../../../components/PageTitle'
import { CatalogEntityByNameResult, CatalogEntityByNameVariables } from '../../../../../graphql-operations'

import { EntityDetailContent } from './EntityDetailContent'
import { CATALOG_ENTITY_BY_NAME } from './gql'

export interface Props
    extends TelemetryProps,
        ExtensionsControllerProps,
        ThemeProps,
        SettingsCascadeProps,
        PlatformContextProps {
    /** The name of the catalog entity. */
    entityName: string
}

/**
 * The catalog entity detail page.
 */
export const EntityDetailPage: React.FunctionComponent<Props> = ({ entityName, telemetryService, ...props }) => {
    useEffect(() => {
        telemetryService.logViewEvent('CatalogEntityDetail')
    }, [telemetryService])

    const { data, error, loading } = useQuery<CatalogEntityByNameResult, CatalogEntityByNameVariables>(
        CATALOG_ENTITY_BY_NAME,
        {
            variables: { type: 'COMPONENT', name: entityName },

            // Cache this data but always re-request it in the background when we revisit
            // this page to pick up newer changes.
            fetchPolicy: 'cache-and-network',

            // For subsequent requests while this page is open, make additional network
            // requests; this is necessary for `refetch` to actually use the network. (see
            // https://github.com/apollographql/apollo-client/issues/5515)
            nextFetchPolicy: 'network-only',
        }
    )

    return (
        <>
            <PageTitle
                title={
                    error
                        ? 'Error loading entity'
                        : loading && !data
                        ? 'Loading entity...'
                        : !data || !data.catalogEntity
                        ? 'Entity not found'
                        : data.catalogEntity.name
                }
            />
            {loading && !data ? (
                <LoadingSpinner className="m-3 icon-inline" />
            ) : error && !data ? (
                <div className="m-3 alert alert-danger">Error: {error.message}</div>
            ) : !data || !data.catalogEntity ? (
                <div className="m-3 alert alert-danger">Entity not found in catalog</div>
            ) : (
                <EntityDetailContent {...props} entity={data.catalogEntity} telemetryService={telemetryService} />
            )}
        </>
    )
}
