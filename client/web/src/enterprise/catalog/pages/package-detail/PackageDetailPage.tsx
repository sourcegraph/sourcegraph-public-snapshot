import classNames from 'classnames'
import React, { useEffect } from 'react'

import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import { FormatListBulletedIcon } from '@sourcegraph/shared/src/components/icons'
import { ExtensionsControllerProps } from '@sourcegraph/shared/src/extensions/controller'
import { useQuery } from '@sourcegraph/shared/src/graphql/apollo'
import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeProps } from '@sourcegraph/shared/src/theme'

import { PageTitle } from '../../../../components/PageTitle'
import { CatalogPackageByNameResult, CatalogPackageByNameVariables } from '../../../../graphql-operations'
import { useTemporarySetting } from '../../../../settings/temporary/useTemporarySetting'
import { useCatalogPackageFilters } from '../../core/entity-filters'

import { CATALOG_ENTITY_BY_NAME } from './gql'
import { PackageDetailContent } from './PackageDetailContent'
import styles from './PackageDetailPage.module.scss'

export interface Props extends TelemetryProps, ExtensionsControllerProps, ThemeProps, SettingsCascadeProps {
    /** The name of the catalog entity. */
    entityName: string
}

/**
 * The catalog entity detail page.
 */
export const PackageDetailPage: React.FunctionComponent<Props> = ({ entityName, telemetryService, ...props }) => {
    useEffect(() => {
        telemetryService.logViewEvent('CatalogPackageDetail')
    }, [telemetryService])

    const { data, error, loading } = useQuery<CatalogPackageByNameResult, CatalogPackageByNameVariables>(
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

    const disableSidebar = true
    const [showSidebar, setShowSidebar] = useTemporarySetting('catalog.sidebar.visible', true)

    const { filters, onFiltersChange } = useCatalogPackageFilters('')

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
            {!disableSidebar &&
                (showSidebar ? (
                    <Sidebar>
                        <EntityList
                            selectedEntityName={entityName}
                            filters={filters}
                            onFiltersChange={onFiltersChange}
                            className="flex-1"
                        />
                        <div className="flex-1" />
                        <button type="button" className="btn btn-link btn-sm" onClick={() => setShowSidebar(false)}>
                            Hide sidebar
                        </button>
                    </Sidebar>
                ) : (
                    <button
                        type="button"
                        className={classNames('btn btn-secondary btn-sm', styles.showSidebarBtn)}
                        onClick={() => setShowSidebar(true)}
                        title="Show sidebar"
                    >
                        <FormatListBulletedIcon className="icon-inline" />
                    </button>
                ))}
            {loading && !data ? (
                <LoadingSpinner className="m-3 icon-inline" />
            ) : error && !data ? (
                <div className="m-3 alert alert-danger">Error: {error.message}</div>
            ) : !data || !data.catalogEntity ? (
                <div className="m-3 alert alert-danger">Entity not found in catalog</div>
            ) : (
                <PackageDetailContent {...props} entity={data.catalogEntity} telemetryService={telemetryService} />
            )}
        </>
    )
}
