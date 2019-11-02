import { isErrorLike } from '../util/errors'

export const EHTTPSTATUS = 'HTTPStatusError'

export class HTTPStatusError extends Error {
    public readonly name = EHTTPSTATUS
    public readonly code = EHTTPSTATUS
    constructor(public response: Response) {
        super(`Request to ${response.url} failed with ${response.status} ${response.statusText}`)
    }
}
export const isHTTPStatusError = (err: unknown): err is HTTPStatusError => isErrorLike(err) && err.code === EHTTPSTATUS

/**
 * Checks if a given fetch Response has a HTTP 2xx status code and throws an Error otherwise.
 */
export function checkOk(response: Response): Response {
    if (!response.ok) {
        throw new HTTPStatusError(response)
    }
    return response
}
