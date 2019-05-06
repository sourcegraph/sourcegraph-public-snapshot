import { Observable, of } from 'rxjs'
import { observeStorageKey, storage } from '../../browser/storage'

export const DEFAULT_SOURCEGRAPH_URL = 'https://sourcegraph.com'

export let sourcegraphUrl =
    window.localStorage.getItem('SOURCEGRAPH_URL') || window.SOURCEGRAPH_URL || DEFAULT_SOURCEGRAPH_URL

if (window.SG_ENV === 'EXTENSION' && globalThis.browser) {
    // tslint:disable-next-line: no-floating-promises TODO just get rid of the global sourcegraphUrl
    storage.sync.get().then(items => {
        if (items.sourcegraphURL) {
            sourcegraphUrl = items.sourcegraphURL
        }
    })
}

export function observeSourcegraphURL(): Observable<string | undefined> {
    if (window.SG_ENV === 'EXTENSION') {
        return observeStorageKey('sync', 'sourcegraphURL')
    }
    return of(sourcegraphUrl)
}

export function setSourcegraphUrl(url: string): void {
    sourcegraphUrl = url
}

export function isSourcegraphDotCom(url: string = sourcegraphUrl): boolean {
    return url === DEFAULT_SOURCEGRAPH_URL
}

export function getPlatformName(): 'phabricator-integration' | 'firefox-extension' | 'chrome-extension' {
    if (window.SOURCEGRAPH_PHABRICATOR_EXTENSION) {
        return 'phabricator-integration'
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
    return window.navigator.userAgent.indexOf('Firefox') !== -1
}
