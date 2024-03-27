export function getHeaders(): { [header: string]: string } {
    const headers: { [header: string]: string } = {
        ...window?.context?.xhrHeaders,
        Accept: 'application/json',
        'Content-Type': 'application/json',
    }
    const parameters = new URLSearchParams(window.location.search)
    const trace = parameters.get('trace')
    if (trace) {
        headers['X-Sourcegraph-Should-Trace'] = trace
    }
    const feat = parameters.getAll('feat')
    if (feat.length) {
        headers['X-Sourcegraph-Override-Feature'] = feat.join(',')
    }
    return headers
}
