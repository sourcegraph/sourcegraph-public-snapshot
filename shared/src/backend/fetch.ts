import { isErrorLike } from '../util/errors'
// eslint-disable-next-line no-restricted-imports
import { AjaxResponse } from 'rxjs/ajax'

export const EHTTPSTATUS = 'HTTPStatusError'

export class HTTPStatusError extends Error {
    public readonly name = EHTTPSTATUS
    public readonly code = EHTTPSTATUS
    constructor(public response: Response | AjaxResponse) {
        super(
            `Request to ${response instanceof Response ? response.url : response.xhr.responseURL} failed with ${
                response.status
            } ${response instanceof Response ? response.statusText : response.xhr.statusText}`
        )
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
