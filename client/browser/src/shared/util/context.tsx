import * as runtime from '../../browser/runtime'
import storage from '../../browser/storage'
import { isPhabricator } from '../../context'
import { EventLogger } from '../tracking/EventLogger'

export const DEFAULT_SOURCEGRAPH_URL = 'https://sourcegraph.com'

export let eventLogger = new EventLogger()

export let sourcegraphUrl =
    window.localStorage.getItem('SOURCEGRAPH_URL') || window.SOURCEGRAPH_URL || DEFAULT_SOURCEGRAPH_URL

export let renderMermaidGraphsEnabled = false

export let inlineSymbolSearchEnabled = false

interface UrlCache {
    [key: string]: string
}

export const repoUrlCache: UrlCache = {}

if (window.SG_ENV === 'EXTENSION') {
    storage.getSync(items => {
        sourcegraphUrl = items.sourcegraphURL
        renderMermaidGraphsEnabled = items.featureFlags.renderMermaidGraphsEnabled
        inlineSymbolSearchEnabled = items.featureFlags.inlineSymbolSearchEnabled
    })
}

export function setSourcegraphUrl(url: string): void {
    sourcegraphUrl = url
}

export function isBrowserExtension(): boolean {
    return window.SOURCEGRAPH_PHABRICATOR_EXTENSION || false
}

export function isSourcegraphDotCom(url: string = sourcegraphUrl): boolean {
    return url === DEFAULT_SOURCEGRAPH_URL
}

export function setRenderMermaidGraphsEnabled(enabled: boolean): void {
    renderMermaidGraphsEnabled = enabled
}

export function setInlineSymbolSearchEnabled(enabled: boolean): void {
    inlineSymbolSearchEnabled = enabled
}

export function getPlatformName():
    | 'phabricator-integration'
    | 'safari-extension'
    | 'firefox-extension'
    | 'chrome-extension' {
    if (window.SOURCEGRAPH_PHABRICATOR_EXTENSION) {
        return 'phabricator-integration'
    }

    if (typeof window.safari !== 'undefined') {
        return 'safari-extension'
    }

    return isFirefoxExtension() ? 'firefox-extension' : 'chrome-extension'
}

export function getExtensionVersionSync(): string {
    return runtime.getExtensionVersionSync()
}

export function isFirefoxExtension(): boolean {
    return window.navigator.userAgent.indexOf('Firefox') !== -1
}

/**
 * Check the DOM to see if we can determine if a repository is private or public.
 */
export function isPrivateRepository(): boolean {
    if (isPhabricator) {
        return true
    }
    const header = document.querySelector('.repohead-details-container')
    if (!header) {
        return false
    }
    return !!header.querySelector('.private')
}

export function canFetchForURL(url: string): boolean {
    if (url === DEFAULT_SOURCEGRAPH_URL && isPrivateRepository()) {
        return false
    }
    return true
}
