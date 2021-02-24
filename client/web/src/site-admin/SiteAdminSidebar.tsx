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
        <div className="border">
            {groups.map(
                ({ header, items, condition = () => true }, index) =>
                    condition({}) &&
                    (items.length > 1 ? (
                        <SidebarCollapseItems
                            icon={header?.icon}
                            label={header?.label}
                            key={index}
                            openByDefault={false}
                        >
                            <SidebarGroupItems>
                                {items.map(
                                    ({ label, to, exact, condition = () => true }) =>
                                        condition({}) && (
                                            <SidebarNavItem to={to} exact={exact} key={label}>
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
                            className="btn sidebar__btn d-flex justify-content-between w-100 px-2 btn btn-outline-secondary"
                        >
                            <span>
                                {header?.icon && <header.icon className="sidebar__icon icon-inline mr-1" />}{' '}
                                {items[0].label}
                            </span>
                        </Link>
                    ))
            )}
        </div>
    </div>
)
