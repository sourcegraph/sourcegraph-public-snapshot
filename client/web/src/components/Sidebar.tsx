import classNames from 'classnames'
import MenuDownIcon from 'mdi-react/MenuDownIcon'
import MenuUpIcon from 'mdi-react/MenuUpIcon'
import React, { useCallback, useState } from 'react'
import { NavLink } from 'react-router-dom'
import { Button, Collapse } from 'reactstrap'

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
 * Sidebar with collapsible items
 */
export const SidebarCollapseItems: React.FunctionComponent<{
    children: JSX.Element
    icon?: React.ComponentType<{ className?: string }>
    label?: string
    openByDefault?: boolean
}> = ({ children, label, icon: Icon, openByDefault = false }) => {
    const [isOpen, setOpen] = useState<boolean>(openByDefault)
    const handleOpen = useCallback(() => setOpen(!isOpen), [isOpen])
    return (
        <>
            <Button
                color="secondary"
                outline={true}
                onClick={handleOpen}
                className={classNames(
                    { 'border-bottom-0': !isOpen },
                    'btn sidebar__btn d-flex justify-content-between w-100 px-2 border'
                )}
            >
                <span>
                    {Icon && <Icon className="sidebar__icon icon-inline mr-1" />} {label}
                </span>
                {isOpen ? <MenuUpIcon className="icon-inline" /> : <MenuDownIcon className="icon-inline" />}
            </Button>
            <Collapse isOpen={isOpen}>{children}</Collapse>
        </>
    )
}

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
