import React, { MouseEvent, useCallback } from 'react'

import { open } from '@tauri-apps/api/shell'
import isAbsoluteUrl from 'is-absolute-url'

import { logger } from '@sourcegraph/common'
import { AnchorLink, Link, RouterLink } from '@sourcegraph/wildcard'

// A link component that uses the Tauri shell API to open external links in the user's default browser.
export const TauriLink = React.forwardRef(({ to, children, onClick, ...rest }, reference) => {
    const onClickShellOpen = useCallback(
        (event: MouseEvent<HTMLAnchorElement>) => {
            onClick?.(event)
            event.preventDefault()
            open(to).catch(error => logger.error('Failed to open link', error))
        },
        [onClick, to]
    )

    if (to && isAbsoluteUrl(to)) {
        return (
            // eslint-disable-next-line no-restricted-syntax
            <AnchorLink to={to} ref={reference} onClick={onClickShellOpen} {...rest}>
                {children}
            </AnchorLink>
        )
    }

    if (to?.startsWith('/help/') || to === '/help') {
        const absoluteTo = to.replace(/^\/help/, 'https://docs.sourcegraph.com')
        // eslint-disable-next-line no-restricted-syntax
        return (
            <TauriLink to={absoluteTo} ref={reference} {...rest}>
                {children}
            </TauriLink>
        )
    }

    return (
        // eslint-disable-next-line no-restricted-syntax
        <RouterLink to={to} ref={reference} onClick={onClick} {...rest}>
            {children}
        </RouterLink>
    )
}) as Link

TauriLink.displayName = 'TauriLink'
