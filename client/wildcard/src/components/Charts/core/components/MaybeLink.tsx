import React from 'react'

import { Link } from '../../../Link'

interface MaybeLinkProps extends React.AnchorHTMLAttributes<HTMLAnchorElement> {
    to?: string | void | null
}

/** Wraps the children in a link if to (link href) prop is passed. */
export const MaybeLink: React.FunctionComponent<React.PropsWithChildren<MaybeLinkProps>> = ({
    children,
    to,
    role,
    ...props
}) =>
    to ? (
        <Link {...props} to={to} role={role}>
            {children}
        </Link>
    ) : (
        <g role={role}>{children}</g>
    )
