import { isBackground } from '../context'
import { getURL } from './extension'
import safariMessager from './safari/SafariMessager'

const safari = window.safari
const chrome = global.chrome

export interface Message {
    type:
        | 'setIdentity'
        | 'getIdentity'
        | 'setEnterpriseUrl'
        | 'setSourcegraphUrl'
        | 'removeEnterpriseUrl'
        | 'insertCSS'
        | 'setBadgeText'
        | 'openOptionsPage'
        | 'fetched-files'
        | 'repo-closed'
        | 'createBlobURL'
        | 'requestGraphQL'
    payload?: any
}

export const sendMessage = (message: Message, responseCallback?: (response: any) => void) => {
    if (chrome && chrome.runtime) {
        chrome.runtime.sendMessage(message, responseCallback)
    }

    if (safari) {
        safariMessager.send(message, responseCallback)
    }
}

export const onMessage = (
    callback: (message: Message, sender: chrome.runtime.MessageSender, sendResponse?: (response: any) => void) => void
) => {
    if (chrome && chrome.runtime && chrome.runtime.onMessage) {
        chrome.runtime.onMessage.addListener(callback)
        return
    }

    if (safari && safari.application) {
        safariMessager.onMessage(callback)
        return
    }

    throw new Error('do not call runtime.onMessage from a content script')
}

export const setUninstallURL = (url: string) => {
    if (chrome && chrome.runtime && chrome.runtime.setUninstallURL) {
        chrome.runtime.setUninstallURL(url)
    }
}

export const getManifest = () => {
    if (chrome && chrome.runtime && chrome.runtime.getManifest) {
        return chrome.runtime.getManifest()
    }

    return null
}

export const getContentScripts = () => {
    if (chrome && chrome.runtime) {
        return chrome.runtime.getManifest().content_scripts
    }
    return []
}

/**
 * openOptionsPage opens the options.html page. This must be called from the background script.
 * @param callback Called when the options page is opened.
 */
export const openOptionsPage = (callback?: () => void): void => {
    if (chrome && chrome.runtime) {
        if (!isBackground) {
            throw new Error('openOptionsPage must be called from the extension background script.')
        }
        if (chrome.runtime.openOptionsPage) {
            chrome.runtime.openOptionsPage(callback)
        }
    }
}

function getSafariExtensionVersion(): Promise<string> {
    return fetch(getURL('manifest.json'))
        .then(res => res.json())
        .then(({ version }: { version: string }) => version)
        .catch(err => {
            console.error('could not load manifest.json', err)

            return 'NO_VERSION'
        })
}

let safariVersion = 'NO_VERSION'

function initSafariVersion(): void {
    getSafariExtensionVersion()
        .then(version => {
            safariVersion = version
        })
        .catch(() => {
            // Don't care
        })
}

if (safari) {
    initSafariVersion()
}

// getExtensionVersionSync will be reliable on Chrome and Firefox but should only be used
// in safari when we know that `initSafariVersion` has had time to resolve. This is not
// sufficient for the options menu.
export const getExtensionVersionSync = (): string => {
    // Content scripts don't have access to the manifest, but do have chrome.app.getDetails
    const c = chrome as any
    if (c && c.app && c.app.getDetails) {
        const details = c.app.getDetails()
        if (details && details.version) {
            return details.version
        }
    }

    if (chrome && chrome.runtime && chrome.runtime.getManifest) {
        const manifest = chrome.runtime.getManifest()
        if (manifest) {
            return manifest.version
        }
    }

    if (safari) {
        return safariVersion
    }

    return 'NO_VERSION'
}

export const getExtensionVersion = (): Promise<string> => {
    if (chrome && chrome.runtime && chrome.runtime.getManifest) {
        return Promise.resolve(getExtensionVersionSync())
    }

    if (safari) {
        return getSafariExtensionVersion()
    }

    return Promise.resolve('NO_VERSION')
}

export const onInstalled = (handler: (info?: chrome.runtime.InstalledDetails) => void) => {
    if (chrome && chrome.runtime && chrome.runtime.onInstalled) {
        chrome.runtime.onInstalled.addListener(handler)
    }

    if (safari && !(safari.extension as SafariExtension).settings._wasInitialized) {
        handler()

        // Access settings directly here because we don't want to go through our
        // storage implementation to avoid onChange events and others from firing.
        //
        // This is ok here because we are only accessing this var from within this
        // function and we are inside a check to ensure safari exists.
        const settings = (safari.extension as SafariExtension).settings
        settings._wasInitialized = true
    }
}
