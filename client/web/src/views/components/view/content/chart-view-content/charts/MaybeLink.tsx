import React from 'react'

interface MaybeLinkProps extends React.AnchorHTMLAttributes<HTMLAnchorElement> {
    to?: string
}

/** Wraps the children in a link if to (link href) prop is passed. */
export const MaybeLink: React.FunctionComponent<MaybeLinkProps> = ({ children, to, ...props }) =>
    to ? (
        <a {...props} href={to}>
            {children}
        </a>
    ) : (
        (children as React.ReactElement)
    )
