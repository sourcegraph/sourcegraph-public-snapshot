import React from 'react'
import { useRouteMatch } from 'react-router'
import { Switch, Route } from 'react-router-dom'

import { useCatalogEntityFilters } from '../../../../core/entity-filters'
import { OverviewEntityGraph } from '../overview-content/OverviewEntityGraph'

import { CatalogExplorerList } from './CatalogExplorerList'
import { CatalogExplorerViewOptionsRow } from './CatalogExplorerViewOptionsRow'

interface Props {}

export const CatalogExplorer: React.FunctionComponent<Props> = () => {
    const { filters, onFiltersChange } = useCatalogEntityFilters()

    const match = useRouteMatch()

    return (
        <>
            <CatalogExplorerViewOptionsRow filters={filters} onFiltersChange={onFiltersChange} className="pb-2" />
            <Switch>
                <Route path={match.path} exact={true}>
                    <CatalogExplorerList filters={filters} />
                </Route>
                <Route path={`${match.path}/graph`} exact={true}>
                    <OverviewEntityGraph className="border-top" />
                </Route>
            </Switch>
        </>
    )
}
