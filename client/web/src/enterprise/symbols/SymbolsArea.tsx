import React, { useCallback, useMemo, useState } from 'react'
import { Route, RouteComponentProps, Switch } from 'react-router'
import { Link } from 'react-router-dom'
import { Resizable } from '../../../../shared/src/components/Resizable'
import { VersionContextProps } from '../../../../shared/src/search/util'
import { SettingsCascadeProps } from '../../../../shared/src/settings/settings'
import { BreadcrumbSetters } from '../../components/Breadcrumbs'
import { RepoHeaderContributionsLifecycleProps } from '../../repo/RepoHeader'
import { RepoHeaderContributionPortal } from '../../repo/RepoHeaderContributionPortal'
import { RepoRevisionContainerContext } from '../../repo/RepoRevisionContainer'
import { SymbolPage, SymbolRouteProps } from './SymbolPage'
import { SymbolsPage } from './SymbolsPage'
// import { SymbolsPage } from './SymbolsPage'
import { SymbolsSidebar, SymbolsSidebarOptions } from './SymbolsSidebar'
import { SymbolsExternalsViewOptionToggle, SymbolsInternalsViewOptionToggle } from './SymbolsViewOptionsButtons'
import { useSymbolsViewOptions } from './useSymbolsViewOptions'

interface Props
    extends Pick<RepoRevisionContainerContext, 'repo' | 'revision' | 'resolvedRev'>,
        RouteComponentProps<{}>,
        RepoHeaderContributionsLifecycleProps,
        SettingsCascadeProps,
        BreadcrumbSetters,
        VersionContextProps {
    isLightTheme: boolean
}

export const SymbolsArea: React.FunctionComponent<Props> = ({
    match,
    useBreadcrumb: useBreadcrumb,
    repoHeaderContributionsLifecycleProps,
    history,
    ...props
}) => {
    const sidebarOptions = {}

    return (
        <>
            {sidebarOptions && (
                <Resizable
                    className="symbols-area__sidebar border-right"
                    handlePosition="right"
                    storageKey="SymbolsSidebar"
                    defaultSize={200 /* px */}
                    element={
                        <SymbolsSidebar {...sidebarOptions} allSymbolsURL={match.url} className="w-100 overflow-auto" />
                    }
                />
            )}
            <div>
                <Switch>
                    <Route
                        path={`${match.url}/:symbolID`}
                        render={(routeProps: RouteComponentProps<SymbolRouteProps>) => (
                            <SymbolPage {...props} {...routeProps} />
                        )}
                    />
                    <Route path={match.url} exact={true} render={(routeProps: any) => <SymbolsPage {...props} />} />
                </Switch>
            </div>
        </>
    )
}
