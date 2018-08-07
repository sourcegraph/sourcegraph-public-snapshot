const safari = window.safari
const chrome = global.chrome

export const onUpdated = (
    callback: (tabId: number, changeInfo: chrome.tabs.TabChangeInfo, tab: chrome.tabs.Tab) => void
) => {
    if (chrome && chrome.tabs && chrome.tabs.onUpdated) {
        chrome.tabs.onUpdated.addListener(callback)
    }

    if (safari && safari.application) {
        safari.application.addEventListener('change', (...args: any[]) => {
            console.log('application on change', ...args)
        })
    }
}

export const getActive = (callback: (tab: chrome.tabs.Tab) => void) => {
    if (chrome && chrome.tabs) {
        chrome.tabs.query({ active: true }, (tabs: chrome.tabs.Tab[]) => {
            for (const tab of tabs) {
                callback(tab)
            }
        })
    } else {
        console.log('SAFARI tabs.getActive')
    }
}

export const query = (queryInfo: chrome.tabs.QueryInfo, handler: (tabs: chrome.tabs.Tab[]) => void) => {
    if (chrome && chrome.tabs && chrome.tabs.onUpdated) {
        chrome.tabs.query(queryInfo, handler)
    }
}

export const reload = (tabId: number) => {
    if (chrome && chrome.tabs && chrome.tabs.reload) {
        chrome.tabs.reload(tabId)
    }
}

interface InjectDetails extends chrome.tabs.InjectDetails {
    origin?: string
    whitelist?: string[]
    blacklist?: string[]
    file?: string
    runAt?: string
}

const patternify = (str: string) => `${str}/*`

export const insertCSS = (tabId: number, details: InjectDetails, callback?: () => void) => {
    if (chrome && chrome.tabs && chrome.tabs.insertCSS) {
        chrome.tabs.insertCSS(tabId, details, callback)
    }

    // Safari doesn't have a target tab filter for injecting CSS.
    // Our workaround is to pass in the current tab's origin as the whitelist.
    if (safari) {
        const extension = safari.extension as SafariExtension

        const whitelist = (details.whitelist || []).map(patternify)
        const blacklist = (details.blacklist || []).map(patternify)

        extension.addContentStyleSheetFromURL(safari.extension.baseURI + details.file, whitelist, blacklist)
    }
}

export const executeScript = (tabId: number, details: InjectDetails, callback?: (result: any[]) => void) => {
    if (chrome && chrome.tabs) {
        const { origin, whitelist, blacklist, ...rest } = details
        chrome.tabs.executeScript(tabId, rest, callback)
    }

    // Safari doesn't have a target tab filter for executing js.
    // Our workaround is to pass in the current tab's origin as the whitelist.
    if (safari) {
        const extension = safari.extension as SafariExtension
        extension.addContentScriptFromURL(safari.extension.baseURI + details.file, [details.origin!], [], true)
    }
}

export const sendMessage = (tabId: number, message: any, responseCallback?: (response: any) => void) => {
    if (chrome && chrome.tabs) {
        chrome.tabs.sendMessage(tabId, message, responseCallback)
    }

    // Noop on safari
    // TODO: Do we actually need this? It's currently only being used to check if the active tab is a sourcegraph server.
}

export const create = (props: chrome.tabs.CreateProperties, callback?: (tab: chrome.tabs.Tab) => void) => {
    if (chrome && chrome.tabs) {
        chrome.tabs.create(props, callback)
    }
}

export const update = (props: chrome.tabs.UpdateProperties, callback?: (tab?: chrome.tabs.Tab) => void) => {
    if (chrome && chrome.tabs) {
        chrome.tabs.update(props, callback)
    }
}
