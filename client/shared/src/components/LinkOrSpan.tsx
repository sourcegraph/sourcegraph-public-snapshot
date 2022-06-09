import * as React from 'react'

import { LocationDescriptor } from 'history'

import { Link } from '@sourcegraph/wildcard'

/**
 * The LinkOrSpan component renders a <Link> if the "to" property is a non-empty string; otherwise it renders the
 * text in a <span> (with no link).
 */
export const LinkOrSpan: React.FunctionComponent<
    React.PropsWithChildren<
        {
            to: LocationDescriptor | undefined | null
            children?: React.ReactNode
        } & React.AnchorHTMLAttributes<HTMLAnchorElement>
    >
> = ({ to, className = '', children, ...otherProps }) => {
    if (to) {
        return (
            <Link to={to} className={className} {...otherProps}>
                {children}
            </Link>
        )
    }

    return (
        <span className={className} {...otherProps}>
            {children}
        </span>
    )
}
