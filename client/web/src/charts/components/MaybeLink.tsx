import React from 'react'

import { Link } from '@sourcegraph/wildcard'

interface MaybeLinkProps extends React.AnchorHTMLAttributes<HTMLAnchorElement> {
    to?: string | void
}

/** Wraps the children in a link if to (link href) prop is passed. */
export const MaybeLink: React.FunctionComponent<React.PropsWithChildren<MaybeLinkProps>> = ({
    children,
    to,
    ...props
}) =>
    to ? (
        <Link {...props} to={to}>
            {children}
        </Link>
    ) : (
        (children as React.ReactElement)
    )
