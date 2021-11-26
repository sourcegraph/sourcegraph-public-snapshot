import classNames from 'classnames'
import React from 'react'

import { Scalars } from '@sourcegraph/shared/src/graphql-operations'

import { useComponentFilters } from '../../core/component-query'
import { CatalogExplorerList } from '../overview/components/catalog-explorer/CatalogExplorerList'
import { CatalogExplorerViewOptionsRow } from '../overview/components/catalog-explorer/CatalogExplorerViewOptionsRow'
import { ViewModeToggle, useViewModeTemporarySettings } from '../overview/components/catalog-explorer/ViewModeToggle'
import { OverviewEntityGraph } from '../overview/components/overview-content/OverviewEntityGraph'

interface Props {
    group: Scalars['ID']
    className?: string
}

export const GroupCatalogExplorer: React.FunctionComponent<Props> = ({ group, className }) => {
    const filtersProps = useComponentFilters('')

    const [viewMode, setViewMode] = useViewModeTemporarySettings()

    const queryScope = `group:${group}`

    return (
        <div className={classNames('card', className)}>
            <CatalogExplorerViewOptionsRow
                before={<h4 className="mb-0 mr-2 font-weight-bold">Components</h4>}
                toggle={<ViewModeToggle viewMode={viewMode} setViewMode={setViewMode} />}
                {...filtersProps}
                className="pl-3 pr-2 py-2 border-bottom"
            />
            {viewMode === 'list' ? (
                <CatalogExplorerList
                    filters={filtersProps.filters}
                    queryScope={queryScope}
                    noBottomBorder={true}
                    itemStartClassName="pl-3"
                    itemEndClassName="pr-3"
                />
            ) : (
                <OverviewEntityGraph filters={filtersProps.filters} queryScope={queryScope} />
            )}
        </div>
    )
}
