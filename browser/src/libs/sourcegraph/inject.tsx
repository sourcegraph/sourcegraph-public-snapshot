export const EXTENSION_MARKER_ID = 'sourcegraph-app-background'

/**
 * Injects a `#sourcegraph-app-background` hidden element.
 *
 * This element is checked for in the webapp to know if the browser extension
 * is installed, and in the browser extension to determine whether a native integration
 * is already running on the page.
 *
 * Not idempotent.
 */
export function injectExtensionMarker(): void {
    const extensionMarker = document.createElement('div')
    extensionMarker.id = EXTENSION_MARKER_ID
    extensionMarker.style.display = 'none'
    document.body.appendChild(extensionMarker)
}

/**
 * Injects the extension marker and dispatches a custom event
 * to signal to Sourcegraph web app that the browser extension is installed.
 *
 *  Not idempotent.
 */
export function signalBrowserExtensionInstalled(): void {
    injectExtensionMarker()

    window.addEventListener(
        'load',
        () => {
            dispatchSourcegraphEvents()
        },
        { once: true }
    )

    if (document.readyState === 'complete' || document.readyState === 'interactive') {
        dispatchSourcegraphEvents()
    }
}

function dispatchSourcegraphEvents(): void {
    // Send custom webapp <-> extension registration event in case webapp listener is attached first.
    document.dispatchEvent(new CustomEvent<{}>('sourcegraph:browser-extension-registration'))
}

export const checkIsSourcegraph = (sourcegraphServerUrl: string): boolean =>
    window.location.origin === sourcegraphServerUrl ||
    /^https?:\/\/(www.)?sourcegraph.com/.test(location.href) ||
    !!document.getElementById('sourcegraph-chrome-webstore-item')
