const chrome = global.chrome

export const setBadgeText = (details: chrome.browserAction.BadgeTextDetails) => {
    if (chrome && chrome.browserAction) {
        chrome.browserAction.setBadgeText(details)
    }
}

export const setPopup = (details: chrome.browserAction.PopupDetails): Promise<void> =>
    new Promise<void>(resolve => {
        if (chrome && chrome.browserAction) {
            chrome.browserAction.setPopup(details, resolve)
            return
        }
        resolve()
    })

export function onClicked(listener: (tab: chrome.tabs.Tab) => void): void {
    if (chrome && chrome.browserAction && chrome.browserAction.onClicked) {
        chrome.browserAction.onClicked.addListener(listener)
    }
}
