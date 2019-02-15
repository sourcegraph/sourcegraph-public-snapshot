const chrome = global.chrome

export function contains(url: string): Promise<boolean> {
    return new Promise((resolve, reject) => {
        chrome.permissions.contains({ origins: [url + '/*'] }, resolve)
    })
}

export function request(urls: string[]): Promise<boolean> {
    return new Promise((resolve, reject) => {
        if (chrome && chrome.permissions) {
            urls = urls.map(url => url + '/')
            chrome.permissions.request(
                {
                    origins: [...urls],
                },
                resolve
            )
        }
    })
}

export function remove(url: string): Promise<boolean> {
    return new Promise((resolve, reject) => {
        if (chrome && chrome.permissions) {
            chrome.permissions.remove(
                {
                    origins: [url + '/*'],
                },
                resolve
            )
        }
    })
}

export function getAll(): Promise<chrome.permissions.Permissions> {
    return new Promise(resolve => {
        if (chrome && chrome.permissions) {
            chrome.permissions.getAll(resolve)
            return
        }
    })
}

export function onAdded(listener: (p: chrome.permissions.Permissions) => void): void {
    if (chrome && chrome.permissions && chrome.permissions.onAdded) {
        chrome.permissions.onAdded.addListener(listener)
    }
}

export function onRemoved(listener: (p: chrome.permissions.Permissions) => void): void {
    if (chrome && chrome.permissions && chrome.permissions.onRemoved) {
        chrome.permissions.onRemoved.addListener(listener)
    }
}
