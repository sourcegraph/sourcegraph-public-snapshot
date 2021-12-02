import React from 'react'
import { RouteProps, useRouteMatch } from 'react-router'
import { NavLink, Switch, Route } from 'react-router-dom'

import { Tab } from '@sourcegraph/wildcard'

interface Props {
    tabs: Tab[]
}

interface Tab extends Pick<RouteProps, 'path' | 'exact'> {
    path: string
    label: string
    icon?: React.ComponentType<{ className?: string }>
    element: JSX.Element
}

export const TabRouter: React.FunctionComponent<Props> = ({ tabs }) => {
    const match = useRouteMatch()
    return (
        <>
            <ul className="nav nav-tabs w-100 mb-3">
                {tabs.map(tab => (
                    <li key={tab.path} className="nav-item">
                        <NavLink
                            to={tabPath(match.url, tab)}
                            exact={tab.exact}
                            className="nav-link px-3"
                            // TODO(sqs): hack so that active items when bolded don't shift the ones to the right over by a few px because bold text is wider
                            style={{ minWidth: '6rem' }}
                        >
                            {tab.label}
                        </NavLink>
                    </li>
                ))}
            </ul>
            <Switch>
                {tabs.map(tab => (
                    <Route key={tab.path} path={tabPath(match.url, tab)} exact={tab.exact}>
                        {tab.element}
                    </Route>
                ))}
            </Switch>
        </>
    )
}

function tabPath(basePath: string, tab: Pick<Tab, 'path'>): string {
    return tab.path ? `${basePath}/${tab.path}` : basePath
}
