import React from 'react'
import { useLocation, useRouteMatch } from 'react-router'
import { Switch, Route, NavLink } from 'react-router-dom'

import { CatalogHealthTable } from '../../../../components/catalog-health-table/CatalogHealthTable'
import { useCatalogEntityFilters } from '../../../../core/entity-filters'
import { OverviewEntityGraph } from '../overview-content/OverviewEntityGraph'

import { CatalogExplorerList } from './CatalogExplorerList'
import { CatalogExplorerViewOptionsRow } from './CatalogExplorerViewOptionsRow'

interface Props {}

export const CatalogExplorer: React.FunctionComponent<Props> = () => {
    const filtersProps = useCatalogEntityFilters('is:component')

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
                        <NavLink
                            to={{ pathname: '/catalog/health', search: location.search }}
                            exact={true}
                            className="btn border"
                            activeClassName="btn-primary"
                        >
                            Health
                        </NavLink>
                    </div>
                }
                {...filtersProps}
                className="pb-2"
            />
            <Switch>
                <Route path={match.path} exact={true}>
                    <CatalogExplorerList filters={filtersProps.filters} />
                </Route>
                <Route path={`${match.path}/graph`} exact={true}>
                    <OverviewEntityGraph filters={filtersProps.filters} className="border-top" />
                </Route>
                <Route path={`${match.path}/health`} exact={true}>
                    <CatalogHealthTable filters={filtersProps.filters} />
                </Route>
            </Switch>
        </>
    )
}
