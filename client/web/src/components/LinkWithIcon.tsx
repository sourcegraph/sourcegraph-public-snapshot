import React from 'react'
import { Link, NavLink } from 'react-router-dom'

/**
 * A link displaying an icon along with text.
 *
 */
export const LinkWithIcon: React.FunctionComponent<{
    to: string
    icon: React.ComponentType<{ className?: string }>
    className?: string
    activeClassName?: string
    exact?: boolean
    dataTestID?: string
}> = ({ to, children, icon: Icon, className = '', activeClassName, exact, dataTestID }) => {
    const LinkComponent = activeClassName ? NavLink : Link
    const linkProps = { to, exact, className: `${className} d-flex align-items-center`, activeClassName }
    return (
        <LinkComponent {...linkProps} data-testid={dataTestID}>
            <Icon className="icon-inline" />
            <span className="inline-block ml-1">{children}</span>
        </LinkComponent>
    )
}
