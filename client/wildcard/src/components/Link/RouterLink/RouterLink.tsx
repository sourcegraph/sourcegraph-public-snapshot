import * as React from 'react'

import isAbsoluteUrl from 'is-absolute-url'
// eslint-disable-next-line no-restricted-imports
import { Link } from 'react-router-dom'

import { ForwardReferenceComponent } from '../../../types'
import { AnchorLink, AnchorLinkProps } from '../AnchorLink'

/**
 * Uses react-router-dom's <Link> for relative URLs, <a> for absolute URLs. This is useful because passing an
 * absolute URL to <Link> will create an (almost certainly invalid) URL where the absolute URL is resolved to the
 * current URL, such as https://example.com/a/b/https://example.com/c/d.
 */
export const RouterLink = React.forwardRef(({ to, children, ...rest }, reference) => (
    <AnchorLink
        to={to}
        as={typeof to === 'string' && isAbsoluteUrl(to) ? undefined : (Link as Link<unknown>)}
        {...rest}
        ref={reference}
    >
        {children}
    </AnchorLink>
)) as ForwardReferenceComponent<Link, AnchorLinkProps>

RouterLink.displayName = 'RouterLink'
