const chrome = global.chrome

export const onUpdated = (
    callback: (tabId: number, changeInfo: chrome.tabs.TabChangeInfo, tab: chrome.tabs.Tab) => void
) => {
    if (chrome && chrome.tabs && chrome.tabs.onUpdated) {
        chrome.tabs.onUpdated.addListener(callback)
    }
}

export const getActive = (callback: (tab: chrome.tabs.Tab) => void) => {
    if (chrome && chrome.tabs) {
        chrome.tabs.query({ active: true }, (tabs: chrome.tabs.Tab[]) => {
            for (const tab of tabs) {
                callback(tab)
            }
        })
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

export const insertCSS = (tabId: number, details: InjectDetails, callback?: () => void) => {
    if (chrome && chrome.tabs && chrome.tabs.insertCSS) {
        chrome.tabs.insertCSS(tabId, details, callback)
    }
}

export const executeScript = (tabId: number, details: InjectDetails, callback?: (result: any[]) => void) => {
    if (chrome && chrome.tabs) {
        const { origin, whitelist, blacklist, ...rest } = details
        chrome.tabs.executeScript(tabId, rest, callback)
    }
}

export const sendMessage = (tabId: number, message: any, responseCallback?: (response: any) => void) => {
    if (chrome && chrome.tabs) {
        chrome.tabs.sendMessage(tabId, message, responseCallback)
    }
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
