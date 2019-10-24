/**
 * Checks if a given fetch Response has a HTTP 2xx status code and throws an Error otherwise.
 */
export function checkOk(response: Response): Response {
    if (!response.ok) {
        throw Object.assign(
            new Error(`Request to ${response.url} failed with ${response.status} ${response.statusText}`),
            { response }
        )
    }
    return response
}
