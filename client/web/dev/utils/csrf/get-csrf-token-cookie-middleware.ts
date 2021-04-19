import { RequestHandler } from 'express'

export const getCSRFTokenCookieMiddleware = (csrfCookieValue: string): RequestHandler => (_request, response, next) => {
    response.cookie('sg_csrf_token', csrfCookieValue, { httpOnly: true })
    next()
}
