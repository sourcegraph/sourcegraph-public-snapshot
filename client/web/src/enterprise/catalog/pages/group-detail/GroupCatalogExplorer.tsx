import classNames from 'classnames'
import React from 'react'

import { Scalars } from '@sourcegraph/shared/src/graphql-operations'

import { useTemporarySetting } from '../../../../settings/temporary/useTemporarySetting'
import { useCatalogEntityFilters } from '../../core/entity-filters'
import { CatalogExplorerList } from '../overview/components/catalog-explorer/CatalogExplorerList'
import { CatalogExplorerViewOptionsRow } from '../overview/components/catalog-explorer/CatalogExplorerViewOptionsRow'
import { OverviewEntityGraph } from '../overview/components/overview-content/OverviewEntityGraph'

interface Props {
    group: Scalars['ID']
    className?: string
}

export const GroupCatalogExplorer: React.FunctionComponent<Props> = ({ group, className }) => {
    const { filters, onFiltersChange } = useCatalogEntityFilters()

    const [viewMode, setViewMode] = useTemporarySetting('catalog.explorer.viewMode', 'list')

    return (
        <div className={classNames('card', className)}>
            <CatalogExplorerViewOptionsRow
                before={<h4 className="mb-0 mr-2 font-weight-bold">Components</h4>}
                toggle={
                    <div className="btn-group" role="group">
                        <button
                            type="button"
                            className={classNames('btn border', viewMode === 'list' ? 'btn-secondary' : 'text-muted')}
                            onClick={() => setViewMode('list')}
                        >
                            List
                        </button>
                        <button
                            type="button"
                            className={classNames('btn border', viewMode === 'graph' ? 'btn-secondary' : 'text-muted')}
                            onClick={() => setViewMode('graph')}
                        >
                            Graph
                        </button>
                    </div>
                }
                filters={filters}
                onFiltersChange={onFiltersChange}
                className="pl-3 pr-2 py-2 border-bottom"
            />
            {viewMode === 'list' ? (
                <CatalogExplorerList
                    filters={filters}
                    queryScope={`group:${group}`}
                    noBottomBorder={true}
                    itemStartClassName="pl-3"
                    itemEndClassName="pr-3"
                />
            ) : (
                <OverviewEntityGraph filters={filters} queryScope={`group:${group}`} className="border-top" />
            )}
        </div>
    )
}
