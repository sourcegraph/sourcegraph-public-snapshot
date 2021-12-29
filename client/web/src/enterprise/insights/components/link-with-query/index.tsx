import React from 'react'
import { LinkProps, useLocation } from 'react-router-dom'

import { RouterLink } from '@sourcegraph/wildcard'

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
        <RouterLink to={to + search} {...otherProps}>
            {children}
        </RouterLink>
    )
}
