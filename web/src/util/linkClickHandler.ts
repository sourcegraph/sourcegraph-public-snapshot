import H from 'history'

/**
 * When used as a click handler on an element, captures clicks on relative links and navigates to their href using
 * the router instead of incurring a full page reload.
 *
 * This is useful on Markdown elements where all links are rendered with <a>, not react-router-dom's <Link>.
 */
export function createLinkClickHandler(history: H.History): React.MouseEventHandler<HTMLElement> {
    return (event: React.MouseEvent<HTMLElement>) => {
        // Capture clicks on relative links and use pushState for them instead of incurring a full
        // page reload.
        if (
            !event.defaultPrevented &&
            event.button === 0 &&
            !event.metaKey &&
            !event.altKey &&
            !event.ctrlKey &&
            !event.shiftKey
        ) {
            // Find nearest ancestor <a>.
            let e: HTMLElement | null = event.target as HTMLElement
            while (e) {
                const href = e.getAttribute('href')
                if (isAnchor(e) && !e.target && href && !/^(https?:)?\/\//.test(href)) {
                    event.preventDefault()
                    const url = new URL(e.href)
                    history.push({ pathname: url.pathname, hash: url.hash })

                    // HACK: Navigate to the in-page anchor. It does not work without this. This is definitely not
                    // the best solution. See RenderedFile for another solution (note that using RenderedFile in
                    // this component's render method instead of Markdown does not work either).
                    if (url.hash.startsWith('#')) {
                        setTimeout(() => (window.location.href = url.hash))
                    }

                    return
                }
                e = e.parentElement
            }
        }
    }
}

function isAnchor(e: HTMLElement): e is HTMLAnchorElement {
    return e.tagName === 'A'
}
