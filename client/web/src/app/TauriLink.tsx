// Since we're using forwardRef for everything in this file, we need
// to forward all the props to the underlying component.
/* eslint-disable no-restricted-syntax */

import React, { useMemo } from 'react'

import isAbsoluteUrl from 'is-absolute-url'

import { addSourcegraphAppOutboundUrlParameters } from '@sourcegraph/shared/src/util/url'
import { AnchorLink, type Link, RouterLink } from '@sourcegraph/wildcard'

/**
 * A link that opens in the browser if the URL is absolute, otherwise uses the router.
 * If the URL is a help page, it will open in the browser with the appropriate outbound URL parameters.
 *
 * With the `shell-open` feature enabled in the Tauri app, `target="_blank"` links will automatically
 * open in the user's default browser; there is no need to call the Tauri API directly.
 */
export const TauriLink = React.forwardRef(({ to, children, ...rest }, reference) => {
    if (to && isAbsoluteUrl(to)) {
        return (
            <AnchorLink {...rest} to={to} ref={reference} target="_blank">
                {children}
            </AnchorLink>
        )
    }

    if (to?.startsWith('/help/') || to === '/help') {
        return (
            <TauriHelpLink {...rest} to={to} ref={reference}>
                {children}
            </TauriHelpLink>
        )
    }

    return (
        <RouterLink {...rest} to={to} ref={reference}>
            {children}
        </RouterLink>
    )
}) as Link

const TauriHelpLink = React.forwardRef(function TauriHelpLink({ to, children, ...rest }, reference) {
    const toWithParams = useMemo(() => {
        const absoluteTo = to.replace(/^\/help/, 'https://docs.sourcegraph.com')
        return addSourcegraphAppOutboundUrlParameters(absoluteTo)
    }, [to])

    return (
        <AnchorLink {...rest} to={toWithParams} ref={reference} target="_blank">
            {children}
        </AnchorLink>
    )
}) as Link

TauriLink.displayName = 'TauriLink'
