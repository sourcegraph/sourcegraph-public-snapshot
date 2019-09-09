import { Observable, of } from 'rxjs'
import { map } from 'rxjs/operators'
import { observeStorageKey } from '../../browser/storage'

export const DEFAULT_SOURCEGRAPH_URL = 'https://sourcegraph.com'
export const DEFAULT_ASSETS_URL: string = new URL('/.assets/extension/', DEFAULT_SOURCEGRAPH_URL).href

export function observeSourcegraphURL(isExtension: boolean): Observable<string> {
    if (isExtension) {
        return observeStorageKey('sync', 'sourcegraphURL').pipe(
            map(sourcegraphURL => sourcegraphURL || DEFAULT_SOURCEGRAPH_URL)
        )
    }
    return of(window.SOURCEGRAPH_URL || window.localStorage.getItem('SOURCEGRAPH_URL') || DEFAULT_SOURCEGRAPH_URL)
}

export function isSourcegraphDotCom(url: string): boolean {
    return url === DEFAULT_SOURCEGRAPH_URL
}

type PlatformName = 'phabricator-integration' | 'bitbucket-integration' | 'firefox-extension' | 'chrome-extension'

export function getPlatformName(): PlatformName {
    if (window.SOURCEGRAPH_PHABRICATOR_EXTENSION) {
        return 'phabricator-integration'
    }
    if (window.SOURCEGRAPH_INTEGRATION) {
        return window.SOURCEGRAPH_INTEGRATION
    }
    return isFirefoxExtension() ? 'firefox-extension' : 'chrome-extension'
}

export function getExtensionVersion(): string {
    if (globalThis.browser) {
        const manifest = browser.runtime.getManifest()
        return manifest.version
    }

    return 'NO_VERSION'
}

function isFirefoxExtension(): boolean {
    return window.navigator.userAgent.includes('Firefox')
}
