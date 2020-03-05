import express from 'express'

/**
 * Create a link header payload with a next link based on the previous endpoint.
 *
 * @param req The HTTP request.
 * @param params The query params to overwrite.
 */
export function nextLink(
    req: express.Request,
    params: { [name: string]: string | number | boolean | undefined }
): string {
    // Requests always have a host header
    // eslint-disable-next-line @typescript-eslint/no-non-null-assertion
    const url = new URL(`${req.protocol}://${req.get('host')!}${req.originalUrl}`)
    for (const [key, value] of Object.entries(params)) {
        if (value !== undefined) {
            url.searchParams.set(key, String(value))
        }
    }

    return `<${url.href}>; rel="next"`
}
