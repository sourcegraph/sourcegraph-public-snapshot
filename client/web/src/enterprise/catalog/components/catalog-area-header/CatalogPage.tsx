import classNames from 'classnames'
import React from 'react'
import { Route, RouteProps, Switch, useRouteMatch } from 'react-router'
import { NavLink } from 'react-router-dom'

import { CatalogAreaHeader } from './CatalogAreaHeader'
import styles from './CatalogPage.module.scss'

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
        <div className="flex-1 d-flex flex-column">
            <CatalogAreaHeader
                path={path}
                nav={
                    <ul className="nav nav-tabs" style={{ marginBottom: '-1px' }}>
                        {tabs.map(({ path, exact, text }) => (
                            <li key={path} className="nav-item">
                                <NavLink
                                    to={path ? `${match.url}/${path}` : match.url}
                                    exact={exact}
                                    className={classNames('nav-link px-3', styles.tab)}
                                    data-tab-content={text}
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
        </div>
    )
}
