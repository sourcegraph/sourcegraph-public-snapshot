import { isBackground } from '../context'

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
}

export const onMessage = (
    callback: (message: Message, sender: chrome.runtime.MessageSender, sendResponse?: (response: any) => void) => void
) => {
    if (chrome && chrome.runtime && chrome.runtime.onMessage) {
        chrome.runtime.onMessage.addListener(callback)
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

    return 'NO_VERSION'
}

export const getExtensionVersion = (): Promise<string> => {
    if (chrome && chrome.runtime && chrome.runtime.getManifest) {
        return Promise.resolve(getExtensionVersionSync())
    }

    return Promise.resolve('NO_VERSION')
}

export const onInstalled = (handler: (info?: chrome.runtime.InstalledDetails) => void) => {
    if (chrome && chrome.runtime && chrome.runtime.onInstalled) {
        chrome.runtime.onInstalled.addListener(handler)
    }
}
