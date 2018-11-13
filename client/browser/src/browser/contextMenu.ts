const chrome = global.chrome

export const create = (details: chrome.contextMenus.CreateProperties, callback?: () => void) => {
    if (chrome && chrome.contextMenus) {
        chrome.contextMenus.create(details, callback)
    }
}
