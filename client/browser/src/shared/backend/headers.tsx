import { getExtensionVersionSync, getPlatformName } from '../util/context'

/**
 * getHeaders emits the required headers for making requests to Sourcegraph server instances.
 * Requests can be blocked for various reasons and therefore the HTTP request MUST use the headers returned here.
 */
export function getHeaders(): { [name: string]: string } | undefined {
    // This is required (in most cases) for requests to be allowed by Sourcegraph's CORS rules. It
    // is not required for browser extension background worker requests, but it's harmless to
    // include it for those, too.
    return {
        'X-Requested-With': `Sourcegraph - ${getPlatformName()} v${getExtensionVersionSync()}`,
    }
}
