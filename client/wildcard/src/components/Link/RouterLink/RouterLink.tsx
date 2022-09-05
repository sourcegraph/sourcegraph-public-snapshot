import React from 'react'

import isAbsoluteUrl from 'is-absolute-url'
// eslint-disable-next-line no-restricted-imports
import { Link as ReactRouterLink } from 'react-router-dom'

import { AnchorLink } from '../AnchorLink'
import type { Link } from '../Link'

import anchorLinkStyles from '../AnchorLink/AnchorLink.module.scss'

/**
 * Uses react-router-dom's Link for relative URLs, AnchorLink for absolute URLs.
 * This is useful because passing an absolute URL to Link will create
 * an (almost certainly invalid) URL where the absolute URL is resolved to the
 * current URL, such as `https://example.com/a/b/https://example.com/c/d`.
 */
export const RouterLink = React.forwardRef(({ to, children, ...rest }, reference) => {
    if (to && isAbsoluteUrl(to)) {
        return (
            <AnchorLink to={to} ref={reference} {...rest}>
                {children}
            </AnchorLink>
        )
    }

    return (
        <ReactRouterLink className={anchorLinkStyles.anchorLink} to={to} ref={reference} {...rest}>
            {children}
        </ReactRouterLink>
    )
}) as Link

RouterLink.displayName = 'RouterLink'
