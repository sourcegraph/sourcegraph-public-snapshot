import React, { useEffect } from 'react'

import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import { ExtensionsControllerProps } from '@sourcegraph/shared/src/extensions/controller'
import { useQuery } from '@sourcegraph/shared/src/graphql/apollo'
import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeProps } from '@sourcegraph/shared/src/theme'

import { PageTitle } from '../../../../../components/PageTitle'
import { CatalogComponentByNameResult, CatalogComponentByNameVariables } from '../../../../../graphql-operations'
import { CatalogComponentFiltersProps } from '../../../core/component-filters'
import { ComponentList } from '../../overview/components/component-list/ComponentList'
import { Sidebar } from '../sidebar/Sidebar'

import { ComponentDetailContent } from './ComponentDetailContent'
import { CATALOG_COMPONENT_BY_NAME } from './gql'

export interface Props
    extends CatalogComponentFiltersProps,
        TelemetryProps,
        ExtensionsControllerProps,
        ThemeProps,
        SettingsCascadeProps {
    /** The name of the CatalogComponent. */
    catalogComponentName: string
}

/**
 * The catalog component detail page.
 */
export const ComponentDetailPage: React.FunctionComponent<Props> = ({
    catalogComponentName,
    filters,
    onFiltersChange,
    telemetryService,
    ...props
}) => {
    useEffect(() => {
        telemetryService.logViewEvent('CatalogComponentDetail')
    }, [telemetryService])

    const { data, error, loading } = useQuery<CatalogComponentByNameResult, CatalogComponentByNameVariables>(
        CATALOG_COMPONENT_BY_NAME,
        {
            variables: { name: catalogComponentName },

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
                        ? 'Error loading component'
                        : loading && !data
                        ? 'Loading component...'
                        : !data || !data.catalogComponent
                        ? 'Component not found'
                        : data.catalogComponent.name
                }
            />
            <Sidebar>
                <ComponentList
                    selectedComponentName={catalogComponentName}
                    filters={filters}
                    onFiltersChange={onFiltersChange}
                    className="flex-1"
                    size="sm"
                />
            </Sidebar>
            <div className="pt-2 px-3 pb-4 overflow-auto w-100">
                {loading && !data ? (
                    <LoadingSpinner className="icon-inline" />
                ) : error && !data ? (
                    <div className="alert alert-danger">Error: {error.message}</div>
                ) : !data || !data.catalogComponent ? (
                    <div className="alert alert-danger">Component not found in catalog</div>
                ) : (
                    <ComponentDetailContent
                        {...props}
                        catalogComponent={data.catalogComponent}
                        telemetryService={telemetryService}
                    />
                )}
            </div>
        </>
    )
}
