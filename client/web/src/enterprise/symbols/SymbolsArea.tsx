import React, { useCallback, useState } from 'react'
import { Route, RouteComponentProps, Switch } from 'react-router'
import { Resizable } from '../../../../shared/src/components/Resizable'
import { VersionContextProps } from '../../../../shared/src/search/util'
import { SettingsCascadeProps } from '../../../../shared/src/settings/settings'
import { BreadcrumbSetters } from '../../components/Breadcrumbs'
import { RepoHeaderContributionsLifecycleProps } from '../../repo/RepoHeader'
import { RepoRevisionContainerContext } from '../../repo/RepoRevisionContainer'
import { SymbolPage, SymbolRouteProps } from './SymbolPage'
import { SymbolsPage } from './SymbolsPage'
import { SymbolsSidebar, SymbolsSidebarOptions } from './SymbolsSidebar'

interface Props
    extends Pick<RepoRevisionContainerContext, 'repo' | 'revision' | 'resolvedRev'>,
        RouteComponentProps<{}>,
        RepoHeaderContributionsLifecycleProps,
        SettingsCascadeProps,
        BreadcrumbSetters,
        VersionContextProps {
    isLightTheme: boolean
}

export interface SymbolsSidebarOptionsSetterProps {
    setSidebarOptions: (options: SymbolsSidebarOptions | null) => void
}

export const SymbolsArea: React.FunctionComponent<Props> = ({
    match,
    useBreadcrumb: useBreadcrumb,
    repoHeaderContributionsLifecycleProps,
    history,
    ...props
}) => {
    const [sidebarOptions, rawSetSidebarOptions] = useState<SymbolsSidebarOptions | null>(null)
    const setSidebarOptions = useCallback<SymbolsSidebarOptionsSetterProps['setSidebarOptions']>(
        options => rawSetSidebarOptions(options),
        []
    )
    return (
        <>
            {sidebarOptions && (
                <Resizable
                    className="symbols-area__sidebar border-right"
                    handlePosition="right"
                    storageKey="SymbolsSidebar"
                    defaultSize={200}
                    element={<SymbolsSidebar repo={props.repo} {...sidebarOptions} className="w-100 overflow-auto" />}
                />
            )}
            <div>
                <Switch>
                    <Route
                        path={`${match.url}/:symbolID+`}
                        render={(routeProps: RouteComponentProps<SymbolRouteProps>) => (
                            <SymbolPage {...props} {...routeProps} setSidebarOptions={setSidebarOptions} />
                        )}
                    />
                    <Route path={match.url} exact={true} render={(routeProps: any) => <SymbolsPage {...props} />} />
                </Switch>
            </div>
        </>
    )
}
