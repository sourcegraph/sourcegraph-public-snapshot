import { formatISO, startOfWeek } from 'date-fns'

export const DOTCOM_URL = new URL('https://sourcegraph.com')

/**
 * Strip provided URL parameters and update window history
 */
export function stripURLParameters(url: string, parametersToRemove: string[] = []): void {
    const parsedUrl = new URL(url)
    const existingParameters = parametersToRemove.filter(key => parsedUrl.searchParams.has(key))

    // Update history state only if we have parameters to remove in the url.
    if (existingParameters.length !== 0) {
        for (const key of existingParameters) {
            parsedUrl.searchParams.delete(key)
        }

        window.history.replaceState(window.history.state, window.document.title, parsedUrl.href)
    }
}

/**
 * Redact the pathname and search query from sourcegraph.com URLs to avoid
 * leaking sensitive information from Sourcegraph Cloud, while maintaining
 * non-sensitive query parameters used for attribution tracking.
 *
 * Note that URL redaction also happens in internal/usagestats/event_handlers.go.
 *
 * @param url the original, full URL
 */
export function redactSensitiveInfoFromAppURL(url: string): string {
    const sourceURL = new URL(url)

    // if hostname is about.sourcegraph or docs.sourcegraph, etc do not redact
    const approved_hosts = [
        'sourcegraph.com',
        'about.sourcegraph.com',
        'docs.sourcegraph.com',
        'info.sourcegraph.com',
        'signup.sourcegraph.com',
    ]

    if (approved_hosts.includes(sourceURL.hostname)) {
        return url
    }

    // Redact all GitHub.com code URLs, GitLab.com code URLs, and search URLs to ensure we do not leak sensitive information.
    if (sourceURL.pathname.startsWith('/github.com')) {
        sourceURL.pathname = '/github.com/redacted'
    } else if (sourceURL.pathname.startsWith('/gitlab.com')) {
        sourceURL.pathname = '/gitlab.com/redacted'
    } else if (sourceURL.pathname.startsWith('/search')) {
        sourceURL.pathname = '/search/redacted'
    } else if (sourceURL.pathname.startsWith('/sign-in')) {
        sourceURL.pathname = '/sign-in/redacted'
    } else {
        sourceURL.pathname = '/redacted'
        return sourceURL.href
    }

    const marketingQueryParameters = new Set([
        'utm_source',
        'utm_campaign',
        'utm_medium',
        'utm_term',
        'utm_content',
        'utm_cid',
    ])
    // Ensure we do not leak search queries or other sensitive info in the URL
    // by only maintaining UTM parameters for attribution.
    for (const [parameter] of sourceURL.searchParams) {
        if (!marketingQueryParameters.has(parameter)) {
            sourceURL.searchParams.set(parameter, 'redacted')
        }
    }

    return sourceURL.href
}

/**
 * Returns the Monday at or before the supplied date, in YYYY-MM-DD format.
 * This is used to generate cohort IDs for users who
 * started using the site on the same week.
 */
export function getPreviousMonday(date: Date): string {
    return formatISO(startOfWeek(date, { weekStartsOn: 1 }), { representation: 'date' })
}
