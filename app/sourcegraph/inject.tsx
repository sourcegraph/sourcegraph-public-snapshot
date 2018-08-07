export function injectSourcegraphApp(marker: HTMLElement): void {
    if (document.getElementById(marker.id)) {
        return
    }

    window.addEventListener('load', () => {
        dispatchSourcegraphEvents(marker)
    })

    if (document.readyState === 'complete' || document.readyState === 'interactive') {
        dispatchSourcegraphEvents(marker)
    }
}

function dispatchSourcegraphEvents(marker: HTMLElement): void {
    // Generate and insert DOM element, in case this code executes first.
    document.body.appendChild(marker)
    // Send custom webapp <-> extension registration event in case webapp listener is attached first.
    document.dispatchEvent(new CustomEvent<{}>('sourcegraph:browser-extension-registration'))
}
