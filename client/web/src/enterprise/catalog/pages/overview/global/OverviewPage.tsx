import React, { useEffect } from 'react'
import { Route, Switch, useRouteMatch } from 'react-router'
import { NavLink } from 'react-router-dom'

import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { Page } from '@sourcegraph/web/src/components/Page'
import { Container, PageHeader } from '@sourcegraph/wildcard'

import { CatalogIcon } from '../../../../../catalog'
import { CatalogEntityFiltersProps } from '../../../core/entity-filters'
import { EntityList } from '../components/entity-list/EntityList'
import { OverviewEntityGraph } from '../components/overview-content/OverviewEntityGraph'

interface Props extends CatalogEntityFiltersProps, TelemetryProps {}

/**
 * The catalog overview page.
 */
export const OverviewPage: React.FunctionComponent<Props> = ({ filters, onFiltersChange, telemetryService }) => {
    useEffect(() => {
        telemetryService.logViewEvent('CatalogOverview')
    }, [telemetryService])

    const match = useRouteMatch()

    return (
        <Page>
            <PageHeader
                path={[{ icon: CatalogIcon, text: 'Catalog' }]}
                className="mb-4"
                description="Explore software components, services, libraries, APIs, and more."
            />

            <ul className="nav nav-tabs w-100 mb-2">
                <li className="nav-item">
                    <NavLink to={match.url} exact={true} className="nav-link px-3">
                        List
                    </NavLink>
                </li>
                <li className="nav-item">
                    <NavLink to={`${match.url}/graph`} exact={true} className="nav-link px-3">
                        Graph
                    </NavLink>
                </li>
            </ul>

            <Switch>
                <Route path={match.url} exact={true}>
                    <Container className="p-0 mb-2">
                        <EntityList filters={filters} onFiltersChange={onFiltersChange} size="sm" />
                    </Container>
                </Route>
                <Route path={`${match.url}/graph`} exact={true}>
                    <OverviewEntityGraph />
                </Route>
            </Switch>
        </Page>
    )
}
