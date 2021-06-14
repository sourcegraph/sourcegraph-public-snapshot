import * as React from 'react'
import { RouteComponentProps } from 'react-router-dom'

import { SidebarGroupItems, SidebarGroupHeader, SidebarGroup, SidebarNavItem } from '../../components/Sidebar'
import { SettingsAreaRepositoryFields } from '../../graphql-operations'
import { NavGroupDescriptor } from '../../util/contributions'

export interface RepoSettingsSideBarGroup extends NavGroupDescriptor {}

export type RepoSettingsSideBarGroups = readonly RepoSettingsSideBarGroup[]

interface Props extends RouteComponentProps<{}> {
    repoSettingsSidebarGroups: RepoSettingsSideBarGroups
    className?: string
    repo: SettingsAreaRepositoryFields
}

/**
 * Sidebar for repository settings pages.
 */
export const RepoSettingsSidebar: React.FunctionComponent<Props> = ({
    repoSettingsSidebarGroups,
    className,
    repo,
}: Props) => (
    <div className={className}>
        {repoSettingsSidebarGroups.map(
            ({ header, items, condition = () => true }, index) =>
                condition({}) && (
                    <SidebarGroup key={index}>
                        {header && <SidebarGroupHeader icon={header.icon} label={header.label} />}
                        <SidebarGroupItems>
                            {items.map(
                                ({ label, to, exact, condition = () => true }) =>
                                    condition({}) && (
                                        <SidebarNavItem to={`${repo.url}/-/settings${to}`} exact={exact} key={label}>
                                            {label}
                                        </SidebarNavItem>
                                    )
                            )}
                        </SidebarGroupItems>
                    </SidebarGroup>
                )
        )}
    </div>
)
