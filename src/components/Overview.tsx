import * as React from 'react'
import { Link } from 'react-router-dom'

/**
 * A container for multiple OverviewItem components.
 */
export const OverviewList: React.SFC<{ children: React.ReactNode | React.ReactNode[] }> = ({ children }) => (
    <ul className="overview-list">{children}</ul>
)

/**
 * A row item used for an overview page, with an icon, linked elements, and right-hand actions.
 */
export const OverviewItem: React.SFC<{
    link?: string
    children: React.ReactNode | React.ReactNode[]
    actions?: React.ReactFragment
    icon?: React.ComponentType<{ className?: string }>
}> = ({ link, children, actions, icon: Icon }) => {
    let e: React.ReactFragment = (
        <>
            {Icon && <Icon className="icon-inline overview-item__header-icon" />}
            {children}
        </>
    )
    if (link !== undefined) {
        e = (
            <Link to={link} className="overview-item__header-link">
                {e}
            </Link>
        )
    }

    return (
        <div className="overview-item">
            <div className="overview-item__header">{e}</div>
            {actions && <div className="overview-item__actions">{actions}</div>}
        </div>
    )
}
