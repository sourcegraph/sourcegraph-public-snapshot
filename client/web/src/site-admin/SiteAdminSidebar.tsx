import React from 'react'

import { Link, Icon } from '@sourcegraph/wildcard'

import { BatchChangesProps } from '../batches'
import { SidebarGroup, SidebarCollapseItems, SidebarNavItem } from '../components/Sidebar'
import { NavGroupDescriptor } from '../util/contributions'

import styles from './SiteAdminSidebar.module.scss'

export interface SiteAdminSideBarGroupContext extends BatchChangesProps {
    isSourcegraphDotCom: boolean
}

export interface SiteAdminSideBarGroup extends NavGroupDescriptor<SiteAdminSideBarGroupContext> {}

export type SiteAdminSideBarGroups = readonly SiteAdminSideBarGroup[]

export interface SiteAdminSidebarProps extends BatchChangesProps {
    isSourcegraphDotCom: boolean
    /** The items for the side bar, by group */
    groups: SiteAdminSideBarGroups
    className?: string
}

/**
 * Sidebar for the site admin area.
 */
export const SiteAdminSidebar: React.FunctionComponent<React.PropsWithChildren<SiteAdminSidebarProps>> = ({
    className,
    groups,
    ...props
}) => (
    <SidebarGroup className={className}>
        <ul className="list-group">
            {groups.map(
                ({ header, items, condition = () => true }, index) =>
                    condition(props) &&
                    (items.length > 1 ? (
                        <li className="p-0 list-group-item" key={index}>
                            <SidebarCollapseItems icon={header?.icon} label={header?.label} openByDefault={true}>
                                {items.map(
                                    ({ label, to, source = 'client', condition = () => true }) =>
                                        condition(props) && (
                                            <SidebarNavItem
                                                to={to}
                                                exact={true}
                                                key={label}
                                                source={source}
                                                className={styles.navItem}
                                            >
                                                {label}
                                            </SidebarNavItem>
                                        )
                                )}
                            </SidebarCollapseItems>
                        </li>
                    ) : (
                        <li className="p-0 list-group-item" key={items[0].label}>
                            <Link to={items[0].to} className="bg-2 border-0 d-flex list-group-item-action p-2 w-100">
                                <span>
                                    {header?.icon && (
                                        <>
                                            <Icon className="sidebar__icon mr-1" as={header.icon} aria-hidden={true} />{' '}
                                        </>
                                    )}
                                    {items[0].label}
                                </span>
                            </Link>
                        </li>
                    ))
            )}
        </ul>
    </SidebarGroup>
)
