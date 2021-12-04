import classNames from 'classnames'
import React from 'react'

import { Scalars } from '@sourcegraph/shared/src/graphql-operations'

import { useCatalogEntityFilters } from '../../../core/entity-filters'
import { CatalogExplorerList } from '../../overview/components/catalog-explorer/CatalogExplorerList'
import { CatalogExplorerViewOptionsRow } from '../../overview/components/catalog-explorer/CatalogExplorerViewOptionsRow'
import { useViewModeTemporarySettings, ViewModeToggle } from '../../overview/components/catalog-explorer/ViewModeToggle'
import { OverviewEntityGraph } from '../../overview/components/overview-content/OverviewEntityGraph'

interface Props {
    className?: string
}

interface Props {
    entity: Scalars['ID']
    className?: string
}

export const EntityCatalogExplorer: React.FunctionComponent<Props> = ({ entity, className }) => {
    const { filters, onFiltersChange } = useCatalogEntityFilters()

    const [viewMode, setViewMode] = useViewModeTemporarySettings()

    const queryScope = `relatedToEntity:${entity}`

    return (
        <div className={classNames('card', className)}>
            <CatalogExplorerViewOptionsRow
                before={<h4 className="mb-0 mr-2 font-weight-bold">Relations</h4>}
                toggle={<ViewModeToggle viewMode={viewMode} setViewMode={setViewMode} />}
                filters={filters}
                onFiltersChange={onFiltersChange}
                className="pl-3 pr-2 py-2 border-bottom"
            />
            {viewMode === 'list' ? (
                <CatalogExplorerList
                    filters={filters}
                    queryScope={queryScope}
                    noBottomBorder={true}
                    itemStartClassName="pl-3"
                    itemEndClassName="pr-3"
                />
            ) : (
                <OverviewEntityGraph
                    filters={filters}
                    queryScope={queryScope}
                    highlightID={entity}
                    errorClassName="p-3"
                />
            )}
        </div>
    )
}
