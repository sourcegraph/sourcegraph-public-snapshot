import { commitIDFromPermalink } from '../../util/dom'

/**
 * Get the commit ID from either permalink element on the page (old UI) or script with embedded data (new UI).
 */
export function getCommitIDFromPermalink(): string {
    // new UI
    const embeddedDataScriptSelector = 'script[data-target="react-app.embeddedData"]'
    const script = document.querySelector<HTMLScriptElement>(embeddedDataScriptSelector)
    if (script) {
        try {
            const data = JSON.parse(script.textContent || '')
            return data.payload.refInfo.currentOid
        } catch {
            throw new Error(`Could not parse '${embeddedDataScriptSelector}' content or extract commit ID from it`)
        }
    }

    // old UI
    return commitIDFromPermalink({
        selector: '.js-permalink-shortcut',
        hrefRegex: /^\/.*?\/.*?\/(?:blob|tree)\/([\da-f]{40})/,
    })
}
