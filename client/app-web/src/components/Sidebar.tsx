import React from 'react'
import { NavLink } from 'react-router-dom'
import classNames from 'classnames'

export const SIDEBAR_BUTTON_CLASS = 'btn btn-secondary d-block w-100 my-2'

/**
 * Item of `SideBarGroupItems`.
 */
export const SidebarNavItem: React.FunctionComponent<{
    to: string
    icon?: React.ComponentType<{ className?: string }>
    className?: string
    exact?: boolean
}> = ({ icon: Icon, children, className, to, exact }) => (
    <NavLink to={to} exact={exact} className={classNames('list-group-item list-group-item-action py-2', className)}>
        {Icon && <Icon className="icon-inline mr-2" />}
        {children}
    </NavLink>
)

/**
 * Header of a `SideBarGroup`
 */
export const SidebarGroupHeader: React.FunctionComponent<{
    icon?: React.ComponentType<{ className?: string }>
    label: string
    children?: undefined
}> = ({ icon: Icon, label }) => (
    <div className="card-header">
        {Icon && <Icon className="icon-inline mr-1" />} {label}
    </div>
)

/**
 * A box of items in the side bar. Use `SideBarGroupHeader` and `SideBarGroupItems` as children.
 */
export const SidebarGroup: React.FunctionComponent<{}> = ({ children }) => (
    <div className="card mb-3 sidebar">{children}</div>
)

/**
 * Container for all `SideBarNavItem` in a `SideBarGroup`.
 */
export const SidebarGroupItems: React.FunctionComponent<{}> = ({ children }) => (
    <div className="list-group list-group-flush">{children}</div>
)
