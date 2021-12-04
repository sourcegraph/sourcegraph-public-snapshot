import classNames from 'classnames'
import React from 'react'

import { Scalars } from '@sourcegraph/shared/src/graphql-operations'

import { useCatalogEntityFilters } from '../../core/entity-filters'
import { CatalogExplorerList } from '../overview/components/catalog-explorer/CatalogExplorerList'
import { CatalogExplorerViewOptionsRow } from '../overview/components/catalog-explorer/CatalogExplorerViewOptionsRow'

interface Props {
    group: Scalars['ID']
    className?: string
}

export const GroupCatalogExplorer: React.FunctionComponent<Props> = ({ group, className }) => {
    const { filters, onFiltersChange } = useCatalogEntityFilters()

    return (
        <div className={classNames('card', className)}>
            <CatalogExplorerViewOptionsRow
                before={<h4 className="mb-0 mr-2 font-weight-bold">Components</h4>}
                filters={filters}
                onFiltersChange={onFiltersChange}
                className="pl-3 pr-2 py-2 border-bottom"
            />
            <CatalogExplorerList
                filters={filters}
                queryScope={`group:${group}`}
                noBottomBorder={true}
                itemStartClassName="pl-3"
                itemEndClassName="pr-3"
            />
        </div>
    )
}
