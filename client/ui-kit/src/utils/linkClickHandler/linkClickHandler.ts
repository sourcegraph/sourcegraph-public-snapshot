import { MouseEventHandler } from 'react'
import { History } from 'history'
import { anyOf, isInstanceOf } from '@sourcegraph/utils/src/types'
import { isExternalLink } from '@sourcegraph/utils/src/url'

/**
 * Returns a click handler that will make sure clicks on in-app links are handled on the client
 * and don't cause a full page reload.
 */
export const createLinkClickHandler = (history: History): MouseEventHandler<unknown> => event => {
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
