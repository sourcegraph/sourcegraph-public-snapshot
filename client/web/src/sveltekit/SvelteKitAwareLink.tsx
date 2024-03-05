import React from 'react'

import isAbsoluteUrl from 'is-absolute-url'
import { useFeatureFlag } from 'src/featureFlags/useFeatureFlag'

import { RouterLink, type Link, AnchorLink } from '@sourcegraph/wildcard'

import { isRolledOutRoute, isSupportedRoute } from './util'

export const SvelteKitAwareLink = React.forwardRef(({ to, children, ...rest }, reference) => {
    const [webNextEnabled] = useFeatureFlag('web-next-rollout', false)
    const [webNext] = useFeatureFlag('web-next', false)

    if (to && !isAbsoluteUrl(to)) {
        const url = new URL(to, window.location.href)
        if ((webNextEnabled && isRolledOutRoute(url.pathname)) || (webNext && isSupportedRoute(url.pathname))) {
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
