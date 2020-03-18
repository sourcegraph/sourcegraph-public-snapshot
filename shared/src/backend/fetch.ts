import { isErrorLike } from '../util/errors'

const EHTTPSTATUS = 'HTTPStatusError'

export class HTTPStatusError extends Error {
    public readonly name = EHTTPSTATUS
    public readonly code = EHTTPSTATUS
    public readonly status: number

    constructor(response: Response) {
        super(`Request to ${response.url} failed with ${response.status} ${response.statusText}`)
        this.status = response.status
    }
}

/**
 * Lenient helper to check if an error was a HTTPStatusError that failed with a specific HTTP status.
 * Works also if the error was sent from a web worker or background page (see https://github.com/mozilla/webextension-polyfill/issues/210).
 *
 * @param error The error
 * @param status The status code to check
 */
export const failedWithHTTPStatus = (error: unknown, status: number): boolean =>
    isErrorLike(error) && error.message.includes(` failed with ${status} `)

/**
 * Checks if a given fetch Response has a HTTP 2xx status code and throws an HTTPStatusError otherwise.
 */
export function checkOk(response: Response): Response {
    if (!response.ok) {
        throw new HTTPStatusError(response)
    }
    return response
}
