import React from 'react'
import { Link, NavLink } from 'react-router-dom'
import { kebabCase } from 'lodash'

/**
 * A link displaying an icon along with text.
 *
 */
export const LinkWithIcon: React.FunctionComponent<{
    to: string
    text: string
    icon: React.ComponentType<{ className?: string }>
    className?: string
    activeClassName?: string
}> = ({ to, text, icon: Icon, className = '', activeClassName }) => {
    const LinkComponent = activeClassName ? NavLink : Link
    const linkProps = { to, className: `${className} d-flex align-items-center`, activeClassName }
    return (
        <LinkComponent {...linkProps} data-testid={kebabCase(text)}>
            <Icon className="icon-inline" />
            <span className="inline-block ml-1">{text}</span>
        </LinkComponent>
    )
}
