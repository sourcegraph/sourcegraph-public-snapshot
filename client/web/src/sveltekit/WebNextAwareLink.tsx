import { forwardRef } from 'react'

import isAbsoluteUrl from 'is-absolute-url'

import { RouterLink, type Link, AnchorLink } from '@sourcegraph/wildcard'

import { isEnabledRoute } from './util'

/**
 * This link causes a full page reload to load the SvelteKit app from the server if
 * the web-next or web-next-rollout feature flags are enabled, and the link is to a
 * supported route.
 * Otherwise it falls back to {@link RouterLink}.
 */
export const WebNextAwareLink = forwardRef(({ to, children, ...rest }, reference) => {
    if (to && !isAbsoluteUrl(to)) {
        const url = new URL(to, window.location.href)
        if (isEnabledRoute(url.pathname)) {
            // Render an AnchorLink to bypass React Router and force
            // a full page reload to fetch the SvelteKit app from the server
            return (
                <AnchorLink to={to} ref={reference} {...rest}>
                    {children}
                </AnchorLink>
            )
        }
    }

    return (
        <RouterLink to={to} {...rest} ref={reference}>
            {children}
        </RouterLink>
    )
}) as Link
WebNextAwareLink.displayName = 'WebNextAwareLink'
