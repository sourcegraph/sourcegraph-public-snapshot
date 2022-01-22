import React, { useEffect, useState } from 'react'
import { useLocation, Switch, Route } from 'react-router'
import { NavLink } from 'react-router-dom'

import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { Page } from '@sourcegraph/web/src/components/Page'
import { FeedbackBadge, PageHeader } from '@sourcegraph/wildcard'

import { CatalogIcon } from '../../../../catalog'
import { CatalogExplorerGraph } from '../../components/catalog-explorer/graph/CatalogExplorerGraph'
import { CatalogExplorerList } from '../../components/catalog-explorer/list/CatalogExplorerList'
import { CatalogExplorerViewOptionsRow } from '../../components/catalog-explorer/view-options/CatalogExplorerViewOptionsRow'
import { useComponentFilters } from '../../core/component-query'

interface Props extends TelemetryProps {
    basePathname: string
}

/**
 * The global catalog overview page, which lists, searches, and visualizes components.
 */
export const GlobalOverviewPage: React.FunctionComponent<Props> = ({ basePathname, telemetryService }) => {
    useEffect(() => {
        telemetryService.logViewEvent('CatalogGlobalOverview')
    }, [telemetryService])

    return (
        <Page>
            <PageHeader
                path={[{ icon: CatalogIcon, text: 'Catalog' }]}
                className="mb-4"
                description="Explore software components, services, libraries, APIs, and more."
                actions={<FeedbackBadge status="prototype" feedback={{ mailto: 'support@sourcegraph.com' }} />}
            />
            <GlobalCatalogExplorer basePathname={basePathname} />
        </Page>
    )
}

const GlobalCatalogExplorer: React.FunctionComponent<{
    basePathname: string
}> = ({ basePathname }) => {
    const filtersProps = useComponentFilters('')

    const location = useLocation()

    const [tags, setTags] = useState<string[]>()

    return (
        <>
            <CatalogExplorerViewOptionsRow
                toggle={
                    <div className="btn-group" role="group">
                        <NavLink
                            to={{ pathname: basePathname, search: location.search }}
                            exact={true}
                            className="btn border"
                            activeClassName="btn-primary"
                        >
                            List
                        </NavLink>
                        <NavLink
                            to={{ pathname: `${basePathname}/graph`, search: location.search }}
                            exact={true}
                            className="btn border"
                            activeClassName="btn-primary"
                        >
                            Graph
                        </NavLink>
                    </div>
                }
                tags={tags}
                {...filtersProps}
                className="pb-2"
            />
            <Switch>
                <Route path={basePathname} exact={true}>
                    <CatalogExplorerList filters={filtersProps.filters} onTagsChange={setTags} />
                </Route>
                <Route path={`${basePathname}/graph`} exact={true}>
                    <CatalogExplorerGraph filters={filtersProps.filters} className="border-top" />
                </Route>
            </Switch>
        </>
    )
}
