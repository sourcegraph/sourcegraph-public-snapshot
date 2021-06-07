/**
 * Strip provided URL parameters and update window history
 */
export function stripURLParameters(url: string, parametersToRemove: string[] = []): void {
    const parsedUrl = new URL(url)
    for (const key of parametersToRemove) {
        if (parsedUrl.searchParams.has(key)) {
            parsedUrl.searchParams.delete(key)
        }
    }
    window.history.replaceState(window.history.state, window.document.title, parsedUrl.href)
}

/**
 * Redact the pathname and search query from URLs to avoid
 * leaking sensitive information, while maintaining
 * non-sensitive query parameters used for attribution tracking.
 *
 * @param url the original, full URL
 */
export function redactSensitiveInfoFromURL(url: string): string {
    // Ensure we do not leak repo and file names in the URL
    const sourceURL = new URL(url)
    sourceURL.pathname = '/redacted'

    // Ensure we do not leak search queries in the URL
    const searchQuery = sourceURL.searchParams.get('q')
    if (searchQuery) {
        sourceURL.searchParams.set('q', 'redacted')
    }

    return sourceURL.href
}
