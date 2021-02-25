import React from 'react'
import { Link } from 'react-router-dom'
import { SidebarCollapseItems, SidebarGroupItems, SidebarNavItem } from '../components/Sidebar'
import { NavGroupDescriptor } from '../util/contributions'

export interface SiteAdminSideBarGroup extends NavGroupDescriptor {}

export type SiteAdminSideBarGroups = readonly SiteAdminSideBarGroup[]

export interface SiteAdminSidebarProps {
    /** The items for the side bar, by group */
    groups: SiteAdminSideBarGroups
    className: string
}

/**
 * Sidebar for the site admin area.
 */
export const SiteAdminSidebar: React.FunctionComponent<SiteAdminSidebarProps> = ({ className, groups }) => (
    <div className={`site-admin-sidebar ${className}`}>
        {groups.map(
            ({ header, items, condition = () => true }, index) =>
                condition({}) &&
                (items.length > 1 ? (
                    <SidebarCollapseItems icon={header?.icon} label={header?.label} key={index} openByDefault={true}>
                        <SidebarGroupItems>
                            {items.map(
                                ({ label, to, exact, condition = () => true }) =>
                                    condition({}) && (
                                        <SidebarNavItem
                                            to={to}
                                            exact={exact}
                                            key={label}
                                            className="border-left border-right"
                                        >
                                            {label}
                                        </SidebarNavItem>
                                    )
                            )}
                        </SidebarGroupItems>
                    </SidebarCollapseItems>
                ) : (
                    <Link
                        key={items[0].label}
                        to={items[0].to}
                        className="btn btn-outline-secondary sidebar__btn sidebar__btn--link d-flex w-100 px-2 border"
                    >
                        <span>
                            {header?.icon && <header.icon className="sidebar__icon icon-inline mr-1" />}{' '}
                            {items[0].label}
                        </span>
                    </Link>
                ))
        )}
    </div>
)
