import React from 'react'
import { RouteProps, useRouteMatch, matchPath, useLocation } from 'react-router'
import { Link, Switch, Route } from 'react-router-dom'

import { TabList, Tab, Tabs } from '@sourcegraph/wildcard'

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
    const location = useLocation()
    const match = useRouteMatch()
    return (
        <>
            <Tabs
                size="medium"
                defaultIndex={tabs.findIndex(tab =>
                    matchPath(location.pathname, { path: tabPath(match.url, tab), exact: tab.exact })
                )}
                className="mb-2"
            >
                <TabList>
                    {tabs.map(tab => (
                        <Tab
                            key={tab.path}
                            as={Link}
                            to={tabPath(match.url, tab)}
                            data-tab-content={tab.label}
                            className="px-3"
                        >
                            {tab.label}
                        </Tab>
                    ))}
                </TabList>
            </Tabs>
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
