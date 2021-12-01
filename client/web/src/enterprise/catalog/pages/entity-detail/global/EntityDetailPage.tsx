import classNames from 'classnames'
import React, { useEffect } from 'react'

import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import { FormatListBulletedIcon } from '@sourcegraph/shared/src/components/icons'
import { ExtensionsControllerProps } from '@sourcegraph/shared/src/extensions/controller'
import { useQuery } from '@sourcegraph/shared/src/graphql/apollo'
import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeProps } from '@sourcegraph/shared/src/theme'

import { PageTitle } from '../../../../../components/PageTitle'
import { CatalogEntityByNameResult, CatalogEntityByNameVariables } from '../../../../../graphql-operations'
import { useTemporarySetting } from '../../../../../settings/temporary/useTemporarySetting'
import { CatalogEntityFiltersProps } from '../../../core/entity-filters'
import { EntityList } from '../../overview/components/entity-list/EntityList'
import { Sidebar } from '../sidebar/Sidebar'

import { EntityDetailContent } from './EntityDetailContent'
import styles from './EntityDetailPage.module.scss'
import { CATALOG_ENTITY_BY_NAME } from './gql'

export interface Props
    extends CatalogEntityFiltersProps,
        TelemetryProps,
        ExtensionsControllerProps,
        ThemeProps,
        SettingsCascadeProps {
    /** The name of the catalog entity. */
    entityName: string
}

/**
 * The catalog entity detail page.
 */
export const EntityDetailPage: React.FunctionComponent<Props> = ({
    entityName,
    filters,
    onFiltersChange,
    telemetryService,
    ...props
}) => {
    useEffect(() => {
        telemetryService.logViewEvent('CatalogEntityDetail')
    }, [telemetryService])

    const { data, error, loading } = useQuery<CatalogEntityByNameResult, CatalogEntityByNameVariables>(
        CATALOG_ENTITY_BY_NAME,
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

    const [showSidebar, setShowSidebar] = useTemporarySetting('catalog.sidebar.visible', true)

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
            {showSidebar ? (
                <Sidebar>
                    <EntityList
                        selectedEntityName={entityName}
                        filters={filters}
                        onFiltersChange={onFiltersChange}
                        className="flex-1"
                        size="sm"
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
            )}
            <div className="pt-2 px-3 pb-4 overflow-auto w-100">
                {loading && !data ? (
                    <LoadingSpinner className="icon-inline" />
                ) : error && !data ? (
                    <div className="alert alert-danger">Error: {error.message}</div>
                ) : !data || !data.catalogEntity ? (
                    <div className="alert alert-danger">Entity not found in catalog</div>
                ) : (
                    <EntityDetailContent {...props} entity={data.catalogEntity} telemetryService={telemetryService} />
                )}
            </div>
        </>
    )
}
