import React, { useEffect } from 'react'

import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import { ExtensionsControllerProps } from '@sourcegraph/shared/src/extensions/controller'
import { useQuery } from '@sourcegraph/shared/src/graphql/apollo'
import { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeProps } from '@sourcegraph/shared/src/theme'

import { PageTitle } from '../../../../components/PageTitle'
import { CatalogPackageByNameResult, CatalogPackageByNameVariables } from '../../../../graphql-operations'

import { PACKAGE_BY_NAME } from './gql'
import { PackageDetailContent } from './PackageDetailContent'

export interface Props
    extends TelemetryProps,
        ExtensionsControllerProps,
        ThemeProps,
        SettingsCascadeProps,
        PlatformContextProps {
    /** The name of the catalog package. */
    entityName: string
}

/**
 * The catalog package detail page.
 */
export const PackageDetailPage: React.FunctionComponent<Props> = ({ entityName, telemetryService, ...props }) => {
    useEffect(() => {
        telemetryService.logViewEvent('CatalogPackageDetail')
    }, [telemetryService])

    const { data, error, loading } = useQuery<CatalogPackageByNameResult, CatalogPackageByNameVariables>(
        PACKAGE_BY_NAME,
        {
            variables: { name: entityName },

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
                        ? 'Error loading package'
                        : loading && !data
                        ? 'Loading package...'
                        : !data || !data.catalogEntity || data.catalogEntity.__typename !== 'Package'
                        ? 'Package not found'
                        : data.catalogEntity.name
                }
            />
            {loading && !data ? (
                <LoadingSpinner className="m-3 icon-inline" />
            ) : error && !data ? (
                <div className="m-3 alert alert-danger">Error: {error.message}</div>
            ) : !data || !data.catalogEntity || data.catalogEntity.__typename !== 'Package' ? (
                <div className="m-3 alert alert-danger">Package not found in catalog</div>
            ) : (
                <PackageDetailContent {...props} entity={data.catalogEntity} telemetryService={telemetryService} />
            )}
        </>
    )
}
