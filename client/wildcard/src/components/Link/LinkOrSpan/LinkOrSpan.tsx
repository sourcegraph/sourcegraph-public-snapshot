import * as React from 'react'

import type { ForwardReferenceComponent } from '../../../types'
import { Link } from '../Link/Link'

type LinkOrSpanProps = React.PropsWithChildren<
    {
        to: string | undefined | null
        children?: React.ReactNode
    } & React.AnchorHTMLAttributes<HTMLAnchorElement>
>

/**
 * The LinkOrSpan component renders a <Link> if the "to" property is a non-empty string; otherwise it renders the
 * text in a <span> (with no link).
 */
const LinkOrSpan = React.forwardRef(({ to, className = '', children, ...otherProps }: LinkOrSpanProps, reference) => {
    if (to) {
        return (
            <Link ref={reference} to={to} className={className} {...otherProps}>
                {children}
            </Link>
        )
    }

    return (
        <span ref={reference} className={className} {...otherProps}>
            {children}
        </span>
    )
}) as ForwardReferenceComponent<typeof Link, LinkOrSpanProps>

LinkOrSpan.displayName = 'LinkOrSpan'

export { LinkOrSpan }
