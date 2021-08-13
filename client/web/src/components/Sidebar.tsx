import classNames from 'classnames'
import MenuDownIcon from 'mdi-react/MenuDownIcon'
import MenuUpIcon from 'mdi-react/MenuUpIcon'
import React, { useCallback, useState } from 'react'
import { NavLink } from 'react-router-dom'
import { Collapse } from 'reactstrap'

export const SIDEBAR_BUTTON_CLASS = 'btn text-left sidebar__link--inactive w-100'

/**
 * Item of `SideBarGroup`.
 */
export const SidebarNavItem: React.FunctionComponent<{
    to: string
    className?: string
    exact?: boolean
    source?: string
}> = ({ children, className, to, exact, source }) => {
    const buttonClassNames = 'btn text-left sidebar__link--inactive d-flex sidebar-nav-link'

    if (source === 'server') {
        return (
            <a href={to} className={classNames(buttonClassNames, className)}>
                {children}
            </a>
        )
    }

    return (
        <NavLink
            to={to}
            exact={exact}
            className={classNames(buttonClassNames, className)}
            activeClassName="btn-primary"
        >
            {children}
        </NavLink>
    )
}
/**
 *
 * Header of a `SideBarGroup`
 */
export const SidebarGroupHeader: React.FunctionComponent<{ label: string }> = ({ label }) => <h3>{label}</h3>

/**
 * Sidebar with collapsible items
 */
export const SidebarCollapseItems: React.FunctionComponent<{
    children: React.ReactNode
    icon?: React.ComponentType<{ className?: string }>
    label?: string
    openByDefault?: boolean
}> = ({ children, label, icon: Icon, openByDefault = false }) => {
    const [isOpen, setOpen] = useState<boolean>(openByDefault)
    const handleOpen = useCallback(() => setOpen(!isOpen), [isOpen])
    return (
        <>
            <button
                aria-expanded={isOpen}
                aria-controls={label}
                type="button"
                onClick={handleOpen}
                className="bg-2 border-0 d-flex justify-content-between list-group-item-action py-2 w-100"
            >
                <span>
                    {Icon && <Icon className="icon-inline mr-1" />} {label}
                </span>
                {isOpen ? (
                    <MenuUpIcon className="sidebar__chevron icon-inline" />
                ) : (
                    <MenuDownIcon className="sidebar__chevron icon-inline" />
                )}
            </button>
            <Collapse id={label} isOpen={isOpen} className="border-top">
                {children}
            </Collapse>
        </>
    )
}

/**
 * A box of items in the side bar. Use `SideBarGroupHeader` as children.
 */
export const SidebarGroup: React.FunctionComponent = ({ children }) => <div className="mb-3 sidebar">{children}</div>
