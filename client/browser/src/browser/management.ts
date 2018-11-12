const chrome = global.chrome

export const getSelf = (responseCallback?: (response?: chrome.management.ExtensionInfo) => void) => {
    if (chrome && chrome.management && chrome.management.getSelf) {
        chrome.management.getSelf(responseCallback)
        return
    }
    if (responseCallback) {
        responseCallback()
    }
}
