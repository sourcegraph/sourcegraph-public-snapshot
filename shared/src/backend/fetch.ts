import { isErrorLike } from '../util/errors'

const EAJAX = 'AjaxError'

interface AjaxError extends Error {
    name: typeof EAJAX
    response: Response
}

export const isAjaxError = (err: unknown): err is AjaxError => isErrorLike(err) && err.name === EAJAX

/**
 * Checks if a given fetch Response has a HTTP 2xx status code and throws an Error otherwise.
 */
export function checkOk(response: Response): Response {
    if (!response.ok) {
        throw Object.assign(
            new Error(`Request to ${response.url} failed with ${response.status} ${response.statusText}`),
            { name: EAJAX, response }
        )
    }
    return response
}
