import { commitIDFromPermalink } from '../../shared/util/dom'

/**
 * Get the commit ID from the permalink element on the page.
 */
export function getCommitIDFromPermalink(): string {
    return commitIDFromPermalink({
        selector: '.js-permalink-shortcut',
        hrefRegex: new RegExp('^/.*?/.*?/blob/([0-9a-f]{40})/'),
    })
}
