import React, { useEffect, useMemo } from 'react'
import { Route, RouteComponentProps, Switch } from 'react-router'
import { Link } from 'react-router-dom'

import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'

import { BreadcrumbSetters } from '../../components/Breadcrumbs'
import { RepoHeaderContributionsLifecycleProps } from '../../repo/RepoHeader'
import { RepoRevisionContainerContext } from '../../repo/RepoRevisionContainer'
import { eventLogger } from '../../tracking/eventLogger'

import { SymbolPage, SymbolRouteProps, ContextPage } from './ContextPage'
import { useContextViewOptions } from './useContextViewOptions'

interface Props
    extends Pick<RepoRevisionContainerContext, 'repo' | 'revision' | 'resolvedRev'>,
        RouteComponentProps<{}>,
        RepoHeaderContributionsLifecycleProps,
        SettingsCascadeProps,
        BreadcrumbSetters {}

export const ContextArea: React.FunctionComponent<Props> = ({
    match,
    useBreadcrumb: useBreadcrumb,
    repoHeaderContributionsLifecycleProps,
    history,
    ...props
}) => {
    useEffect(() => {
        eventLogger.logViewEvent('Context')
    }, [])

    useBreadcrumb = useBreadcrumb(
        useMemo(() => ({ key: 'context', element: <Link to={match.url}>Context</Link> }), [match.url])
    ).useBreadcrumb

    const { viewOptions } = useContextViewOptions(props)

    return (
        <>
            <div style={{ overflow: 'auto' }} className="w-100">
                <Switch>
                    {/* eslint-disable react/jsx-no-bind */}
                    <Route
                        path={match.url}
                        exact={true}
                        render={(routeProps: RouteComponentProps<SymbolRouteProps>) => (
                            <ContextPage
                                {...props}
                                {...routeProps}
                                viewOptions={viewOptions}
                                setSidebarOptions={setSidebarOptions}
                            />
                        )}
                    />
                    <Route
                        path={`${match.url}/symbol/:scheme/:identifier+`}
                        sensitive={true}
                        render={(routeProps: RouteComponentProps<SymbolRouteProps>) => (
                            <SymbolPage
                                {...props}
                                {...routeProps}
                                useBreadcrumb={useBreadcrumb}
                                viewOptions={viewOptions}
                                setSidebarOptions={setSidebarOptions}
                            />
                        )}
                    />
                    {/* eslint-enable react/jsx-no-bind */}
                </Switch>
            </div>
        </>
    )
}
