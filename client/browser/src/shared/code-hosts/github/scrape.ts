import { commitIDFromPermalink } from '../../util/dom'

/**
 * Get the commit ID from the permalink element on the page.
 */
export function getCommitIDFromPermalink(): string {
    // new UI
    const script = document.querySelector<HTMLScriptElement>('script[data-target="react-app.embeddedData"]')
    if (script) {
        const data = JSON.parse(script.textContent || '')
        return data.payload.refInfo.currentOid
    }

    // old UI
    return commitIDFromPermalink({
        selector: '.js-permalink-shortcut',
        hrefRegex: /^\/.*?\/.*?\/(?:blob|tree)\/([\da-f]{40})/,
    })
}
