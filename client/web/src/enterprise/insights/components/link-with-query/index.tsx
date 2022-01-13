import React from 'react'
import { useLocation } from 'react-router-dom'

import { Link } from '@sourcegraph/wildcard'
import type { LinkProps } from '@sourcegraph/wildcard/src/components/Link'

export interface LinkWithQueryProps extends Omit<LinkProps, 'to'> {
    to: string
}

/**
 * Renders react router link component with query params preserving between route transitions.
 */
export const LinkWithQuery: React.FunctionComponent<LinkWithQueryProps> = props => {
    const { children, to, ...otherProps } = props
    const { search } = useLocation()

    return (
        <Link to={to + search} {...otherProps}>
            {children}
        </Link>
    )
}
