import classNames from 'classnames'
import MenuDownIcon from 'mdi-react/MenuDownIcon'
import MenuUpIcon from 'mdi-react/MenuUpIcon'
import React, { useCallback, useState } from 'react'
import { useRouteMatch } from 'react-router-dom'
import { Collapse } from 'reactstrap'

import { Button, Link } from '@sourcegraph/wildcard'

import styles from './Sidebar.module.scss'

/**
 * Item of `SideBarGroup`.
 */
export const SidebarNavItem: React.FunctionComponent<{
    to: string
    className?: string
    exact?: boolean
    source?: string
}> = ({ children, className, to, exact, source }) => {
    const buttonClassNames = classNames('text-left d-flex', styles.linkInactive, className)
    const routeMatch = useRouteMatch(to)

    if (source === 'server') {
        return (
            <Button as="a" href={to} className={buttonClassNames}>
                {children}
            </Button>
        )
    }

    return (
        <Button to={to} className={buttonClassNames} variant={routeMatch?.isExact ? 'primary' : undefined} as={Link}>
            {children}
        </Button>
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
                    <MenuUpIcon className={classNames('icon-inline', styles.chevron)} />
                ) : (
                    <MenuDownIcon className={classNames('icon-inline', styles.chevron)} />
                )}
            </button>
            <Collapse id={label} isOpen={isOpen} className="border-top">
                {children}
            </Collapse>
        </>
    )
}

interface SidebarGroupProps {
    className?: string
}

/**
 * A box of items in the side bar. Use `SideBarGroupHeader` as children.
 */
export const SidebarGroup: React.FunctionComponent<SidebarGroupProps> = ({ children, className }) => (
    <div className={classNames('mb-3', styles.sidebar, className)}>{children}</div>
)
