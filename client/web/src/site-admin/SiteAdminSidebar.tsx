import React from 'react'
import { Link } from '@sourcegraph/shared/src/components/Link'

import { BatchChangesProps } from '../batches'
import { SidebarGroup, SidebarGroupHeader, SidebarNavItem } from '../components/Sidebar'
import { NavGroupDescriptor } from '../util/contributions'

export interface SiteAdminSideBarGroupContext extends BatchChangesProps {
    isSourcegraphDotCom: boolean
}

export interface SiteAdminSideBarGroup extends NavGroupDescriptor<SiteAdminSideBarGroupContext> {}

export type SiteAdminSideBarGroups = readonly SiteAdminSideBarGroup[]

export interface SiteAdminSidebarProps extends BatchChangesProps {
    isSourcegraphDotCom: boolean
    /** The items for the side bar, by group */
    groups: SiteAdminSideBarGroups
    className: string
}

/**
 * Sidebar for the site admin area.
 */
export const SiteAdminSidebar: React.FunctionComponent<SiteAdminSidebarProps> = ({ className, groups, ...props }) => (
    <div className={`site-admin-sidebar ${className}`}>
        {groups.map(
            ({ header, items, condition = () => true }, index) =>
                condition(props) && (
                    <SidebarGroup key={index}>
                        <SidebarGroupHeader label={header.label} />
                        {items.map(
                            ({ label, to, source = 'client', condition = () => true }) =>
                                condition(props) && (
                                    <SidebarNavItem to={to} exact={true} key={label} source={source}>
                                        {label}
                                    </SidebarNavItem>
                                )
                        )}
                    </SidebarGroup>
                )
        )}
        <Link to="/api/console">API console</Link>
    </div>
)
