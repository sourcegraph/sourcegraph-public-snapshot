import classNames from 'classnames'
import React from 'react'

import { Scalars } from '@sourcegraph/shared/src/graphql-operations'

import { CatalogOverviewGraph } from '../../components/catalog-overview/graph/CatalogOverviewGraph'
import { CatalogExplorerRelationList } from '../../components/catalog-overview/relation-list/CatalogExplorerRelationList'
import { CatalogExplorerViewOptionsRow } from '../../components/catalog-overview/view-options/CatalogExplorerViewOptionsRow'
import {
    useViewModeTemporarySettings,
    ViewModeToggle,
} from '../../components/catalog-overview/view-options/ViewModeToggle'
import { useComponentFilters } from '../../core/component-query'

interface Props {
    component: Scalars['ID']
    useURLForConnectionParams?: boolean
    className?: string
}

export const CatalogExplorer: React.FunctionComponent<Props> = ({
    component,
    useURLForConnectionParams,
    className,
}) => {
    const filtersProps = useComponentFilters('')

    const [viewMode, setViewMode] = useViewModeTemporarySettings()

    const queryScope = `relatedToEntity:${component}`

    return (
        <div className={classNames('card', className)}>
            <CatalogExplorerViewOptionsRow
                before={<h4 className="mb-0 mr-2 font-weight-bold">Relations</h4>}
                toggle={<ViewModeToggle viewMode={viewMode} setViewMode={setViewMode} />}
                {...filtersProps}
                className="pl-3 pr-2 py-2 border-bottom"
            />
            {viewMode === 'list' ? (
                <CatalogExplorerRelationList
                    component={component}
                    useURLForConnectionParams={useURLForConnectionParams}
                    filters={filtersProps.filters}
                    queryScope={queryScope}
                    noBottomBorder={true}
                    itemStartClassName="pl-3"
                    itemEndClassName="pr-3"
                />
            ) : (
                <CatalogOverviewGraph
                    filters={filtersProps.filters}
                    queryScope={queryScope}
                    highlightID={component}
                    errorClassName="p-3"
                />
            )}
        </div>
    )
}
