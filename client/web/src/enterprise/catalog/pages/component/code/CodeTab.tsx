import classNames from 'classnames'
import React from 'react'
import { Route, Switch } from 'react-router'

import { ExtensionsControllerProps } from '@sourcegraph/shared/src/extensions/controller'
import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeProps } from '@sourcegraph/shared/src/theme'

import { SidebarGroup, SidebarNavItem } from '../../../../../components/Sidebar'
import { ComponentStateDetailFields } from '../../../../../graphql-operations'

import { ComponentCodeOwners } from './CodeOwners'
import { ComponentCommits } from '../ComponentCommits'
import { ComponentContributors } from '../ComponentContributors'
import { ComponentSourceDefinitions } from '../ComponentSourceLocations'
import { ComponentSources } from '../ComponentSources'
import { ComponentBranches } from './ComponentBranches'

interface Props extends TelemetryProps, ExtensionsControllerProps, ThemeProps, SettingsCascadeProps {
    component: ComponentStateDetailFields
    className?: string
}

export const CodeTab: React.FunctionComponent<Props> = ({ component, className, ...props }) => {
    const pathPrefix = component.url

    return (
        <div className={classNames('row no-gutters', className)}>
            <div className="col-md-2 col-lg-2 col-xl-1 p-3 pr-md-0">
                <SidebarGroup style={{ width: 'unset' }}>
                    <SidebarNavItem to={`${pathPrefix}/code`} exact={true} className="mb-1">
                        Files
                    </SidebarNavItem>
                    <SidebarNavItem to={`${pathPrefix}/code/commits`} className="mb-1">
                        Commits
                    </SidebarNavItem>
                    <SidebarNavItem to={`${pathPrefix}/code/branches`} className="mb-1">
                        Branches
                    </SidebarNavItem>
                </SidebarGroup>
            </div>
            <Switch>
                <Route path={`${pathPrefix}/code`} exact={true}>
                    <div className="col-md-6 col-lg-7 col-xl-9 p-3 pr-md-0">
                        {component.__typename === 'Component' && (
                            <>
                                <h4 className="sr-only">Sources</h4>
                                <ComponentSourceDefinitions component={component} className="mb-3" />
                            </>
                        )}
                        <h4 className="sr-only">All files</h4>
                        <ComponentSources {...props} component={component} className="mb-3 card p-2" />
                    </div>
                    <div className="col-md-4 col-lg-3 col-xl-2 p-3">
                        <ComponentCodeOwners component={component} className="mb-3" />
                        <ComponentContributors component={component} className="mb-3" />
                    </div>
                </Route>
                <Route path={`${pathPrefix}/code/commits`}>
                    <div className="col-md-10 col-lg-10 col-xl-11 p-3">
                        <h4 className="sr-only">Recent commits</h4>
                        <ComponentCommits component={component} className="mb-3 card p-2" />
                    </div>
                </Route>
                <Route path={`${pathPrefix}/code/branches`}>
                    <div className="col-md-10 col-lg-10 col-xl-11 p-3">
                        <ComponentBranches component={component.id} />
                    </div>
                </Route>
            </Switch>
        </div>
    )
}
