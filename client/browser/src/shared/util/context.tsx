import { isFirefox } from '@sourcegraph/shared/src/util/browserDetection'

export const CLOUD_SOURCEGRAPH_URL = 'https://sourcegraph.com'

export function isCloudSourcegraphUrl(url: string): boolean {
    return url.replace(/\/$/, '') === CLOUD_SOURCEGRAPH_URL
}

/**
 * Returns the base URL where assets will be fetched from
 * (CSS, extension host worker, bundle...).
 *
 * The returned URL is guaranteed to have a trailing slash.
 *
 * If `window.SOURCEGRAPH_ASSETS_URL` is defined by a code host
 * self-hosting the integration bundle, it will be returned.
 * Otherwise, the given `sourcegraphURL` will be used.
 */
export function getAssetsURL(sourcegraphURL: string): string {
    const assetsURL = window.SOURCEGRAPH_ASSETS_URL || new URL('/.assets/extension/', sourcegraphURL).href
    return assetsURL.endsWith('/') ? assetsURL : assetsURL + '/'
}

export type PlatformName =
    | NonNullable<typeof globalThis.SOURCEGRAPH_INTEGRATION>
    | 'firefox-extension'
    | 'chrome-extension'
    | 'safari-extension'

export function getPlatformName(): PlatformName {
    if (window.SOURCEGRAPH_PHABRICATOR_EXTENSION) {
        return 'phabricator-integration'
    }
    if (window.SOURCEGRAPH_INTEGRATION) {
        return window.SOURCEGRAPH_INTEGRATION
    }
    if (isSafari()) {
        return 'safari-extension'
    }
    return isFirefox() ? 'firefox-extension' : 'chrome-extension'
}

export function getExtensionVersion(): string {
    if (globalThis.browser) {
        const manifest = browser.runtime.getManifest()
        return manifest.version
    }

    return 'NO_VERSION'
}

function isSafari(): boolean {
    // Chrome's user agent contains "Safari" as well as "Chrome", so for Safari
    // we must check that it does not include "Chrome"
    return window.navigator.userAgent.includes('Safari') && !window.navigator.userAgent.includes('Chrome')
}
