import React, { useMemo } from 'react'
import { Route, RouteComponentProps, Switch } from 'react-router'
import { Link } from 'react-router-dom'

import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'

import { BreadcrumbSetters } from '../../components/Breadcrumbs'
import { RepoHeaderContributionsLifecycleProps } from '../../repo/RepoHeader'
import { RepoRevisionContainerContext } from '../../repo/RepoRevisionContainer'

import { UsageHelpPage } from './UsageHelpPage'
import { UsagePage } from './UsagePage'

interface Props
    extends Pick<RepoRevisionContainerContext, 'repo' | 'revision' | 'resolvedRev'>,
        RouteComponentProps<{}>,
        RepoHeaderContributionsLifecycleProps,
        SettingsCascadeProps,
        BreadcrumbSetters {}

export const UsageArea: React.FunctionComponent<Props> = ({
    match,
    useBreadcrumb: useBreadcrumb,
    repoHeaderContributionsLifecycleProps,
    history,
    ...props
}) => {
    useBreadcrumb = useBreadcrumb(
        useMemo(() => ({ key: 'guide', element: <Link to={match.url}>Usage</Link> }), [match.url])
    ).useBreadcrumb

    return (
        <>
            <div className="overflow-auto w-100">
                <Switch>
                    {/* eslint-disable react/jsx-no-bind */}
                    <Route path={match.url} exact={true}>
                        <UsageHelpPage />
                    </Route>
                    <Route
                        path={`${match.url}/symbol/:scheme/:identifier+`}
                        sensitive={true}
                        render={(routeProps: RouteComponentProps<UsageRouteProps>) => (
                            <UsagePage {...props} {...routeProps} useBreadcrumb={useBreadcrumb} />
                        )}
                    />
                    <Route>
                        <p>Not found</p>
                    </Route>
                    {/* eslint-enable react/jsx-no-bind */}
                </Switch>
            </div>
        </>
    )
}
