import classNames from 'classnames'
import React from 'react'
import { matchPath, RouteProps, useLocation, useRouteMatch } from 'react-router'
import { NavLink, Switch, Route } from 'react-router-dom'

import { CatalogAreaHeader } from './CatalogAreaHeader'
import styles from './CatalogPage.module.scss'

interface Tab extends Pick<RouteProps, 'path' | 'exact'> {
    path: string | string[]
    text: string
    content: React.ReactFragment
}

interface Props {
    path: React.ComponentProps<typeof CatalogAreaHeader>['path']
    tabs: Tab[]
    actions?: React.ReactFragment
}

export const CatalogPage: React.FunctionComponent<Props> = ({ path, tabs, actions }) => (
    <CatalogPage2 header={<CatalogAreaHeader path={path} actions={actions} />} tabs={tabs} actions={actions} />
)

function pathWithPrefix(path: string | string[], prefix: string, separator: string): string[] {
    return toPaths(path).map(path => (path ? `${prefix}${separator}${path}` : prefix))
}

export const CatalogPage2: React.FunctionComponent<{
    header?: React.ReactFragment
    tabs: Tab[]
    useHash?: boolean
    tabsClassName?: string
}> = ({ header, tabs, useHash, tabsClassName }) => {
    const match = useRouteMatch()
    const location = useLocation()
    const separator = useHash ? '#' : '/'
    return (
        <div className="flex-1 d-flex flex-column w-100">
            {header}
            <ul className={classNames('nav nav-tabs', tabsClassName)} style={{ marginBottom: '-1px' }}>
                {tabs.map(({ path, exact, text }) => (
                    <li key={Array.isArray(path) ? path[0] : path} className="nav-item">
                        <NavLink
                            to={pathWithPrefix(path, match.url, separator)[0]}
                            isActive={() =>
                                Boolean(
                                    useHash
                                        ? pathMatchesHash(path, location.hash)
                                        : matchPath(location.pathname, {
                                              path: pathWithPrefix(path, match.url, separator),
                                              exact,
                                          })
                                )
                            }
                            exact={exact}
                            className={classNames('nav-link px-3', styles.tab)}
                            data-tab-content={text}
                        >
                            {text}
                        </NavLink>
                    </li>
                ))}
            </ul>
            {/* TODO(sqs): hack to make the router work with hashes */}
            <Switch location={useHash ? { ...location, pathname: location.pathname + location.hash } : undefined}>
                {tabs.map(({ path, exact, content }) => (
                    <Route
                        key={Array.isArray(path) ? path[0] : path}
                        path={pathWithPrefix(path, match.url, separator)}
                        exact={exact}
                    >
                        {content}
                    </Route>
                ))}
            </Switch>
        </div>
    )
}
function toPaths(path: string | string[]): string[] {
    return Array.isArray(path) ? path : [path]
}
function pathMatchesHash(path: string | string[], hash: string): boolean {
    return toPaths(path).some(path => (path === '' && hash === '') || hash === `#${path}`)
}
