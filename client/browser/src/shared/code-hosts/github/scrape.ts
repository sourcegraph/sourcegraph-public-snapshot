import { commitIDFromPermalink } from '../../util/dom'

import { getPermalinkSelector } from './util'

/**
 * Get the commit ID from the permalink element on the page.
 */
export function getCommitIDFromPermalink(): string {
    return commitIDFromPermalink({
        selector: getPermalinkSelector()!,
        hrefRegex: /^\/.*?\/.*?\/(?:blob|tree)\/([\da-f]{40})/,
    })
}
