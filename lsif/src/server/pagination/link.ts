import express from 'express'

/**
 * Create a link header payload with a next link based on the previous endpoint.
 *
 * @param req The HTTP request.
 * @param params The query params to overwrite.
 */
export function nextLink(req: express.Request, params: { [K: string]: any }): string {
    const url = new URL(`${req.protocol}://${req.get('host')}${req.originalUrl}`)
    for (const [key, value] of Object.entries(params)) {
        url.searchParams.set(key, String(value))
    }

    return `<${url.href}>; rel="next"`
}
