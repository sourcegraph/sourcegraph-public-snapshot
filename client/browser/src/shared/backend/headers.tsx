import { isInPage, isPhabricator } from '../../context'
import { getExtensionVersionSync, getPlatformName } from '../util/context'

/**
 * getHeaders emits the required headers for making requests to Sourcegraph server instances.
 * Requests can be blocked for various reasons and therefore the HTTP request MUST use the headers returned here.
 */
export function getHeaders(): { [name: string]: string } | undefined {
    if (isPhabricator && isInPage) {
        return {
            'X-Requested-With': `Sourcegraph - ${getPlatformName()} v${getExtensionVersionSync()}`,
        }
    }

    const headers: { [name: string]: string } = {}

    // The HTTP request MUST contain at least one of the following for the Sourcegraph server to accept it
    // (according to its CORS rules).
    //
    // - An Origin header with a URI that Sourcegraph trusts. The prod and dev Chrome extension origins
    //   (chrome-extension://...) are trusted, and site admins can specify other trusted origins in the site config
    //   "corsOrigin" property.
    // - An X-Requested-With header (with any nonempty value). This tells the server that the request is from a
    //   trusted origin, because that header could not be added to a cross-domain request unless it already passed
    //   the server's CORS rules. (See
    //   https://stackoverflow.com/questions/17478731/whats-the-point-of-the-x-requested-with-header for more
    //   info.)
    // - Using application/json as the Content-Type or Accept result in CORS blocking the request.
    //
    // The browsers all handle this situation differently.
    //
    // - Chrome (usually) automatically includes "Origin: chrome-extension://..." but does NOT allow us to include
    //   other headers (such as X-Requested-With), or else they are blocked by the CSP policy of GitHub (or other
    //   target page). Chrome sometimes sends "Origin: <window.location.origin>" instead; it's not clear in what
    //   cases or why this occurs.
    // - Safari includes "Origin: <window.location.origin>" (where "<window.location.origin>" is the
    //   `window.location.origin` of the current page).
    // - Firefox does NOT include any Origin header, so we need to send an "X-Requested-With" header.
    const needsCORSHeader = getPlatformName() === 'firefox-extension' || isPhabricator
    if (needsCORSHeader) {
        headers['X-Requested-With'] = `Sourcegraph - ${getPlatformName()} v${getExtensionVersionSync()}`
    }

    return headers
}
