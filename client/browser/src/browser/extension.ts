const chrome = global.chrome

export const getURL = (path: string) =>
    chrome.extension.getURL(path)
