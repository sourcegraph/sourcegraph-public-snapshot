const chrome = global.chrome
const safari = window.safari

export const getURL = (path: string) => {
    if (chrome && chrome.extension && chrome.extension.getURL) {
        return chrome.extension.getURL(path)
    }

    return safari.extension.baseURI + path
}
