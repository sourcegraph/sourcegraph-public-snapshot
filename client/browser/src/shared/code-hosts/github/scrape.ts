import { commitIDFromPermalink } from '../../util/dom'

import { getPermalinkSelector } from './util'

/**
 * Get the commit ID from the permalink element on the page.
 */
export function getCommitIDFromPermalink(): string {
    return '1287d3965940425a60b780020a65a0d760909592' // TODO: replace hardcoded value
    return commitIDFromPermalink({
        // TODO: this selector doesn't fit as it don't have the full commit SHA all the time
        // Use payload.refInfo.currentOid from script[data-target='react-app.embeddedData'] (the only place on the page with the full commit SHA)
        selector: getPermalinkSelector()!,
        hrefRegex: /^\/.*?\/.*?\/(?:blob|tree)\/([\da-f]{40})/,
    })
}
