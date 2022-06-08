import React from 'react'

import isAbsoluteUrl from 'is-absolute-url'
// eslint-disable-next-line no-restricted-imports
import { Link as ReactRouterLink } from 'react-router-dom'

import { ForwardReferenceComponent } from '@sourcegraph/wildcard'

import { AnchorLink } from '../AnchorLink'
import type { LinkProps } from '../Link'

/**
 * Uses react-router-dom's Link for relative URLs, AnchorLink for absolute URLs.
 * This is useful because passing an absolute URL to Link will create
 * an (almost certainly invalid) URL where the absolute URL is resolved to the
 * current URL, such as `https://example.com/a/b/https://example.com/c/d`.
 *
 * @param as Only supported when `to` is an absolute URL
 */
export const RouterLink = React.forwardRef(({ as: Component, to, children, ...rest }, reference) => {
    if ((to && isAbsoluteUrl(to)) || !Component) {
        return (
            <AnchorLink to={to} ref={reference} {...rest}>
                {children}
            </AnchorLink>
        )
    }

    return (
        <ReactRouterLink to={to} component={Component} ref={reference} {...rest}>
            {children}
        </ReactRouterLink>
    )
}) as ForwardReferenceComponent<ReactRouterLink, LinkProps>

RouterLink.displayName = 'RouterLink'
