import storage from '../storage'

export const URLError = {
    Empty: 1,
    Invalid: 2,
    HTTPNotSupported: 3,
}

export function upsertSourcegraphUrl(input: string, onDone?: (urls: string[]) => void): null | number {
    try {
        const url = new URL(input)
        if (!url || !url.origin || url.origin === 'null') {
            return URLError.Empty
        }

        if (window.safari && url.protocol === 'http:') {
            return URLError.HTTPNotSupported
        }

        storage.getSync(items => {
            let serverUrls = items.serverUrls || []
            serverUrls = [...serverUrls, url.origin]

            storage.setSync(
                {
                    sourcegraphURL: url.origin,
                    serverUrls: [...new Set(serverUrls)],
                },
                onDone ? () => onDone(serverUrls) : undefined
            )
        })
    } catch {
        return URLError.Invalid
    }

    return null
}
