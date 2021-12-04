import React from 'react'
import { useLocation, useRouteMatch } from 'react-router'
import { Switch, Route, NavLink } from 'react-router-dom'

import { useCatalogEntityFilters } from '../../../../core/entity-filters'
import { OverviewEntityGraph } from '../overview-content/OverviewEntityGraph'

import { CatalogExplorerList } from './CatalogExplorerList'
import { CatalogExplorerViewOptionsRow } from './CatalogExplorerViewOptionsRow'

interface Props {}

export const CatalogExplorer: React.FunctionComponent<Props> = () => {
    const { filters, onFiltersChange } = useCatalogEntityFilters()

    const match = useRouteMatch()
    const location = useLocation()

    return (
        <>
            <CatalogExplorerViewOptionsRow
                toggle={
                    <div className="btn-group" role="group">
                        <NavLink
                            to={{ pathname: '/catalog', search: location.search }}
                            exact={true}
                            className="btn border"
                            activeClassName="btn-primary"
                        >
                            List
                        </NavLink>
                        <NavLink
                            to={{ pathname: '/catalog/graph', search: location.search }}
                            exact={true}
                            className="btn border"
                            activeClassName="btn-primary"
                        >
                            Graph
                        </NavLink>
                    </div>
                }
                filters={filters}
                onFiltersChange={onFiltersChange}
                className="pb-2"
            />
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
