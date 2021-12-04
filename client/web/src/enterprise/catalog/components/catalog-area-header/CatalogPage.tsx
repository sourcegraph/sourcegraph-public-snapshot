import React from 'react'
import { Route, RouteProps, Switch, useRouteMatch } from 'react-router'
import { NavLink } from 'react-router-dom'

import { CatalogAreaHeader } from './CatalogAreaHeader'

interface Tab extends Pick<RouteProps, 'path' | 'exact'> {
    path: string
    text: string
    content: React.ReactFragment
}

interface Props {
    path: React.ComponentProps<typeof CatalogAreaHeader>['path']
    tabs: Tab[]
}

export const CatalogPage: React.FunctionComponent<Props> = ({ path, tabs }) => {
    const match = useRouteMatch()
    return (
        <>
            <CatalogAreaHeader
                path={path}
                nav={
                    <ul className="nav nav-tabs" style={{ marginBottom: '-1px' }}>
                        {tabs.map(({ path, exact, text }) => (
                            <li key={path} className="nav-item">
                                <NavLink
                                    to={path ? `${match.url}/${path}` : match.url}
                                    exact={exact}
                                    className="nav-link px-3"
                                    // TODO(sqs): hack so that active items when bolded don't shift the ones to the right over by a few px because bold text is wider
                                    style={{ minWidth: '6rem' }}
                                >
                                    {text}
                                </NavLink>
                            </li>
                        ))}
                    </ul>
                }
            />
            <Switch>
                {tabs.map(({ path, exact, content }) => (
                    <Route key={path} path={path ? `${match.url}/${path}` : match.url} exact={exact}>
                        {content}
                    </Route>
                ))}
            </Switch>
        </>
    )
}
