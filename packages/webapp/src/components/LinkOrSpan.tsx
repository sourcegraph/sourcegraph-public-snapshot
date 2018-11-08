import * as React from 'react'
import { RouterLinkOrAnchor } from './RouterLinkOrAnchor'

/**
 * The LinkOrSpan component renders a <Link> (from react-router-dom) if the "to" property is a non-empty string;
 * otherwise it renders the text in a <span> (with no link).
 */
export const LinkOrSpan: React.SFC<
    {
        to: string | undefined | null
        children?: React.ReactNode
    } & React.AnchorHTMLAttributes<HTMLAnchorElement>
> = ({ to, className = '', children, ...otherProps }) => {
    const props = {
        ...otherProps,
        unselectable: otherProps.unselectable && otherProps.unselectable === 'on',
    }

    return to ? (
        <RouterLinkOrAnchor to={to} className={className} {...props}>
            {children}
        </RouterLinkOrAnchor>
    ) : (
        <span className={className} {...props}>
            {children}
        </span>
    )
}
