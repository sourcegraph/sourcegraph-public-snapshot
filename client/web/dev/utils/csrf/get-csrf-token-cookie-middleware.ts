import { RequestHandler } from 'express'

import { CSRF_COOKIE_NAME } from './get-csrf-token-and-cookie'

// Attach `CSRF_COOKIE_NAME` cookie to every response to avoid "CSRF token is invalid" API error.
export const getCSRFTokenCookieMiddleware = (csrfCookieValue: string): RequestHandler => (request, response, next) => {
    // Based on the PR https://github.com/sourcegraph/sourcegraph/pull/27313
    // we should add a special header x-requested-with to all requests to
    // pass trust logic on BE request resolver.
    request.headers['x-requested-with'] = 'Sourcegraph FE local server'
    response.cookie(CSRF_COOKIE_NAME, csrfCookieValue, { httpOnly: true })
    next()
}
