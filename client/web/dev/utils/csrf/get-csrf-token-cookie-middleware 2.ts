import { RequestHandler } from 'express'

import { CSRF_COOKIE_NAME } from './get-csrf-token-and-cookie'

// Attach `CSRF_COOKIE_NAME` cookie to every response to avoid "CSRF token is invalid" API error.
export const getCSRFTokenCookieMiddleware = (csrfCookieValue: string): RequestHandler => (_request, response, next) => {
    response.cookie(CSRF_COOKIE_NAME, csrfCookieValue, { httpOnly: true })
    next()
}
