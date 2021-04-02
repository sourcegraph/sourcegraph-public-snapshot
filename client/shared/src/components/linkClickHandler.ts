import * as React from 'react'
import { anyOf, isInstanceOf } from '../util/types'
import * as H from 'history'
import { isExternalLink } from '../util/url'

/**
 * Returns a click handler that will make sure clicks on in-app links are handled on the client
 * and don't cause a full page reload.
 */
export const createLinkClickHandler = (history: H.History) => (event: React.MouseEvent<unknown>, target?: string): void => {
    // Do nothing if the link was requested to open in a new tab
    if (event.ctrlKey || event.metaKey) {
        return
    }

    // Try get href from target or obtain href value from event
    const href = typeof target === 'string'
        ? target
        : getHref(event)

    // In case if click happened within an anchor inside the markdown or target wasn't set
    if (!href) {
        return;
    }

    // Check if URL is outside the app
    if (isExternalLink(href)) {
        return
    }

    // Handle navigation programmatically
    event.preventDefault()
    const url = new URL(href)
    history.push(url.pathname + url.search + url.hash)
}

function getHref(event: React.MouseEvent<unknown>): string | undefined {
    // Check if click happened within an anchor inside the markdown
    const anchor = event.nativeEvent
        .composedPath()
        .slice(0, event.nativeEvent.composedPath().indexOf(event.currentTarget) + 1)
        .find(anyOf(isInstanceOf(HTMLAnchorElement), isInstanceOf(SVGAElement)))

    if (!anchor) {
        return
    }

    return typeof anchor.href === 'string'
        ? anchor.href
        : anchor.href.baseVal
}
