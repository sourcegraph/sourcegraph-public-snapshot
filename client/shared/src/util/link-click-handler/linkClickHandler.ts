import * as H from 'history'
import * as React from 'react'

import { anyOf, isInstanceOf } from '../types'
import { isExternalLink } from '../url'

/**
 * Returns a click handler for link element that will make sure clicks on in-app links are handled on the client
 * and don't cause a full page reload.
 */
export const createLinkClickHandler = (history: H.History): React.MouseEventHandler<unknown> => event => {
    // Do nothing if the link was requested to open in a new tab
    if (event.ctrlKey || event.metaKey) {
        return
    }

    // Check if click happened within an anchor inside the markdown
    const anchor = event.nativeEvent
        .composedPath()
        .slice(0, event.nativeEvent.composedPath().indexOf(event.currentTarget) + 1)
        .find(anyOf(isInstanceOf(HTMLAnchorElement), isInstanceOf(SVGAElement)))
    if (!anchor) {
        return
    }
    const href = typeof anchor.href === 'string' ? anchor.href : anchor.href.baseVal

    // Check if URL is outside the app
    if (isExternalLink(href)) {
        return
    }

    // Handle navigation programmatically
    event.preventDefault()
    const url = new URL(href)
    history.push(url.pathname + url.search + url.hash)
}

/**
 * Returns a click handler for any element that takes event and target URL
 * that will redirect to this URL and in case if cmd+click happened that will open a new browser tab
 */
export const createProgrammaticLinkHandler = (history: H.History) => (
    event: React.MouseEvent<unknown>,
    target: string
): void => {
    const url = new URL(target)

    // Do nothing if the link was requested to open in a new tab
    if (event.ctrlKey || event.metaKey) {
        window.open(url.origin + url.pathname + url.search + url.hash, '_blank')?.focus()

        return
    }

    // Handle navigation programmatically
    history.push(url.pathname + url.search + url.hash)
}
