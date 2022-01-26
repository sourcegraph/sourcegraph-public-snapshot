import isAbsoluteUrl from 'is-absolute-url'
import * as React from 'react'
import { Link } from 'react-router-dom'

import { AnchorLink, AnchorLinkProps } from '../AnchorLink'

/**
 * Uses react-router-dom's <Link> for relative URLs, <a> for absolute URLs. This is useful because passing an
 * absolute URL to <Link> will create an (almost certainly invalid) URL where the absolute URL is resolved to the
 * current URL, such as https://example.com/a/b/https://example.com/c/d.
 */
export const RouterLink: React.FunctionComponent<AnchorLinkProps> = React.forwardRef(
    ({ to, children, ...rest }: AnchorLinkProps, reference) => (
        <AnchorLink
            to={to}
            as={typeof to === 'string' && isAbsoluteUrl(to) ? undefined : Link}
            {...rest}
            ref={reference}
        >
            {children}
        </AnchorLink>
    )
)
