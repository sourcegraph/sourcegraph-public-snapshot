interface FetchOptions {
    headers: Headers
}

export function combineHeaders(a: Headers, b: Headers): Headers {
    const headers = new Headers(a)
    for (const [name, val] of b) {
        headers.append(name, val)
    }
    return headers
}

function defaultOptions(): FetchOptions | undefined {
    if (typeof Headers === 'undefined') {
        return undefined // for unit tests
    }
    const headers = new Headers()
    if (window.context) {
        for (const [name, value] of Object.entries(window.context.xhrHeaders)) {
            headers.set(name, value)
        }
    }
    return {
        headers
    }
}

export function doFetch(url: string, opt?: any): Promise<Response> {
    const defaults = defaultOptions()
    const fetchOptions = { ...defaults, ...opt }
    if (opt && opt.headers && defaults) {
        // the above object merge might override the auth headers. add those back in.
        fetchOptions.headers = combineHeaders(opt.headers, defaults.headers)
    }
    return fetch(url, fetchOptions)
}
