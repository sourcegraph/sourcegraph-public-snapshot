const chrome = global.chrome

export const setBadgeText = (details: chrome.browserAction.BadgeTextDetails) => {
    if (chrome && chrome.browserAction) {
        chrome.browserAction.setBadgeText(details)
    }
}

export const setPopup = (details: chrome.browserAction.PopupDetails) => {
    if (chrome && chrome.browserAction) {
        chrome.browserAction.setPopup(details)
    }
}

export function onClicked(listener: ((tab: chrome.tabs.Tab) => void)): void {
    if (chrome && chrome.browserAction && chrome.browserAction.onClicked) {
        chrome.browserAction.onClicked.addListener(listener)
    }
}
